package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"path/filepath"
)

type Vk struct {
	Token NotEmptyString `env:"VK_TOKEN" env-required:"true"`
}

type Telegram struct {
	ApiId     string `env:"API_ID"`
	ApiHash   string `env:"API_HASH"`
	BotToken  string `env:"BOT_TOKEN"`
	ChannelId int64  `env:"CHANNEL_ID"`
}

type Database struct {
	User   string `env:"DB_USERNAME" env-required:"true"`
	Pass   string `env:"DB_PASSWORD" env-required:"true"`
	Dbname string `env:"DB_DATABASE" env-required:"true"`
	Port   string `env:"DB_PORT" env-required:"true"`
	Host   string `env:"DB_HOST" env-required:"true"`
}

type DataFolder struct {
	Base   string `env:"DATA_PATH" env-default:".data"`
	Tdlib  string `env:"DATA_TDLIB_PATH"`
	Videos string `env:"DATA_VIDEOS_PATH"`
}

type Config struct {
	Vk         Vk
	Telegram   Telegram
	Database   Database
	DataFolder DataFolder
}

func NewConfig(path string) (*Config, error) {
	var cfg Config

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return nil, err
	}

	if cfg.DataFolder.Tdlib == "" {
		cfg.DataFolder.Tdlib = filepath.Join(cfg.DataFolder.Base, "tdlib")
	}

	if cfg.DataFolder.Videos == "" {
		cfg.DataFolder.Videos = filepath.Join(cfg.DataFolder.Base, "videos")
	}

	return &cfg, nil
}
