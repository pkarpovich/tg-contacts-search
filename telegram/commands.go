package telegram

import (
	"context"
	"fmt"
	"github.com/gotd/contrib/bg"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/nyaruka/phonenumbers"
	"go.uber.org/zap"
)

func validatePhoneNum(phoneNum string) bool {
	parsedNum, err := phonenumbers.Parse(phoneNum, "ZZ")
	if err != nil {
		return false
	}

	return phonenumbers.IsValidNumber(parsedNum)
}

func (tl *Listener) getUsername(phoneNum string) (string, error) {
	ctx := context.Background()
	client := tl.NewUserClient()
	stop, err := bg.Connect(client)
	if err != nil {
		return "", err
	}
	defer func() { _ = stop() }()

	users, err := client.API().ContactsImportContacts(ctx, []tg.InputPhoneContact{{Phone: phoneNum}})
	if err != nil {
		return "", fmt.Errorf("failed to import contacts: %w", err)
	}

	if len(users.GetUsers()) > 0 {
		user, ok := users.GetUsers()[0].(*tg.User)
		if !ok {
			return "", nil
		}

		if err := tl.removeContact(user); err != nil {
			return "", fmt.Errorf("failed to remove contact: %w", err)
		}

		return user.Username, nil
	}

	return "", nil
}

func (tl *Listener) removeContact(user *tg.User) error {
	ctx := context.Background()
	client := tl.NewUserClient()
	stop, err := bg.Connect(client)
	if err != nil {
		return err
	}
	defer func() { _ = stop() }()

	_, err = client.API().ContactsDeleteContacts(ctx, []tg.InputUserClass{&tg.InputUser{
		UserID:     user.ID,
		AccessHash: user.AccessHash,
	}})
	if err != nil {
		return fmt.Errorf("failed to delete contacts: %w", err)
	}

	return nil
}

type MessageCtx struct {
	u        *tg.UpdateNewMessage
	msg      *tg.Message
	entities tg.Entities
}

func (tl *Listener) handleNewBotMessage(ctx context.Context, entities tg.Entities, u *tg.UpdateNewMessage) error {
	msg, ok := u.Message.(*tg.Message)
	if !ok {
		return nil
	}
	tl.Logger.Info("Bot Message", zap.Any("message", msg))

	msgContext := &MessageCtx{u, msg, entities}

	if msg.Message == "/start" {
		return tl.handleStartCommand(msgContext)
	}

	if msg.Message == "/ping" {
		return tl.handlePingCommand(msgContext)
	}

	return tl.processPhoneNum(msgContext)
}

func (tl *Listener) handleStartCommand(msgContext *MessageCtx) error {
	entities := msgContext.entities
	u := msgContext.u

	sender := message.NewSender(tg.NewClient(tl.BotClient))
	ctx := context.Background()

	_, err := sender.Reply(entities, u).Text(ctx, "Welcome to the contacts search bot! Send me a phone number to get the username of the user.")
	if err != nil {
		tl.Logger.Error("failed to send message", zap.Error(err))
	}

	return nil
}

func (tl *Listener) handlePingCommand(msgContext *MessageCtx) error {
	entities := msgContext.entities
	u := msgContext.u

	sender := message.NewSender(tg.NewClient(tl.BotClient))
	ctx := context.Background()

	_, err := tl.GetSelfUsername()
	if err != nil {
		_, err := sender.Reply(entities, u).Text(ctx, fmt.Sprintf("Failed to access user API: %s", err))
		if err != nil {
			tl.Logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}

	_, err = sender.Reply(entities, u).Text(ctx, "pong")
	if err != nil {
		tl.Logger.Error("failed to send message", zap.Error(err))
	}

	return nil
}

func (tl *Listener) processPhoneNum(msgContext *MessageCtx) error {
	entities := msgContext.entities
	u := msgContext.u

	sender := message.NewSender(tg.NewClient(tl.BotClient))
	ctx := context.Background()

	msg := msgContext.msg.Message

	if !validatePhoneNum(msg) {
		_, err := sender.Reply(entities, u).Text(ctx, "Invalid phone number")
		if err != nil {
			tl.Logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}

	username, err := tl.getUsername(msg)
	if err != nil {
		_, err := sender.Reply(entities, u).Text(ctx, fmt.Sprintf("Failed to get username: %s", err))
		if err != nil {
			tl.Logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}

	if username == "" {
		_, err := sender.Reply(entities, u).Text(ctx, "User not found")
		if err != nil {
			tl.Logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}

	_, err = sender.Reply(entities, u).Text(ctx, fmt.Sprintf("https://t.me/%s", username))
	if err != nil {
		tl.Logger.Error("failed to send message", zap.Error(err))
		return nil
	}

	return nil
}
