package config

import (
	"errors"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	HueBridgeHost string `env:"HUE_BRIDGE_HOST"`

	// Optional — falls back to a DB-stored value from /setup if unset (see internal/hue.Pair).
	HueAppKey string `env:"HUE_APP_KEY"`

	// Optional; must be set together (see validate). Mutually exclusive with DiscordWebhookURL.
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN"`
	TelegramChatID   string `env:"TELEGRAM_CHAT_ID"`

	// Optional; mutually exclusive with TelegramBotToken/TelegramChatID.
	DiscordWebhookURL string `env:"DISCORD_WEBHOOK_URL"`

	DBPath string `env:"DB_PATH" envDefault:"data/watchue.db"`
}

func Load() (*Config, error) {
	cfg := &Config{}
	if err := env.Parse(cfg); err != nil {
		return nil, err
	}
	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.HueBridgeHost == "" {
		return errors.New("HUE_BRIDGE_HOST must be set")
	}
	if (c.TelegramBotToken == "") != (c.TelegramChatID == "") {
		return errors.New("TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID must be set together")
	}
	telegramConfigured := c.TelegramBotToken != "" && c.TelegramChatID != ""
	if telegramConfigured && c.DiscordWebhookURL != "" {
		return errors.New("TELEGRAM_BOT_TOKEN/TELEGRAM_CHAT_ID and DISCORD_WEBHOOK_URL cannot both be set — only one notification provider may be active")
	}
	return nil
}
