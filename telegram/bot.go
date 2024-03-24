package telegram

import (
	"context"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"path"
)

func (tl *Listener) StartBotClient(ctx context.Context) error {
	botUpdates := tg.NewUpdateDispatcher()
	tl.BotClient = telegram.NewClient(tl.Config.AppId, tl.Config.AppHash, telegram.Options{
		SessionStorage: &session.FileStorage{
			Path: path.Join(tl.Config.SessionFolder, ".tg-bot-session.json"),
		},
		UpdateHandler: botUpdates,
		Logger:        tl.Logger,
	})

	return tl.BotClient.Run(ctx, func(ctx context.Context) error {
		status, err := tl.BotClient.Auth().Status(ctx)
		if err != nil {
			return err
		}

		if !status.Authorized {
			if _, err := tl.BotClient.Auth().Bot(ctx, tl.Config.BotToken); err != nil {
				return err
			}
		}

		botUpdates.OnNewMessage(tl.handleNewBotMessage)
		tl.Logger.Info("Bot started successfully.")
		return telegram.RunUntilCanceled(ctx, tl.BotClient)
	})
}
