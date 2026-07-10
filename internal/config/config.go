package config

import (
	"errors"

	"github.com/caarlos0/env/v11"
)

type Config struct {
	HueBridgeHost string `env:"HUE_BRIDGE_HOST"`

	// Optional: obtained once via the bridge's link-button pairing flow. If
	// unset here, cmd/web falls back to a value stored in the settings table
	// by the /setup pairing flow (see internal/hue.Pair).
	HueAppKey string `env:"HUE_APP_KEY"`

	// Optional; must be set together (see validate).
	TelegramBotToken string `env:"TELEGRAM_BOT_TOKEN"`
	TelegramChatID   string `env:"TELEGRAM_CHAT_ID"`

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
	return nil
}
