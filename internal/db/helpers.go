package db

import (
	"context"
	"database/sql"
	"errors"
)

// NotifyEnabledKey controls whether the active notification channel is on
// (defaults to enabled). Key name predates multi-provider support.
const NotifyEnabledKey = "telegram_enabled"

// HueAppKeyKey is the settings key the /setup pairing flow stores the Hue
// Bridge application key under, used when HUE_APP_KEY isn't set via env.
const HueAppKeyKey = "hue_app_key"

// NotifyProviderKey is the settings key storing which provider ("telegram"
// or "discord") the /provider page saved, used when no provider is
// configured via env.
const NotifyProviderKey = "notify_provider"

// NotifyTelegramBotTokenKey and NotifyTelegramChatIDKey store the Telegram
// credentials entered via the /provider page.
const (
	NotifyTelegramBotTokenKey = "notify_telegram_bot_token"
	NotifyTelegramChatIDKey   = "notify_telegram_chat_id"
)

// NotifyDiscordWebhookURLKey stores the Discord webhook URL entered via the
// /provider page.
const NotifyDiscordWebhookURLKey = "notify_discord_webhook_url"

// GetBoolSetting returns the boolean value stored under key, or defaultValue
// if it has never been set.
func (q *Queries) GetBoolSetting(ctx context.Context, key string, defaultValue bool) (bool, error) {
	value, err := q.GetSetting(ctx, key)
	if errors.Is(err, sql.ErrNoRows) {
		return defaultValue, nil
	}
	if err != nil {
		return false, err
	}
	return value == "true", nil
}

// SetBoolSetting stores a boolean value under key.
func (q *Queries) SetBoolSetting(ctx context.Context, key string, value bool) error {
	v := "false"
	if value {
		v = "true"
	}
	return q.SetSetting(ctx, SetSettingParams{Key: key, Value: v})
}
