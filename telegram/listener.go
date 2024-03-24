package telegram

import (
	"github.com/gotd/td/telegram"
	"github.com/pkarpovich/tg-contacts-search/config"
	"go.uber.org/zap"
)

type Listener struct {
	BotClient *telegram.Client
	Logger    *zap.Logger
	Config    config.Config
}

func NewListener(logger *zap.Logger, config config.Config) *Listener {
	return &Listener{
		Logger: logger,
		Config: config,
	}
}
