package config

type Config struct {
	BotToken      string `env:"BOT_TOKEN"`
	AppHash       string `env:"APP_HASH"`
	AppId         int    `env:"APP_ID"`
	Phone         string `env:"PHONE"`
	Password      string `env:"PASSWORD"`
	SessionFolder string `env:"SESSION_FOLDER" env-default:"./sessions"`
}
