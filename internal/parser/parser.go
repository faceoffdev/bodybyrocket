package parser

import (
	"bodybyrocket/internal/database"
	"github.com/SevereCloud/vksdk/v3/api"
	"github.com/SevereCloud/vksdk/v3/object"
	"gorm.io/gorm"
	"regexp"
	"strings"
	"time"
)

type Parser struct {
	vk *api.VK
	db *gorm.DB
}

func New(vk *api.VK, db *gorm.DB) *Parser {
	return &Parser{vk, db}
}

func (p *Parser) GetAllPosts(groupId int, yield func(post object.WallWallpost) bool) {
	const maxCountItems = 10

	var (
		offset int
		wall   api.WallGetResponse
		err    error
	)

	for ok := true; ok; ok = offset < wall.Count {
		wall, err = p.vk.WallGet(api.Params{"owner_id": groupId, "count": maxCountItems, "offset": offset})
		if err != nil {
			return
		}

		for _, post := range wall.Items {
			if !yield(post) {
				return
			}
		}

		offset += maxCountItems
	}
}

func (p *Parser) ExportFromGroup(groupId int, isFree bool) error {
	var posts []database.Post

	last := &database.Post{}
	p.db.Where("group_id = ?", groupId).Order("post_id DESC").First(last)

	p.GetAllPosts(groupId, func(wallpost object.WallWallpost) bool {
		if last.PostID == wallpost.ID {
			return false
		}

		if wallpost.IsPinned {
			return true
		}

		if len(wallpost.Attachments) != 1 {
			return true
		}

		attachment := wallpost.Attachments[0]
		if attachment.Type != "video" {
			return true
		}

		if attachment.Video.Duration <= 300 {
			return true
		}

		text := prepareText(wallpost.Text, isFree)
		if text == "" {
			return true
		}

		posts = append(posts, database.Post{
			PostID:      wallpost.ID,
			PublishedAt: time.Unix(int64(wallpost.Date), 0),
			GroupID:     attachment.Video.OwnerID,
			VideoID:     attachment.Video.ID,
			Text:        text,
		})

		return true
	})

	if posts == nil {
		return nil
	}

	return p.db.CreateInBatches(posts, 100).Error
}

func prepareText(text string, isFree bool) string {
	if isFree {
		return "Бесплатная тренировка 🏷️"
	}

	// если не находим ключевые фразы, то выходим
	if !regexp.MustCompile(`длинную тренировку|короткую тренировку|зарядку`).MatchString(text) {
		return ""
	}

	// интересует только текст до фразы "Ракеты, напоминаю", т.к. дальше однотипный рекламный текст
	match := strings.Split(text, "\n\n")
	if len(match) > 0 {
		return match[0]
	}

	return ""
}
