package uploader

import (
	"bodybyrocket/internal/database"
	"bodybyrocket/internal/lib"
	"bodybyrocket/internal/tdlib"
	"errors"
	"fmt"
	"github.com/SevereCloud/vksdk/v3/api"
	"gorm.io/gorm"
	"sync"
)

const (
	DataVideoFolder = ".data/videos"
	maxWorkers      = 2
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

	var wg sync.WaitGroup
	sem := make(chan struct{}, maxWorkers)

	for _, post := range posts {
		wg.Add(1)
		sem <- struct{}{}

		go func() {
			defer wg.Done()
			defer t.db.Model(&database.Post{}).Where("id = ?", post.ID).Update("processed", true)

			fmt.Printf("скачиваю видео %d...\n", post.ID)

			file, err := t.download(post)
			if err == nil {
				file.Caption = post.Text

				fmt.Printf("загружаю видео %d в Telegram...\n", post.ID)
				_, err = t.tg.SendVideo(chatId, file)
			}

			if err != nil {
				fmt.Printf("ошибка загрузки видео %d: %v\n", post.ID, err)
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
	videoURL := video.Files.Mp4_720
	if videoURL == "" {
		videoURL = video.Files.Mp4_480
		if videoURL == "" {
			return nil, errors.New("video url not found")
		}
	}

	filePath := fmt.Sprintf("%s/%d.mp4", DataVideoFolder, post.ID)
	if err = lib.DownloadFile(videoURL, filePath); err != nil {
		return nil, err
	}

	previewPath := fmt.Sprintf("%s/%d.jpg", DataVideoFolder, post.ID)
	previewURL := video.Image[len(video.Image)-1].URL
	if err = lib.DownloadFile(previewURL, previewPath); err != nil {
		previewPath = ""
	}

	return &tdlib.VideoLocalFile{
		Path:        filePath,
		PreviewPath: previewPath,
		Width:       1280,
		Height:      720,
	}, nil
}
