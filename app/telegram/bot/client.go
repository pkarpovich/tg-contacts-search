package bot

import (
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/pkarpovich/tg-contacts-search/app/config"
	"go.uber.org/zap"
	"path"
)

type Client struct {
	client  *telegram.Client
	auth    *auth.Client
	updates tg.UpdateDispatcher
	logger  *zap.Logger
	cfg     config.TelegramConfig
}

func NewClient(logger *zap.Logger, cfg config.TelegramConfig) *Client {
	updates := tg.NewUpdateDispatcher()
	client := telegram.NewClient(cfg.AppId, cfg.AppHash, telegram.Options{
		SessionStorage: &session.FileStorage{
			Path: path.Join(cfg.SessionFolder, ".tg-bot-session.json"),
		},
		UpdateHandler: updates,
		Logger:        logger,
	})

	return &Client{client, client.Auth(), updates, logger, cfg}
}
