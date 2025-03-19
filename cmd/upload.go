package cmd

import (
	"bodybyrocket/internal/config"
	"bodybyrocket/internal/database"
	"bodybyrocket/internal/lib"
	"bodybyrocket/internal/tdlib"
	"bodybyrocket/internal/uploader"
	"fmt"
	"github.com/SevereCloud/vksdk/v3/api"
	"github.com/spf13/cobra"
	"os"
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "upload",
		Short: "Загрузить видео в Telegram",
		RunE:  UploadHandler,
	})
}

func UploadHandler(_ *cobra.Command, _ []string) error {
	cfg, err := config.NewConfig(".env")
	if err != nil {
		return fmt.Errorf("ошибка создания конфигурации: %v", err)
	}

	for _, f := range [...]string{cfg.DataFolder.Base, cfg.DataFolder.Tdlib, cfg.DataFolder.Videos} {
		if ok, err := lib.IsDirectory(f); !ok || err != nil {
			os.Mkdir(f, os.ModePerm)
		}
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}
	defer database.Close(db)

	tg, err := tdlib.NewTelegram(cfg.Telegram, cfg.DataFolder.Tdlib)
	if err != nil {
		return fmt.Errorf("ошибка подключения к Telegram: %v", err)
	}
	defer tg.Shutdown()

	u := uploader.NewUploader(api.NewVK(string(cfg.Vk.Token)), tg, db, cfg.DataFolder.Videos)
	u.Upload(cfg.Telegram.ChannelId)

	return nil
}
