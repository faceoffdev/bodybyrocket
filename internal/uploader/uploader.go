package uploader

import (
	"bodybyrocket/internal/database"
	"bodybyrocket/internal/lib"
	"bodybyrocket/internal/tdlib"
	"errors"
	"fmt"
	"github.com/SevereCloud/vksdk/v3/api"
	"github.com/SevereCloud/vksdk/v3/object"
	"gorm.io/gorm"
	"sync"
	"time"
)

const (
	maxWorkers    = 2
	uploadTimeout = 30 * time.Minute
)

const (
	maxWidth  = 1280
	maxHeight = 720
)

type Uploader struct {
	vk         *api.VK
	tg         *tdlib.Telegram
	db         *gorm.DB
	dataFolder string
}

func NewUploader(vk *api.VK, tg *tdlib.Telegram, db *gorm.DB, dataFolder string) *Uploader {
	return &Uploader{vk, tg, db, dataFolder}
}

func (u *Uploader) Upload(chatId int64) {
	posts := u.getPosts()

	if len(posts) == 0 {
		return
	}

	var wg sync.WaitGroup
	wg.Add(len(posts))

	sem := lib.NewSemaphore(maxWorkers)

	for _, post := range posts {
		postID := post.ID

		sem.Acquire()
		go func() {
			defer func() {
				wg.Done()
				sem.Release()
			}()

			fmt.Printf("скачиваю видео %d из VK...\n", postID)

			if file, err := u.download(post); err == nil {
				fmt.Printf("загружаю видео %d в Telegram...\n", postID)

				if err = u.tg.SendVideo(chatId, file, uploadTimeout); err != nil {
					fmt.Printf("ошибка загрузки видео %d в Telegram: %v\n", postID, err)
				} else {
					fmt.Printf("видео %d успешно загружено в Telegram\n", postID)

					u.db.Model(&database.Post{}).Where("id = ?", postID).Update("processed", true)
				}

				lib.RemoveFiles(file.Path, file.PreviewPath)
			} else {
				fmt.Printf("ошибка скачивания видео %d из VK: %v\n", postID, err)
			}
		}()
	}

	wg.Wait()
}

func (u *Uploader) getPosts() []database.Post {
	var posts []database.Post

	u.db.Model(&database.Post{}).
		Select("id", "group_id", "video_id", "text").
		Where("processed = ?", false).
		Order("published_at").
		Find(&posts)

	return posts
}

func (u *Uploader) download(post database.Post) (*tdlib.VideoLocalFile, error) {
	v, err := u.vk.VideoGet(api.Params{"videos": fmt.Sprintf("%d_%d", post.GroupID, post.VideoID)})
	if err != nil {
		return nil, err
	}

	if v.Count == 0 {
		return nil, errors.New("video not found")
	}

	video := v.Items[0]

	filePath := fmt.Sprintf("%s/%d.mp4", u.dataFolder, post.ID)
	if err = downloadVideo(filePath, video.Files); err != nil {
		return nil, err
	}

	previewPath := fmt.Sprintf("%s/%d.jpg", u.dataFolder, post.ID)
	if err = downloadImage(previewPath, video.Image); err != nil {
		return nil, err
	}

	return &tdlib.VideoLocalFile{
		Caption:     post.Text,
		Path:        filePath,
		PreviewPath: previewPath,
		Width:       maxWidth,
		Height:      maxHeight,
	}, nil
}

func downloadVideo(path string, file object.VideoVideoFiles) (err error) {
	for _, url := range [...]string{file.Mp4_720, file.Mp4_480, file.Mp4_1080} {
		if url == "" {
			err = errors.New("video url not found")

			continue
		}

		if err = lib.DownloadFile(url, path); err == nil {
			break
		}
	}

	return
}

func downloadImage(path string, images []object.VideoVideoImage) error {
	if len(images) == 0 {
		return errors.New("no images found")
	}

	url := images[0].URL
	for i := len(images) - 1; i > 0; i-- {
		if img := images[i]; img.Width <= maxWidth && img.Height <= maxHeight {
			url = img.URL
			break
		}
	}

	return lib.DownloadFile(url, path)
}
