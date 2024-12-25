package parser

import (
	"bodybyrocket/internal/database"
	"bodybyrocket/internal/lib"
	"github.com/SevereCloud/vksdk/v3/api"
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

func (p *Parser) ExportFromGroup(groupId int, isFree bool) error {
	var posts []database.Post

	last := &database.Post{}
	p.db.Where("group_id = ?", groupId).Order("post_id DESC").First(last)

	for wallpost, err := range lib.IterateWallPosts(p.vk, groupId) {
		if err != nil {
			return err
		}

		if last.PostID == wallpost.ID {
			break
		}

		if wallpost.IsPinned {
			continue
		}

		if len(wallpost.Attachments) != 1 {
			continue
		}

		attachment := wallpost.Attachments[0]
		if attachment.Type != "video" {
			continue
		}

		if attachment.Video.Duration <= 300 {
			continue
		}

		text := prepareText(wallpost.Text, isFree)
		if text == "" {
			continue
		}

		posts = append(posts, database.Post{
			PostID:      wallpost.ID,
			PublishedAt: time.Unix(int64(wallpost.Date), 0),
			GroupID:     attachment.Video.OwnerID,
			VideoID:     attachment.Video.ID,
			Text:        text,
		})
	}

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
