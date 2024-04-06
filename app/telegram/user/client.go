package user

import (
	"github.com/gotd/contrib/bg"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"github.com/pkarpovich/tg-contacts-search/app/config"
	"go.uber.org/zap"
	"path/filepath"
	"time"
)

type Client struct {
	api    *tg.Client
	auth   *auth.Client
	cfg    config.TelegramConfig
	logger *zap.Logger
}

func NewClient(logger *zap.Logger, cfg config.TelegramConfig) (*Client, bg.StopFunc, error) {
	dispatcher := tg.NewUpdateDispatcher()
	manager := updates.New(updates.Config{
		Handler: dispatcher,
	})

	client := telegram.NewClient(cfg.AppId, cfg.AppHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: filepath.Join(cfg.SessionFolder, ".tg-user-session.json"),
		},
		DialTimeout:   time.Minute * 5,
		Device:        deviceConfig(),
		Logger:        logger,
		UpdateHandler: manager,
		DC:            5,
	})

	stop, err := bg.Connect(client)
	if err != nil {
		return nil, nil, err
	}

	client.Auth()

	return &Client{client.API(), client.Auth(), cfg, logger}, stop, nil
}

func deviceConfig() telegram.DeviceConfig {
	return telegram.DeviceConfig{
		DeviceModel:    "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/116.0",
		SystemVersion:  "Win32",
		AppVersion:     "2.1.9 K",
		LangPack:       "webk",
		SystemLangCode: "en",
		LangCode:       "en",
	}
}
