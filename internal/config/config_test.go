package config_test

import (
	"strings"
	"testing"

	"github.com/florentsorel/watchue/internal/config"
)

func TestLoad_Validation(t *testing.T) {
	tests := []struct {
		name    string
		env     map[string]string
		wantErr string
	}{
		{
			name:    "missing bridge host",
			env:     map[string]string{"HUE_APP_KEY": "key"},
			wantErr: "HUE_BRIDGE_HOST",
		},
		{
			name:    "missing app key",
			env:     map[string]string{"HUE_BRIDGE_HOST": "192.168.1.10"},
			wantErr: "HUE_APP_KEY",
		},
		{
			name: "both set",
			env:  map[string]string{"HUE_BRIDGE_HOST": "192.168.1.10", "HUE_APP_KEY": "key"},
		},
		{
			name: "telegram not configured at all is fine",
			env:  map[string]string{"HUE_BRIDGE_HOST": "192.168.1.10", "HUE_APP_KEY": "key"},
		},
		{
			name: "telegram fully configured is fine",
			env: map[string]string{
				"HUE_BRIDGE_HOST": "192.168.1.10", "HUE_APP_KEY": "key",
				"TELEGRAM_BOT_TOKEN": "bot-token", "TELEGRAM_CHAT_ID": "chat-id",
			},
		},
		{
			name: "telegram bot token without chat id",
			env: map[string]string{
				"HUE_BRIDGE_HOST": "192.168.1.10", "HUE_APP_KEY": "key",
				"TELEGRAM_BOT_TOKEN": "bot-token",
			},
			wantErr: "TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID",
		},
		{
			name: "telegram chat id without bot token",
			env: map[string]string{
				"HUE_BRIDGE_HOST": "192.168.1.10", "HUE_APP_KEY": "key",
				"TELEGRAM_CHAT_ID": "chat-id",
			},
			wantErr: "TELEGRAM_BOT_TOKEN and TELEGRAM_CHAT_ID",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			for k, v := range tt.env {
				t.Setenv(k, v)
			}

			_, err := config.Load()
			if tt.wantErr != "" {
				if err == nil {
					t.Fatal("expected error, got nil")
				}
				if !strings.Contains(err.Error(), tt.wantErr) {
					t.Errorf("error %q does not mention %q", err.Error(), tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}
