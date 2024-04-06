package main

import (
	"context"
	"github.com/pkarpovich/tg-contacts-search/app/config"
	"github.com/pkarpovich/tg-contacts-search/app/telegram/bot"
	"github.com/pkarpovich/tg-contacts-search/app/telegram/user"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"os/signal"
)

func run(ctx context.Context, cfg *config.Config, logger *zap.Logger) error {
	if err := user.CheckAuthState(logger, cfg.Telegram); err != nil {
		logger.Error("Failed to check auth state", zap.Error(err))
		return err
	}

	if err := bot.NewClient(logger, cfg.Telegram).Listen(); err != nil {
		logger.Error("Failed to start bot client", zap.Error(err))
		return err
	}

	return nil
}

func main() {
	logger, err := zap.NewDevelopment(zap.IncreaseLevel(zapcore.InfoLevel), zap.AddStacktrace(zapcore.FatalLevel))
	if err != nil {
		log.Fatalf("Failed to create logger: %s", err)
	}

	cfg, err := config.Init()
	if err != nil {
		log.Fatalf("Failed to read config: %s", err)
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx, cfg, logger); err != nil {
		panic(err)
	}
}
