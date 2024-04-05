package telegram

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gotd/contrib/bg"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
	"go.uber.org/zap"
	"os"
	"path/filepath"
	"strings"
	"time"
)

func codePrompt(ctx context.Context, sentCode *tg.AuthSentCode) (string, error) {
	fmt.Print("Enter code: ")
	code, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(code), nil
}

func deviceConfig() telegram.DeviceConfig {
	return telegram.DeviceConfig{
		DeviceModel:    "Mozilla/5.0 (X11; Ubuntu; Linux x86_64; rv:109.0) Gecko/20100101 Firefox/116.0",
		SystemVersion:  "Win32",
		AppVersion:     "2.1.9 K",
		LangPack:       "webk",
		SystemLangCode: "en",
		LangCode:       "en",
	}
}

func (tl *Listener) NewUserClient() *telegram.Client {
	dispatcher := tg.NewUpdateDispatcher()
	manager := updates.New(updates.Config{
		Handler: dispatcher,
	})

	return telegram.NewClient(tl.Config.AppId, tl.Config.AppHash, telegram.Options{
		SessionStorage: &telegram.FileSessionStorage{
			Path: filepath.Join(tl.Config.SessionFolder, ".tg-user-session.json"),
		},
		DialTimeout:   time.Minute * 5,
		Device:        deviceConfig(),
		Logger:        tl.Logger,
		UpdateHandler: manager,
		DC:            5,
	})
}

func (tl *Listener) StartUserClient() error {
	ctx := context.Background()
	client := tl.NewUserClient()

	flow := auth.NewFlow(
		auth.Constant(tl.Config.Phone, tl.Config.Password, auth.CodeAuthenticatorFunc(codePrompt)),
		auth.SendCodeOptions{},
	)

	stop, err := bg.Connect(client)
	if err != nil {
		return err
	}
	defer func() { _ = stop() }()

	if err := client.Auth().IfNecessary(ctx, flow); err != nil {
		return err
	}

	tl.Logger.Info("User client started successfully.")
	return nil
}

func (tl *Listener) handleNewUserMessage(ctx context.Context, entities tg.Entities, u *tg.UpdateNewMessage) error {
	msg, ok := u.Message.(*tg.Message)
	if !ok {
		tl.Logger.Warn("Not a message", zap.Any("update", u))
		return nil
	}

	tl.Logger.Info("User Message", zap.Any("message", msg))
	return nil
}

func (tl *Listener) GetSelfUsername() (string, error) {
	ctx := context.Background()
	client := tl.NewUserClient()

	stop, err := bg.Connect(client)
	if err != nil {
		return "", err
	}
	defer func() { _ = stop() }()

	status, err := client.Auth().Status(ctx)
	if err != nil {
		return "", err
	}

	if !status.Authorized {
		return "", fmt.Errorf("user client is not authorized")
	}

	return status.User.Username, nil
}
