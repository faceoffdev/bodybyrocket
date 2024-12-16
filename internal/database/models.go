package database

import "time"

type Post struct {
	ID          int `gorm:"primaryKey"`
	GroupID     int `gorm:"index"`
	PostID      int
	VideoID     int
	PublishedAt time.Time `gorm:"type:timestamp"`
	Text        string    `gorm:"type:varchar(4096);nullable"`
	Processed   bool      `gorm:"default:false"`
}
