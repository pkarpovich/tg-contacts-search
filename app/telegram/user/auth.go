package user

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"github.com/pkarpovich/tg-contacts-search/app/config"
	"go.uber.org/zap"
	"os"
	"strings"
)

func CheckAuthState(logger *zap.Logger, cfg config.TelegramConfig) error {
	client, stop, err := NewClient(logger, cfg)
	if err != nil {
		return err
	}
	defer func() { _ = stop() }()

	flow := auth.NewFlow(
		auth.Constant(cfg.Phone, cfg.Password, auth.CodeAuthenticatorFunc(codePrompt)),
		auth.SendCodeOptions{},
	)

	if err := client.auth.IfNecessary(context.Background(), flow); err != nil {
		return err
	}

	logger.Info("User client started successfully.")
	return nil
}

func codePrompt(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}
