package db

import (
	"context"
	"database/sql"
	"errors"
)

// TelegramEnabledKey is the settings key controlling whether the Telegram
// notification channel is active. Missing means never toggled by the user.
const TelegramEnabledKey = "telegram_enabled"

// HueAppKeyKey is the settings key the /setup pairing flow stores the Hue
// Bridge application key under, used when HUE_APP_KEY isn't set via env.
const HueAppKeyKey = "hue_app_key"

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
