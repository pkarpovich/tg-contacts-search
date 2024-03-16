package telegram

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gotd/td/session"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/tg"
	"os"
	"path"
	"strings"
)

func codePrompt(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

func (tl *Listener) StartUserClient(ctx context.Context) error {
	userUpdates := tg.NewUpdateDispatcher()
	tl.UserClient = telegram.NewClient(tl.Config.AppId, tl.Config.AppHash, telegram.Options{
		SessionStorage: &session.FileStorage{
			Path: path.Join(tl.Config.SessionFolder, ".tg-session.json"),
		},
		UpdateHandler: userUpdates,
		Logger:        tl.Logger,
	})

	flow := auth.NewFlow(
		auth.Constant(tl.Config.Phone, tl.Config.Password, auth.CodeAuthenticatorFunc(codePrompt)),
		auth.SendCodeOptions{},
	)

	return tl.UserClient.Run(ctx, func(ctx context.Context) error {
		if err := tl.UserClient.Auth().IfNecessary(ctx, flow); err != nil {
			return err
		}

		tl.Logger.Info("User client started successfully.")
		userUpdates.OnNewMessage(tl.handleNewBotMessage)

		return telegram.RunUntilCanceled(ctx, tl.UserClient)
	})
}
