package telegram

import (
	"context"
	"fmt"
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
	clientApi := tl.UserClient.API()

	users, err := clientApi.ContactsImportContacts(context.Background(), []tg.InputPhoneContact{{Phone: phoneNum}})
	if err != nil {
		return "", fmt.Errorf("failed to import contacts: %w", err)
	}

	if len(users.GetUsers()) > 0 {
		user, ok := users.GetUsers()[0].(*tg.User)
		if !ok {
			return "", nil
		}

		return user.Username, nil
	}

	return "", nil
}

func (tl *Listener) handleNewBotMessage(ctx context.Context, entities tg.Entities, u *tg.UpdateNewMessage) error {
	sender := message.NewSender(tg.NewClient(tl.UserClient))

	msg, ok := u.Message.(*tg.Message)
	if !ok {
		return nil
	}
	tl.Logger.Info("Message", zap.Any("message", u.Message))

	if msg.Message == "/start" {
		_, err := sender.Reply(entities, u).Text(ctx, "Welcome to the contacts search bot! Send me a phone number to get the username of the user.")
		if err != nil {
			tl.Logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}

	if msg.Message == "/ping" {
		_, err := sender.Reply(entities, u).Text(ctx, "pong")
		if err != nil {
			tl.Logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}

	if !validatePhoneNum(msg.Message) {
		_, err := sender.Reply(entities, u).Text(ctx, "Invalid phone number")
		if err != nil {
			tl.Logger.Error("failed to send message", zap.Error(err))
		}

		return nil
	}

	username, err := tl.getUsername(msg.Message)
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
