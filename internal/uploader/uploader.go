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
	"os"
	"sync"
	"time"
)

const (
	DataVideoFolder = ".data/videos"
	maxWorkers      = 2
	uploadTimeout   = 30 * time.Minute
)

const (
	maxWidth  = 1280
	maxHeight = 720
)

type Uploader struct {
	vk *api.VK
	tg *tdlib.Telegram
	db *gorm.DB
}

func New(vk *api.VK, tg *tdlib.Telegram, db *gorm.DB) *Uploader {
	return &Uploader{vk, tg, db}
}

func (t *Uploader) getPosts() []database.Post {
	var posts []database.Post

	t.db.Model(&database.Post{}).
		Select("id", "group_id", "video_id", "text").
		Where("processed = ?", false).
		Order("published_at").
		Find(&posts)

	return posts
}

func (t *Uploader) Upload(chatId int64) {
	posts := t.getPosts()

	if len(posts) == 0 {
		return
	}

	sem := make(chan struct{}, maxWorkers)

	var wg sync.WaitGroup
	wg.Add(len(posts))

	for _, post := range posts {
		sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer t.db.Model(&database.Post{}).Where("id = ?", post.ID).Update("processed", true)

			fmt.Printf("скачиваю видео %d из VK...\n", post.ID)

			if file, err := t.download(post); err == nil {
				fmt.Printf("загружаю видео %d в Telegram...\n", post.ID)

				if _, err = t.tg.SendVideo(chatId, file, uploadTimeout); err != nil {
					fmt.Printf("ошибка загрузки видео %d в Telegram: %v\n", post.ID, err)
				}

				go func() {
					_ = os.Remove(file.Path)

					if file.PreviewPath != "" {
						_ = os.Remove(file.PreviewPath)
					}
				}()
			} else {
				fmt.Printf("ошибка скачивания видео %d из VK: %v\n", post.ID, err)
			}

			<-sem
		}()
	}

	wg.Wait()
}

func (t *Uploader) download(post database.Post) (*tdlib.VideoLocalFile, error) {
	v, err := t.vk.VideoGet(api.Params{"videos": fmt.Sprintf("%d_%d", post.GroupID, post.VideoID)})
	if err != nil {
		return nil, err
	}

	if v.Count == 0 {
		return nil, errors.New("video not found")
	}

	video := v.Items[0]

	filePath, err := downloadVideo(post.ID, video.Files)
	if err != nil {
		return nil, err
	}

	return &tdlib.VideoLocalFile{
		Caption:     post.Text,
		Path:        filePath,
		PreviewPath: downloadImage(post.ID, video.Image),
		Width:       maxWidth,
		Height:      maxHeight,
	}, nil
}

func downloadVideo(postId int, file object.VideoVideoFiles) (string, error) {
	var err error

	path := fmt.Sprintf("%s/%d.mp4", DataVideoFolder, postId)
	for _, url := range [...]string{file.Mp4_720, file.Mp4_480} {
		if url == "" {
			err = errors.New("video url not found")

			continue
		}

		if err = lib.DownloadFile(url, path); err == nil {
			break
		}
	}

	if err != nil {
		return "", err
	}

	return path, nil
}

func downloadImage(postId int, images []object.VideoVideoImage) string {
	if len(images) == 0 {
		return ""
	}

	path := fmt.Sprintf("%s/%d.jpg", DataVideoFolder, postId)
	url := images[0].URL
	for i := len(images) - 1; i > 0; i-- {
		if img := images[i]; img.Width <= maxWidth && img.Height <= maxHeight {
			url = img.URL
			break
		}
	}

	if err := lib.DownloadFile(url, path); err != nil {
		return ""
	}

	return path
}
