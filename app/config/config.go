package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log"
)

type TelegramConfig struct {
	BotToken      string `env:"BOT_TOKEN"`
	AppHash       string `env:"APP_HASH"`
	AppId         int    `env:"APP_ID"`
	Phone         string `env:"PHONE"`
	Password      string `env:"PASSWORD"`
	SessionFolder string `env:"SESSION_FOLDER" env-default:"./sessions"`
}

type Config struct {
	Telegram TelegramConfig
}

func Init() (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		log.Printf("[WARN] error while loading .env file: %v", err)
	}

	var cfg Config
	err = cleanenv.ReadEnv(&cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}
