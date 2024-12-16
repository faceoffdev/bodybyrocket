package config

import (
	"github.com/ilyakaznacheev/cleanenv"
)

type Vk struct {
	Token NotEmptyString `env:"VK_TOKEN" env-required:"true"`
}

type Database struct {
	User   string `env:"DB_USERNAME" env-required:"true"`
	Pass   string `env:"DB_PASSWORD" env-required:"true"`
	Dbname string `env:"DB_DATABASE" env-required:"true"`
	Port   string `env:"DB_PORT" env-required:"true"`
	Host   string `env:"DB_HOST" env-required:"true"`
}

type Config struct {
	Vk       Vk
	Database Database
}

func New(path string) (*Config, error) {
	var cfg Config

	err := cleanenv.ReadConfig(path, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
