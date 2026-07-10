package handler_test

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/florentsorel/watchue/internal/config"
	"github.com/florentsorel/watchue/internal/db"
	"github.com/florentsorel/watchue/internal/handler"
)

type notifyTestResp struct {
	OK bool `json:"ok"`
}

func stubFactory(n handler.Notifier, err error) handler.NotifierFactory {
	return func(cfg handler.NotifyConfig) (handler.Notifier, error) { return n, err }
}

func TestPostNotifyTest_Success(t *testing.T) {
	h, _, _ := newSetupTestSetupWithNotify(t, &config.Config{HueBridgeHost: "192.168.1.10"}, nil, stubFactory(&mockNotifier{}, nil))

	rec, c := newCtx(t, http.MethodPost, "/api/notify/test", `{"provider":"telegram","telegram_bot_token":"t","telegram_chat_id":"c"}`)
	if err := h.PostNotifyTest(c); err != nil {
		t.Fatalf("PostNotifyTest: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	got := decodeJSON[notifyTestResp](t, rec.Body.Bytes())
	if !got.OK {
		t.Error("OK = false, want true")
	}
}

func TestPostNotifyTest_FactoryError(t *testing.T) {
	h, _, _ := newSetupTestSetupWithNotify(t, &config.Config{HueBridgeHost: "192.168.1.10"}, nil, stubFactory(nil, errors.New("invalid webhook url")))

	rec, c := newCtx(t, http.MethodPost, "/api/notify/test", `{"provider":"discord","discord_webhook_url":"not-a-real-url"}`)
	if err := h.PostNotifyTest(c); err != nil {
		t.Fatalf("PostNotifyTest: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPostNotifyTest_FallsBackToStoredCredentialsWhenBlank(t *testing.T) {
	var gotCfg handler.NotifyConfig
	factory := func(cfg handler.NotifyConfig) (handler.Notifier, error) {
		gotCfg = cfg
		return &mockNotifier{}, nil
	}
	h, _, _ := newSetupTestSetupWithNotify(t, &config.Config{HueBridgeHost: "192.168.1.10"}, nil, factory)

	_, c := newCtx(t, http.MethodPost, "/api/notify", `{"provider":"discord","discord_webhook_url":"https://discord.example/webhook"}`)
	if err := h.PostNotify(c); err != nil {
		t.Fatalf("PostNotify: %v", err)
	}

	// No credentials in the request body — must fall back to what's stored,
	// letting a user re-verify an already-configured provider without
	// retyping its webhook URL.
	rec, c := newCtx(t, http.MethodPost, "/api/notify/test", `{"provider":"discord"}`)
	if err := h.PostNotifyTest(c); err != nil {
		t.Fatalf("PostNotifyTest: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	if gotCfg.DiscordWebhookURL != "https://discord.example/webhook" {
		t.Errorf("factory received %+v, want the stored webhook url", gotCfg)
	}
}

func TestPostNotifyTest_BlankAndNotConfigured(t *testing.T) {
	h, _, _ := newSetupTestSetupWithNotify(t, &config.Config{HueBridgeHost: "192.168.1.10"}, nil, stubFactory(&mockNotifier{}, nil))

	rec, c := newCtx(t, http.MethodPost, "/api/notify/test", `{"provider":"discord"}`)
	if err := h.PostNotifyTest(c); err != nil {
		t.Fatalf("PostNotifyTest: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPostNotifyTest_SendError(t *testing.T) {
	h, _, _ := newSetupTestSetupWithNotify(t, &config.Config{HueBridgeHost: "192.168.1.10"}, nil,
		stubFactory(&mockNotifier{sendTestErr: errors.New("unauthorized")}, nil))

	rec, c := newCtx(t, http.MethodPost, "/api/notify/test", `{"provider":"telegram","telegram_bot_token":"t","telegram_chat_id":"c"}`)
	if err := h.PostNotifyTest(c); err != nil {
		t.Fatalf("PostNotifyTest: %v", err)
	}
	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadGateway)
	}
}

func TestPostNotify_SavesAndSwapsNotifier(t *testing.T) {
	notifier := &mockNotifier{}
	h, queries, _ := newSetupTestSetupWithNotify(t, &config.Config{HueBridgeHost: "192.168.1.10"}, nil, stubFactory(notifier, nil))

	rec, c := newCtx(t, http.MethodPost, "/api/notify", `{"provider":"discord","discord_webhook_url":"https://discord.example/webhook"}`)
	if err := h.PostNotify(c); err != nil {
		t.Fatalf("PostNotify: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	provider, err := queries.GetSetting(context.Background(), db.NotifyProviderKey)
	if err != nil || provider != "discord" {
		t.Errorf("stored provider = %q, err = %v, want %q", provider, err, "discord")
	}
	url, err := queries.GetSetting(context.Background(), db.NotifyDiscordWebhookURLKey)
	if err != nil || url != "https://discord.example/webhook" {
		t.Errorf("stored webhook url = %q, err = %v", url, err)
	}
}

func TestPostNotify_ConflictWhenEnvConfigured(t *testing.T) {
	called := false
	factory := func(cfg handler.NotifyConfig) (handler.Notifier, error) {
		called = true
		return &mockNotifier{}, nil
	}
	h, _, _ := newSetupTestSetupWithNotify(t, &config.Config{
		HueBridgeHost: "192.168.1.10", TelegramBotToken: "t", TelegramChatID: "c",
	}, nil, factory)

	rec, c := newCtx(t, http.MethodPost, "/api/notify", `{"provider":"discord","discord_webhook_url":"https://discord.example/webhook"}`)
	if err := h.PostNotify(c); err != nil {
		t.Fatalf("PostNotify: %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
	if called {
		t.Error("notifierFactory must not be invoked when already configured via env")
	}
}

func TestPostNotify_RejectsIncompleteConfig(t *testing.T) {
	h, _, _ := newSetupTestSetupWithNotify(t, &config.Config{HueBridgeHost: "192.168.1.10"}, nil,
		stubFactory(nil, errors.New("telegram_bot_token and telegram_chat_id are required")))

	rec, c := newCtx(t, http.MethodPost, "/api/notify", `{"provider":"telegram"}`)
	if err := h.PostNotify(c); err != nil {
		t.Fatalf("PostNotify: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPostNotify_NeverEchoesCredentialsBack(t *testing.T) {
	h, _, _ := newSetupTestSetupWithNotify(t, &config.Config{HueBridgeHost: "192.168.1.10"}, nil, stubFactory(&mockNotifier{}, nil))

	rec, c := newCtx(t, http.MethodPost, "/api/notify", `{"provider":"telegram","telegram_bot_token":"super-secret","telegram_chat_id":"c"}`)
	if err := h.PostNotify(c); err != nil {
		t.Fatalf("PostNotify: %v", err)
	}
	if body := rec.Body.String(); strings.Contains(body, "super-secret") {
		t.Errorf("response leaked a credential: %s", body)
	}
}

type notifyStatusResp struct {
	ActiveProvider string                   `json:"active_provider"`
	EnvLocked      bool                     `json:"env_locked"`
	Telegram       notifyProviderStatusJSON `json:"telegram"`
	Discord        notifyProviderStatusJSON `json:"discord"`
}

type notifyProviderStatusJSON struct {
	Configured bool `json:"configured"`
}

func TestGetNotify_ReportsBothProvidersIndependently(t *testing.T) {
	h, _, _ := newSetupTestSetupWithNotify(t, &config.Config{HueBridgeHost: "192.168.1.10"}, nil, stubFactory(&mockNotifier{}, nil))

	// Save telegram first, then switch to discord — reproduces the bug
	// report: switching the active provider must not forget telegram.
	_, c := newCtx(t, http.MethodPost, "/api/notify", `{"provider":"telegram","telegram_bot_token":"t","telegram_chat_id":"c"}`)
	if err := h.PostNotify(c); err != nil {
		t.Fatalf("PostNotify (telegram): %v", err)
	}
	_, c = newCtx(t, http.MethodPost, "/api/notify", `{"provider":"discord","discord_webhook_url":"https://discord.example/webhook"}`)
	if err := h.PostNotify(c); err != nil {
		t.Fatalf("PostNotify (discord): %v", err)
	}

	rec, c := newCtx(t, http.MethodGet, "/api/notify", "")
	if err := h.GetNotify(c); err != nil {
		t.Fatalf("GetNotify: %v", err)
	}
	got := decodeJSON[notifyStatusResp](t, rec.Body.Bytes())
	if got.ActiveProvider != "discord" {
		t.Errorf("ActiveProvider = %q, want %q", got.ActiveProvider, "discord")
	}
	if !got.Telegram.Configured {
		t.Error("Telegram.Configured = false, want true — switching to discord must not forget it")
	}
	if !got.Discord.Configured {
		t.Error("Discord.Configured = false, want true")
	}
	if got.EnvLocked {
		t.Error("EnvLocked = true, want false — nothing configured via env in this test")
	}
}

func TestGetNotify_ReportsEnvLocked(t *testing.T) {
	h, _, _ := newSetupTestSetupWithNotify(t, &config.Config{
		HueBridgeHost: "192.168.1.10", DiscordWebhookURL: "https://discord.example/webhook",
	}, nil, stubFactory(&mockNotifier{}, nil))

	rec, c := newCtx(t, http.MethodGet, "/api/notify", "")
	if err := h.GetNotify(c); err != nil {
		t.Fatalf("GetNotify: %v", err)
	}
	got := decodeJSON[notifyStatusResp](t, rec.Body.Bytes())
	if !got.EnvLocked {
		t.Error("EnvLocked = false, want true")
	}
	if got.ActiveProvider != "discord" {
		t.Errorf("ActiveProvider = %q, want %q", got.ActiveProvider, "discord")
	}
}

func TestPostNotifyActivate_ReusesStoredCredentials(t *testing.T) {
	var gotCfg handler.NotifyConfig
	var lastNotifier *mockNotifier
	factory := func(cfg handler.NotifyConfig) (handler.Notifier, error) {
		gotCfg = cfg
		lastNotifier = &mockNotifier{}
		return lastNotifier, nil
	}
	h, _, _ := newSetupTestSetupWithNotify(t, &config.Config{HueBridgeHost: "192.168.1.10"}, nil, factory)

	_, c := newCtx(t, http.MethodPost, "/api/notify", `{"provider":"telegram","telegram_bot_token":"t-token","telegram_chat_id":"t-chat"}`)
	if err := h.PostNotify(c); err != nil {
		t.Fatalf("PostNotify (telegram): %v", err)
	}
	_, c = newCtx(t, http.MethodPost, "/api/notify", `{"provider":"discord","discord_webhook_url":"https://discord.example/webhook"}`)
	if err := h.PostNotify(c); err != nil {
		t.Fatalf("PostNotify (discord): %v", err)
	}

	rec, c := newCtx(t, http.MethodPost, "/api/notify/activate", `{"provider":"telegram"}`)
	if err := h.PostNotifyActivate(c); err != nil {
		t.Fatalf("PostNotifyActivate: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if gotCfg.Provider != "telegram" || gotCfg.TelegramBotToken != "t-token" || gotCfg.TelegramChatID != "t-chat" {
		t.Errorf("factory received %+v, want the previously-stored telegram credentials", gotCfg)
	}
	if lastNotifier.testCalls != 0 {
		t.Errorf("SendTest called %d times, want 0 — activation must not send a notification", lastNotifier.testCalls)
	}

	rec, c = newCtx(t, http.MethodGet, "/api/notify", "")
	if err := h.GetNotify(c); err != nil {
		t.Fatalf("GetNotify: %v", err)
	}
	got := decodeJSON[notifyStatusResp](t, rec.Body.Bytes())
	if got.ActiveProvider != "telegram" {
		t.Errorf("ActiveProvider = %q, want %q", got.ActiveProvider, "telegram")
	}
	if !got.Discord.Configured {
		t.Error("Discord.Configured = false, want true — activating telegram must not clear it")
	}
}

func TestPostNotifyActivate_NotConfigured(t *testing.T) {
	h, _, _ := newSetupTestSetupWithNotify(t, &config.Config{HueBridgeHost: "192.168.1.10"}, nil, stubFactory(&mockNotifier{}, nil))

	rec, c := newCtx(t, http.MethodPost, "/api/notify/activate", `{"provider":"discord"}`)
	if err := h.PostNotifyActivate(c); err != nil {
		t.Fatalf("PostNotifyActivate: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPostNotifyActivate_ConflictWhenEnvConfigured(t *testing.T) {
	h, _, _ := newSetupTestSetupWithNotify(t, &config.Config{
		HueBridgeHost: "192.168.1.10", TelegramBotToken: "t", TelegramChatID: "c",
	}, nil, stubFactory(&mockNotifier{}, nil))

	rec, c := newCtx(t, http.MethodPost, "/api/notify/activate", `{"provider":"discord"}`)
	if err := h.PostNotifyActivate(c); err != nil {
		t.Fatalf("PostNotifyActivate: %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
}
