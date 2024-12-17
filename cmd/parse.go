package cmd

import (
	"bodybyrocket/internal/config"
	"bodybyrocket/internal/database"
	"bodybyrocket/internal/parser"
	"fmt"
	"github.com/SevereCloud/vksdk/v3/api"
	"github.com/spf13/cobra"
	"sync"
)

const (
	FreeGroupID = -211102099
	PaidGroupID = -224167761
)

func init() {
	rootCmd.AddCommand(&cobra.Command{
		Use:   "parse",
		Short: "Экспорт тренировок",
		RunE:  ParseHandler,
	})
}

func ParseHandler(_ *cobra.Command, _ []string) error {
	cfg, err := config.New(".env")
	if err != nil {
		return fmt.Errorf("ошибка создания конфигурации: %v", err)
	}

	db, err := database.Connect(cfg.Database)
	if err != nil {
		return fmt.Errorf("ошибка подключения к базе данных: %v", err)
	}
	defer database.Close(db)

	p := parser.New(api.NewVK(string(cfg.Vk.Token)), db)

	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()

		if err = p.ExportFromGroup(FreeGroupID, true); err != nil {
			fmt.Printf("ошибка экспорта из бесплатной группы: %v", err)
		}
	}()

	go func() {
		defer wg.Done()

		if err = p.ExportFromGroup(PaidGroupID, false); err != nil {
			fmt.Printf("ошибка экспорта из платной группы: %v", err)
		}
	}()

	wg.Wait()

	return nil
}
