package tdlib

import (
	"bodybyrocket/internal/config"
	"errors"
	"fmt"
	"github.com/zelenin/go-tdlib/client"
	"log"
	"path/filepath"
	"strconv"
	"time"
)

const (
	SystemVersion      = "1.0.0"
	ApplicationVersion = "1.0.0"
)

type Telegram struct {
	client *client.Client
}

func NewTelegram(cfg config.Telegram, dataFolder string) (*Telegram, error) {
	tdlibClient, err := newClient(cfg, dataFolder)
	if err != nil {
		return nil, fmt.Errorf("failed to create tdlib client: %w", err)
	}

	t := &Telegram{
		client: tdlibClient,
	}

	return t, nil
}

func newClient(cfg config.Telegram, dataFolder string) (*client.Client, error) {
	apiId64, err := strconv.ParseInt(cfg.ApiId, 10, 32)
	if err != nil {
		return nil, fmt.Errorf("failed to parse ApiId from config: %w", err)
	}

	authorizer := client.BotAuthorizer(cfg.BotToken)

	authorizer.TdlibParameters <- &client.SetTdlibParametersRequest{
		DatabaseDirectory:  filepath.Join(dataFolder, "database"),
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

	c, err := client.NewClient(authorizer)
	if err != nil {
		return nil, fmt.Errorf("failed to create new tdlib client: %w", err)
	}

	return c, nil
}

func (t *Telegram) Shutdown() {
	t.client.Stop()
}

func (t *Telegram) SendVideo(chatId int64, file *VideoLocalFile, uploadTimeout time.Duration) error {
	if file.Path == "" {
		return fmt.Errorf("filePath is empty")
	}
	if chatId == 0 {
		return fmt.Errorf("chatId is empty")
	}

	chat, err := t.client.GetChat(&client.GetChatRequest{
		ChatId: chatId,
	})
	if err != nil {
		return fmt.Errorf("failed to get chat with id %d: %w", chatId, err)
	}

	var formattedText *client.FormattedText
	if file.Caption != "" {
		formattedText, err = client.ParseTextEntities(&client.ParseTextEntitiesRequest{
			Text:      file.Caption,
			ParseMode: &client.TextParseModeHTML{},
		})
		if err != nil {
			return fmt.Errorf("failed to parse text entities: %w", err)
		}
	}
	var thumbnail *client.InputThumbnail
	if file.PreviewPath != "" {
		thumbnail = &client.InputThumbnail{
			Thumbnail: &client.InputFileLocal{Path: file.PreviewPath},
		}
	}

	err = t.SendMessageWithTimeout(&client.SendMessageRequest{
		ChatId: chat.Id,
		InputMessageContent: &client.InputMessageVideo{
			Video:             &client.InputFileLocal{Path: file.Path},
			Thumbnail:         thumbnail,
			Width:             file.Width,
			Height:            file.Height,
			Caption:           formattedText,
			SupportsStreaming: true,
		},
	}, uploadTimeout)
	if err != nil {
		return fmt.Errorf("failed to send video message: %w", err)
	}

	return nil
}

func (t *Telegram) SendMessageWithTimeout(req *client.SendMessageRequest, d time.Duration) error {
	listener := t.client.GetListener()
	defer listener.Close()

	msg, err := t.client.SendMessage(req)
	if err != nil {
		return err
	}

	timeout := time.After(d)
	for {
		select {
		case <-timeout:
			return errors.New("timeout")
		case update := <-listener.Updates:
			switch upd := update.(type) {
			case *client.UpdateMessageSendSucceeded:
				if upd.OldMessageId == msg.Id {
					return nil
				}
			}
		}
	}
}
