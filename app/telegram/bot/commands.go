package bot

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/pkarpovich/tg-contacts-search/app/telegram/user"
	"github.com/pkarpovich/tg-contacts-search/app/utils"
	"go.uber.org/zap"
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

func (c *Client) processPhoneNum(msgContext *MessageCtx) error {
	entities := msgContext.entities
	msg := msgContext.msg.Message
	u := msgContext.u

	sender := message.NewSender(tg.NewClient(c.client))
	ctx := context.Background()

	if !utils.ValidatePhoneNum(msg) {
		_, err := sender.Reply(entities, u).Text(ctx, "Invalid phone number")
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
		_, err := sender.Reply(entities, u).Text(ctx, "User not found")
		if err != nil {
			c.logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}

	_, err = sender.Reply(entities, u).Text(ctx, fmt.Sprintf("https://t.me/%s", username))
	if err != nil {
		c.logger.Error("failed to send message", zap.Error(err))
		return nil
	}

	return nil
}
