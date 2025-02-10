package parser

import (
	"bodybyrocket/internal/database"
	"bodybyrocket/internal/lib"
	"github.com/SevereCloud/vksdk/v3/api"
	"gorm.io/gorm"
	"regexp"
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

		for _, attachment := range wallpost.Attachments {
			if attachment.Type != "video" {
				break
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
	if !regexp.MustCompile(`длинную тренировку|короткую тренировку|зарядку|новые тренировки`).MatchString(text) {
		return ""
	}

	// интересует только текст до фразы "Ракеты, напоминаю", т.к. дальше однотипный рекламный текст
	re := regexp.MustCompile(`\(\w+\s+весь\s+пост\..*важная\s+памятка.*\)`)
	text = re.ReplaceAllString(text, "")

	// Регулярка для удаления всех акцентов
	reAccent := regexp.MustCompile(`Акценты:\s*[\s\S]+?`)
	text = reAccent.ReplaceAllString(text, "")

	return text
}
