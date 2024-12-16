package database

import (
	"bodybyrocket/internal/config"
	"fmt"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func Connect(c config.Database) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable application_name=bodybyrocket TimeZone=Europe/Moscow",
		c.Host, c.User, c.Pass, c.Dbname, c.Port)

	db, err := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true, // disables implicit prepared statement usage
	}), &gorm.Config{})

	if err != nil {
		return nil, err
	}

	if err = db.AutoMigrate(&Post{}); err != nil {
		return nil, err
	}

	var version string
	if err = db.Raw("select version();").Row().Scan(&version); err != nil {
		return nil, err
	}

	log.Println(version)

	return db, nil
}

func Close(db *gorm.DB) {
	if sqlDB, err := db.DB(); err == nil {
		_ = sqlDB.Close()
	}
}
