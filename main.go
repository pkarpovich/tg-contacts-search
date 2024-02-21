package main

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/message"
	"github.com/gotd/td/tg"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/nyaruka/phonenumbers"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"os/signal"
	"path"
	"strings"
)

type Config struct {
	BotToken      string `env:"BOT_TOKEN"`
	AppHash       string `env:"APP_HASH"`
	AppId         int    `env:"APP_ID"`
	Phone         string `env:"PHONE"`
	Password      string `env:"PASSWORD"`
	SessionFolder string `env:"SESSION_FOLDER" env-default:"./sessions"`
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

func codePrompt(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

func run(ctx context.Context, log *zap.Logger) error {
	userClient := telegram.NewClient(cfg.AppId, cfg.AppHash, telegram.Options{
		SessionStorage: &session.FileStorage{
			Path: path.Join(cfg.SessionFolder, ".tg-session.json"),
		},
	})
	flow := auth.NewFlow(
		auth.Constant(cfg.Phone, cfg.Password, auth.CodeAuthenticatorFunc(codePrompt)),
		auth.SendCodeOptions{},
	)

	return userClient.Run(ctx, func(ctx context.Context) error {
		if err := userClient.Auth().IfNecessary(ctx, flow); err != nil {
			return fmt.Errorf("auth: %w", err)
		}

		updates := tg.NewUpdateDispatcher()
		opts := telegram.Options{
			UpdateHandler: updates,
			Logger:        log,
		}

		log.Info("Starting bot")

		return telegram.BotFromEnvironment(ctx, opts, func(ctx context.Context, client *telegram.Client) error {
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

				username, err := getUsername(userClient.API(), msg.Message)
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
	logger, err := zap.NewDevelopment(zap.IncreaseLevel(zapcore.InfoLevel), zap.AddStacktrace(zapcore.FatalLevel))
	if err != nil {
		log.Fatalf("Failed to create logger: %s", err)
	}

	err = godotenv.Load()
	if err != nil {
		logger.Warn("Failed to load .env file", zap.Error(err))
	}

	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		logger.Fatal("Failed to read environment variables", zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx, logger); err != nil {
		panic(err)
	}
}
