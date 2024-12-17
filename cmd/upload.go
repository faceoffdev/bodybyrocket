package cmd

import (
	"bodybyrocket/internal/config"
	"bodybyrocket/internal/database"
	"bodybyrocket/internal/tdlib"
	"bodybyrocket/internal/uploader"
	"fmt"
	"github.com/SevereCloud/vksdk/v3/api"
	"github.com/spf13/cobra"
)

const chatId = -1002298937261

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "upload",
		Short: "Загрузить видео в Telegram",
		RunE:  UploadHandler,
	})
}

func UploadHandler(_ *cobra.Command, _ []string) error {
	cfg, err := config.New(".env")
	if err != nil {
		return fmt.Errorf("ошибка создания конфигурации: %v", err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}
	defer database.Close(db)

	tg, err := tdlib.New(cfg.Telegram)
	if err != nil {
		return fmt.Errorf("ошибка подключения к Telegram: %v", err)
	}
	defer tg.Shutdown()

	u := uploader.New(api.NewVK(string(cfg.Vk.Token)), tg, db)
	u.Upload(chatId)

	return nil
}
