package main

import (
	"context"
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"github.com/pkarpovich/tg-contacts-search/app/config"
	"github.com/pkarpovich/tg-contacts-search/app/telegram"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"log"
	"os"
	"os/signal"
)

func run(ctx context.Context, cfg config.Config, log *zap.Logger) error {
	listener := telegram.NewListener(log, cfg)

	if err := listener.StartUserClient(); err != nil {
		log.Error("Failed to start user client", zap.Error(err))
	}

	if err := listener.StartBotClient(ctx); err != nil {
		log.Error("Failed to start bot client", zap.Error(err))
	}

	return nil
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

	var cfg config.Config
	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		logger.Fatal("Failed to read environment variables", zap.Error(err))
	}

	ctx, cancel := signal.NotifyContext(context.Background(), os.Interrupt)
	defer cancel()
	if err := run(ctx, cfg, logger); err != nil {
		panic(err)
	}
}
