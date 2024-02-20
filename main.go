package main

import (
	"context"
	"fmt"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/nyaruka/phonenumbers"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"os/signal"
)

type Config struct {
	BotToken string `env:"BOT_TOKEN"`
	AppHash  string `env:"APP_HASH"`
	AppId    int    `env:"APP_ID"`
}

var cfg Config

func validatePhoneNum(phoneNum string) bool {
	parsedNum, err := phonenumbers.Parse(phoneNum, "ZZ")
	if err != nil {
		return false
	}

	return phonenumbers.IsValidNumber(parsedNum)
}

func getUsername(client *tg.Client, phoneNum string) (string, error) {
	users, err := client.ContactsImportContacts(context.Background(), []tg.InputPhoneContact{{Phone: phoneNum}})
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

func run(ctx context.Context, log *zap.Logger) error {
	client := telegram.NewClient(cfg.AppId, cfg.AppHash, telegram.Options{})

	return client.Run(ctx, func(ctx context.Context) error {
		updates := tg.NewUpdateDispatcher()
		opts := telegram.Options{
			UpdateHandler: updates,
			Logger:        log,
		}

		log.Info("Starting bot")

		return telegram.BotFromEnvironment(ctx, opts, func(ctx context.Context, client *telegram.Client) error {
			api := tg.NewClient(client)
			sender := message.NewSender(tg.NewClient(client))

			updates.OnNewMessage(func(ctx context.Context, entities tg.Entities, u *tg.UpdateNewMessage) error {
				msg, ok := u.Message.(*tg.Message)
				if !ok {
					return nil
				}
				log.Info("Message", zap.Any("message", u.Message))

				if msg.Message == "ping" {
					_, err := sender.Reply(entities, u).Text(ctx, "pong")
					if err != nil {
						log.Error("failed to send message", zap.Error(err))
					}

					return nil
				}

				if !validatePhoneNum(msg.Message) {
					_, err := sender.Reply(entities, u).Text(ctx, "Invalid phone number")
					if err != nil {
						log.Error("failed to send message", zap.Error(err))
					}

					return nil
				}

				username, err := getUsername(api, msg.Message)
				if err != nil {
					_, err := sender.Reply(entities, u).Text(ctx, fmt.Sprintf("Failed to get username: %s", err))
					if err != nil {
						log.Error("failed to send message", zap.Error(err))
					}

					return nil
				}

				if username == "" {
					_, err := sender.Reply(entities, u).Text(ctx, "User not found")
					if err != nil {
						log.Error("failed to send message", zap.Error(err))
					}

					return nil
				}

				_, err = sender.Reply(entities, u).Text(ctx, fmt.Sprintf("https://t.me/%s", username))
				if err != nil {
					log.Error("failed to send message", zap.Error(err))
					return nil
				}

				return nil
			})

			return nil
		}, telegram.RunUntilCanceled)
	})
}

func main() {
	err := cleanenv.ReadEnv(&cfg)
	if err != nil {
		log.Fatalf("Failed to read config.yml: %s", err)
	}

	logger, err := zap.NewDevelopment(zap.IncreaseLevel(zapcore.InfoLevel), zap.AddStacktrace(zapcore.FatalLevel))
	if err != nil {
		log.Fatalf("Failed to create logger: %s", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx, logger); err != nil {
		panic(err)
	}
}
