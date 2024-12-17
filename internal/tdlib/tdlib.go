package tdlib

import (
	"bodybyrocket/internal/config"
	"fmt"
	"github.com/zelenin/go-tdlib/client"
	"log"
	"path/filepath"
	"strconv"
)

const (
	SystemVersion      = "1.0.0"
	ApplicationVersion = "1.0.0"
)

type Telegram struct {
	client   *client.Client
	listener *client.Listener
}

func New(cfg config.Telegram) (*Telegram, error) {
	tdlibClient, err := newTelegramClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create tdlib client: %w", err)
	}

	u := &Telegram{
		client:   tdlibClient,
		listener: NewListener(tdlibClient),
	}

	return u, nil
}

func newTelegramClient(cfg config.Telegram) (*client.Client, error) {
	apiId64, err := strconv.ParseInt(cfg.ApiId, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ApiId from config: %w", err)
	}

	authorizer := client.BotAuthorizer(cfg.BotToken)

	authorizer.TdlibParameters <- &client.SetTdlibParametersRequest{
		DatabaseDirectory:  filepath.Join(".data", "tdlib", "database"),
		ApiId:              int32(apiId64),
		ApiHash:            cfg.ApiHash,
		SystemLanguageCode: "ru",
		DeviceModel:        "Telegram",
		SystemVersion:      SystemVersion,
		ApplicationVersion: ApplicationVersion,
	}

	if _, err = client.SetLogVerbosityLevel(&client.SetLogVerbosityLevelRequest{NewVerbosityLevel: 0}); err != nil {
		return nil, fmt.Errorf("failed to set tdlib log verbosity level: %w", err)
	}

	optionValue, err := client.GetOption(&client.GetOptionRequest{
		Name: "version",
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get TDLib version: %w", err)
	}
	if v, ok := optionValue.(*client.OptionValueString); ok {
		log.Printf("TDLib version: %s", v.Value)
	} else {
		log.Printf("TDLib version could not be retrieved properly")
	}

	newClient, err := client.NewClient(authorizer)
	if err != nil {
		return nil, fmt.Errorf("failed to create new tdlib client: %w", err)
	}

	return newClient, nil
}

func (u *Telegram) Shutdown() {
	u.listener.Close()
	u.client.Stop()
}

func (u *Telegram) SendVideo(chatId int64, file *VideoLocalFile) (*client.Message, error) {
	if file.Path == "" {
		return nil, fmt.Errorf("filePath is empty")
	}
	if chatId == 0 {
		return nil, fmt.Errorf("chatId is empty")
	}

	chat, err := u.client.GetChat(&client.GetChatRequest{
		ChatId: chatId,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to get chat with id %d: %w", chatId, err)
	}

	var formattedText *client.FormattedText
	if file.Caption != "" {
		formattedText, err = client.ParseTextEntities(&client.ParseTextEntitiesRequest{
			Text:      file.Caption,
			ParseMode: &client.TextParseModeHTML{},
		})
		if err != nil {
			return nil, fmt.Errorf("failed to parse text entities: %w", err)
		}
	}
	var thumbnail *client.InputThumbnail
	if file.PreviewPath != "" {
		thumbnail = &client.InputThumbnail{
			Thumbnail: &client.InputFileLocal{Path: file.PreviewPath},
		}
	}

	msg, err := u.client.SendMessage(&client.SendMessageRequest{
		ChatId: chat.Id,
		InputMessageContent: &client.InputMessageVideo{
			Video:             &client.InputFileLocal{Path: file.Path},
			Thumbnail:         thumbnail,
			Width:             file.Width,
			Height:            file.Height,
			Caption:           formattedText,
			SupportsStreaming: true,
		},
	})
	if err != nil {
		return nil, fmt.Errorf("failed to send video message: %w", err)
	}

	return msg, nil
}
