package bot

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/pkarpovich/tg-contacts-search/app/telegram/user"
	"github.com/pkarpovich/tg-contacts-search/app/utils"
	"go.uber.org/zap"
	"sync"
)

func (c *Client) handleStartCommand(msgContext *MessageCtx) error {
	entities := msgContext.entities
	u := msgContext.u

	sender := message.NewSender(tg.NewClient(c.client))
	ctx := context.Background()

	_, err := sender.Reply(entities, u).Text(ctx, "Welcome to the contacts search bot! Send me a phone number to get the username of the user.")
	if err != nil {
		c.logger.Error("failed to send message", zap.Error(err))
	}

	return nil
}

func (c *Client) handlePingCommand(msgContext *MessageCtx) error {
	entities := msgContext.entities
	ctx := context.Background()
	u := msgContext.u

	sender := message.NewSender(tg.NewClient(c.client))

	userClient, stop, err := user.NewClient(c.logger, c.cfg)
	if err != nil {
		_, err := sender.Reply(entities, u).Text(ctx, fmt.Sprintf("Failed to create user client: %s", err))
		if err != nil {
			c.logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}
	defer func() { _ = stop() }()

	_, err = userClient.GetSelfUsername()
	if err != nil {
		_, err := sender.Reply(entities, u).Text(ctx, fmt.Sprintf("Failed to access user API: %s", err))
		if err != nil {
			c.logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}

	_, err = sender.Reply(entities, u).Text(ctx, "pong")
	if err != nil {
		c.logger.Error("failed to send message", zap.Error(err))
	}

	return nil
}

func (c *Client) handleResetCacheCommand(msgContext *MessageCtx) error {
	entities := msgContext.entities
	ctx := context.Background()
	u := msgContext.u

	sender := message.NewSender(tg.NewClient(c.client))

	c.usernameCache = sync.Map{}

	_, err := sender.Reply(entities, u).Text(ctx, "Cache has been reset")
	if err != nil {
		c.logger.Error("failed to send message", zap.Error(err))
	}

	return nil
}

func (c *Client) processPhoneNum(msgContext *MessageCtx) error {
	entities := msgContext.entities
	msg := msgContext.msg.Message
	u := msgContext.u

	sender := message.NewSender(tg.NewClient(c.client))
	ctx := context.Background()

	if !utils.ValidatePhoneNum(msg) {
		c.logger.Info("Invalid phone number received", zap.String("phoneNum", msg))

		_, err := sender.Reply(entities, u).Text(ctx, "Invalid phone number")
		if err != nil {
			c.logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}

	if username, found := c.usernameCache.Load(msg); found {
		c.logger.Info("Username found in cache", zap.String("username", username.(string)), zap.String("phone", msg))

		_, err := sender.Reply(entities, u).Text(ctx, c.formatUsernameUrl(username.(string)))
		if err != nil {
			c.logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}

	userClient, stop, err := user.NewClient(c.logger, c.cfg)
	if err != nil {
		_, err := sender.Reply(entities, u).Text(ctx, fmt.Sprintf("Failed to create user client: %s", err))
		if err != nil {
			c.logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}
	defer func() { _ = stop() }()

	username, err := userClient.GetUsername(msg)
	if err != nil {
		_, err := sender.Reply(entities, u).Text(ctx, fmt.Sprintf("Failed to get username: %s", err))
		if err != nil {
			c.logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}

	if username == "" {
		c.logger.Info("Username not found", zap.String("phone", msg))

		_, err := sender.Reply(entities, u).Text(ctx, "User not found")
		if err != nil {
			c.logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}

	c.usernameCache.Store(msg, username)
	c.logger.Info("Username found", zap.String("username", username), zap.String("phone", msg))

	_, err = sender.Reply(entities, u).Text(ctx, c.formatUsernameUrl(username))
	if err != nil {
		c.logger.Error("failed to send message", zap.Error(err))
		return nil
	}

	return nil
}

func (c *Client) formatUsernameUrl(username string) string {
	return fmt.Sprintf("https://t.me/%s", username)
}
