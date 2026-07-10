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
	TelegramEnabled    bool   `json:"telegram_enabled"`
	TelegramConfigured bool   `json:"telegram_configured"`
	HueBridgeHost      string `json:"hue_bridge_host"`
	BridgeOnline       bool   `json:"bridge_online"`
	Version            string `json:"version"`
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
	if !got.TelegramConfigured {
		t.Error("TelegramConfigured = false, want true")
	}
	if got.HueBridgeHost != "192.168.1.10" {
		t.Errorf("HueBridgeHost = %q, want %q", got.HueBridgeHost, "192.168.1.10")
	}
	if body := rec.Body.String(); strings.Contains(body, "secret-token") || strings.Contains(body, "secret-chat-id") {
		t.Errorf("response leaked a secret: %s", body)
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
	if !got.TelegramEnabled {
		t.Error("TelegramEnabled = false, want true (default) when never toggled")
	}
}

func TestPutTelegramEnabled(t *testing.T) {
	h, _ := newTestSetup(t, &mockHue{})

	rec, c := newCtx(t, http.MethodPut, "/api/settings/telegram-enabled", `{"enabled":false}`)
	if err := h.PutTelegramEnabled(c); err != nil {
		t.Fatalf("PutTelegramEnabled: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	rec, c = newCtx(t, http.MethodGet, "/api/settings", "")
	if err := h.GetSettings(c); err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	got := decodeJSON[settingsResp](t, rec.Body.Bytes())
	if got.TelegramEnabled {
		t.Error("TelegramEnabled = true, want false after disabling")
	}

	rec, c = newCtx(t, http.MethodPut, "/api/settings/telegram-enabled", `{"enabled":true}`)
	if err := h.PutTelegramEnabled(c); err != nil {
		t.Fatalf("PutTelegramEnabled (re-enable): %v", err)
	}
	rec, c = newCtx(t, http.MethodGet, "/api/settings", "")
	if err := h.GetSettings(c); err != nil {
		t.Fatalf("GetSettings: %v", err)
	}
	got = decodeJSON[settingsResp](t, rec.Body.Bytes())
	if !got.TelegramEnabled {
		t.Error("TelegramEnabled = false, want true after re-enabling")
	}
}

func TestPutTelegramEnabled_InvalidRequest(t *testing.T) {
	h, _ := newTestSetup(t, &mockHue{})

	t.Run("missing enabled field", func(t *testing.T) {
		rec, c := newCtx(t, http.MethodPut, "/api/settings/telegram-enabled", `{}`)
		if err := h.PutTelegramEnabled(c); err != nil {
			t.Fatalf("PutTelegramEnabled: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		rec, c := newCtx(t, http.MethodPut, "/api/settings/telegram-enabled", `not json`)
		if err := h.PutTelegramEnabled(c); err != nil {
			t.Fatalf("PutTelegramEnabled: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}
