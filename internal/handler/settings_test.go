package handler_test

import (
	"net/http"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/florentsorel/watchue/internal/config"
	"github.com/florentsorel/watchue/internal/stream"
)

type settingsResp struct {
	NotifyEnabled    bool   `json:"notify_enabled"`
	NotifyConfigured bool   `json:"notify_configured"`
	NotifyProvider   string `json:"notify_provider"`
	HueBridgeHost    string `json:"hue_bridge_host"`
	BridgeOnline     bool   `json:"bridge_online"`
	Version          string `json:"version"`
}

func TestSettings_ExposesConfigStatusWithoutSecrets(t *testing.T) {
	cfg := &config.Config{
		HueBridgeHost:    "192.168.1.10",
		TelegramBotToken: "secret-token",
		TelegramChatID:   "secret-chat-id",
	}
	h, _ := newTestSetupWithConfig(t, &mockHue{}, cfg)

	rec, c := newCtx(t, http.MethodGet, "/api/settings", "")
	if err := h.GetSettings(c); err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	got := decodeJSON[settingsResp](t, rec.Body.Bytes())
	if !got.NotifyConfigured {
		t.Error("NotifyConfigured = false, want true")
	}
	if got.NotifyProvider != "telegram" {
		t.Errorf("NotifyProvider = %q, want %q", got.NotifyProvider, "telegram")
	}
	if got.HueBridgeHost != "192.168.1.10" {
		t.Errorf("HueBridgeHost = %q, want %q", got.HueBridgeHost, "192.168.1.10")
	}
	if body := rec.Body.String(); strings.Contains(body, "secret-token") || strings.Contains(body, "secret-chat-id") {
		t.Errorf("response leaked a secret: %s", body)
	}
}

func TestSettings_ExposesDiscordProviderWithoutSecrets(t *testing.T) {
	cfg := &config.Config{
		HueBridgeHost:     "192.168.1.10",
		DiscordWebhookURL: "https://discord.com/api/webhooks/1/secret-token",
	}
	h, _ := newTestSetupWithConfig(t, &mockHue{}, cfg)

	rec, c := newCtx(t, http.MethodGet, "/api/settings", "")
	if err := h.GetSettings(c); err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	got := decodeJSON[settingsResp](t, rec.Body.Bytes())
	if got.NotifyProvider != "discord" {
		t.Errorf("NotifyProvider = %q, want %q", got.NotifyProvider, "discord")
	}
	if body := rec.Body.String(); strings.Contains(body, "secret-token") {
		t.Errorf("response leaked the webhook URL: %s", body)
	}
}

func TestSettings_ReflectsBridgeOnlineState(t *testing.T) {
	bridgeOnline := &atomic.Bool{}
	bridgeOnline.Store(false)
	h, _ := newTestSetupWithBridgeOnline(t, &mockHue{}, &config.Config{}, stream.NewHub(), bridgeOnline)

	rec, c := newCtx(t, http.MethodGet, "/api/settings", "")
	if err := h.GetSettings(c); err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	got := decodeJSON[settingsResp](t, rec.Body.Bytes())
	if got.BridgeOnline {
		t.Error("BridgeOnline = true, want false")
	}

	bridgeOnline.Store(true)
	rec, c = newCtx(t, http.MethodGet, "/api/settings", "")
	if err := h.GetSettings(c); err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	got = decodeJSON[settingsResp](t, rec.Body.Bytes())
	if !got.BridgeOnline {
		t.Error("BridgeOnline = false, want true")
	}
}

func TestSettings_ExposesVersion(t *testing.T) {
	h, _ := newTestSetup(t, &mockHue{})

	rec, c := newCtx(t, http.MethodGet, "/api/settings", "")
	if err := h.GetSettings(c); err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	got := decodeJSON[settingsResp](t, rec.Body.Bytes())
	if got.Version != "test" {
		t.Errorf("Version = %q, want %q", got.Version, "test")
	}
}

func TestSettings_DefaultsToEnabled(t *testing.T) {
	h, _ := newTestSetup(t, &mockHue{})

	rec, c := newCtx(t, http.MethodGet, "/api/settings", "")
	if err := h.GetSettings(c); err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	got := decodeJSON[settingsResp](t, rec.Body.Bytes())
	if !got.NotifyEnabled {
		t.Error("NotifyEnabled = false, want true (default) when never toggled")
	}
}

func TestPutNotifyEnabled(t *testing.T) {
	h, _ := newTestSetup(t, &mockHue{})

	rec, c := newCtx(t, http.MethodPut, "/api/settings/notify-enabled", `{"enabled":false}`)
	if err := h.PutNotifyEnabled(c); err != nil {
		t.Fatalf("PutNotifyEnabled: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	rec, c = newCtx(t, http.MethodGet, "/api/settings", "")
	if err := h.GetSettings(c); err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	got := decodeJSON[settingsResp](t, rec.Body.Bytes())
	if got.NotifyEnabled {
		t.Error("NotifyEnabled = true, want false after disabling")
	}

	_, c = newCtx(t, http.MethodPut, "/api/settings/notify-enabled", `{"enabled":true}`)
	if err := h.PutNotifyEnabled(c); err != nil {
		t.Fatalf("PutNotifyEnabled (re-enable): %v", err)
	}
	rec, c = newCtx(t, http.MethodGet, "/api/settings", "")
	if err := h.GetSettings(c); err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	got = decodeJSON[settingsResp](t, rec.Body.Bytes())
	if !got.NotifyEnabled {
		t.Error("NotifyEnabled = false, want true after re-enabling")
	}
}

func TestPutNotifyEnabled_InvalidRequest(t *testing.T) {
	h, _ := newTestSetup(t, &mockHue{})

	t.Run("missing enabled field", func(t *testing.T) {
		rec, c := newCtx(t, http.MethodPut, "/api/settings/notify-enabled", `{}`)
		if err := h.PutNotifyEnabled(c); err != nil {
			t.Fatalf("PutNotifyEnabled: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		rec, c := newCtx(t, http.MethodPut, "/api/settings/notify-enabled", `not json`)
		if err := h.PutNotifyEnabled(c); err != nil {
			t.Fatalf("PutNotifyEnabled: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}
