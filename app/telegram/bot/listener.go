package bot

import (
	"context"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
)

type MessageCtx struct {
	u        *tg.UpdateNewMessage
	msg      *tg.Message
	entities tg.Entities
}

func (c *Client) Listen(ctx context.Context) error {
	return c.client.Run(ctx, func(ctx context.Context) error {
		status, err := c.auth.Status(ctx)
		if err != nil {
			return err
		}

		if !status.Authorized {
			if _, err := c.client.Auth().Bot(ctx, c.cfg.BotToken); err != nil {
				return err
			}
		}

		c.updates.OnNewMessage(c.handleNewMessage)
		c.logger.Info("Bot started successfully.")
		return telegram.RunUntilCanceled(ctx, c.client)
	})
}

func (c *Client) handleNewMessage(ctx context.Context, entities tg.Entities, u *tg.UpdateNewMessage) error {
	msg, ok := u.Message.(*tg.Message)
	if !ok {
		return nil
	}
	c.logger.Info("Bot Message", zap.Any("message", msg))

	msgContext := &MessageCtx{u, msg, entities}

	if msg.Message == "/start" {
		return c.handleStartCommand(msgContext)
	}

	if msg.Message == "/reset_cache" {
		return c.handleResetCacheCommand(msgContext)
	}

	if msg.Message == "/ping" {
		return c.handlePingCommand(msgContext)
	}

	return c.processPhoneNum(msgContext)
}
