package handler_test

import (
	"context"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/florentsorel/watchue/internal/config"
	"github.com/florentsorel/watchue/internal/db"
	"github.com/florentsorel/watchue/internal/handler"
	"github.com/florentsorel/watchue/internal/hue"
	"github.com/florentsorel/watchue/internal/stream"
)

type setupStatusResp struct {
	Configured    bool   `json:"configured"`
	HueBridgeHost string `json:"hue_bridge_host"`
}

type pairResp struct {
	Paired bool   `json:"paired"`
	Reason string `json:"reason"`
}

func newSetupTestSetup(t *testing.T, cfg *config.Config, pair handler.PairFunc) (*handler.Handler, *db.Queries, *bool) {
	t.Helper()
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	queries := db.New(conn)
	stopped := false
	bridgeOnline := &atomic.Bool{}
	h := handler.New(&mockHue{}, queries, cfg, stream.NewHub(), bridgeOnline, "test", func() { stopped = true }, pair)
	return h, queries, &stopped
}

func TestGetSetupStatus_Unconfigured(t *testing.T) {
	h, _, _ := newSetupTestSetup(t, &config.Config{HueBridgeHost: "192.168.1.10"}, nil)

	rec, c := newCtx(t, http.MethodGet, "/api/setup/status", "")
	if err := h.GetSetupStatus(c); err != nil {
		t.Fatalf("GetSetupStatus: %v", err)
	}
	got := decodeJSON[setupStatusResp](t, rec.Body.Bytes())
	if got.Configured {
		t.Error("Configured = true, want false")
	}
	if got.HueBridgeHost != "192.168.1.10" {
		t.Errorf("HueBridgeHost = %q, want %q", got.HueBridgeHost, "192.168.1.10")
	}
}

func TestGetSetupStatus_ConfiguredViaEnv(t *testing.T) {
	h, _, _ := newSetupTestSetup(t, &config.Config{HueBridgeHost: "192.168.1.10", HueAppKey: "env-key"}, nil)

	rec, c := newCtx(t, http.MethodGet, "/api/setup/status", "")
	if err := h.GetSetupStatus(c); err != nil {
		t.Fatalf("GetSetupStatus: %v", err)
	}
	got := decodeJSON[setupStatusResp](t, rec.Body.Bytes())
	if !got.Configured {
		t.Error("Configured = false, want true")
	}
}

func TestGetSetupStatus_ConfiguredViaDB(t *testing.T) {
	h, queries, _ := newSetupTestSetup(t, &config.Config{HueBridgeHost: "192.168.1.10"}, nil)
	if err := queries.SetSetting(context.Background(), db.SetSettingParams{Key: db.HueAppKeyKey, Value: "db-key"}); err != nil {
		t.Fatalf("SetSetting: %v", err)
	}

	rec, c := newCtx(t, http.MethodGet, "/api/setup/status", "")
	if err := h.GetSetupStatus(c); err != nil {
		t.Fatalf("GetSetupStatus: %v", err)
	}
	got := decodeJSON[setupStatusResp](t, rec.Body.Bytes())
	if !got.Configured {
		t.Error("Configured = false, want true")
	}
}

func TestPostSetupPair_Success(t *testing.T) {
	pair := func(ctx context.Context, bridgeAddr string) (string, error) { return "new-key", nil }
	h, queries, stopped := newSetupTestSetup(t, &config.Config{HueBridgeHost: "192.168.1.10"}, pair)

	rec, c := newCtx(t, http.MethodPost, "/api/setup/pair", "")
	if err := h.PostSetupPair(c); err != nil {
		t.Fatalf("PostSetupPair: %v", err)
	}
	got := decodeJSON[pairResp](t, rec.Body.Bytes())
	if !got.Paired {
		t.Error("Paired = false, want true")
	}
	stored, err := queries.GetSetting(context.Background(), db.HueAppKeyKey)
	if err != nil {
		t.Fatalf("GetSetting: %v", err)
	}
	if stored != "new-key" {
		t.Errorf("stored key = %q, want %q", stored, "new-key")
	}
	if !*stopped {
		t.Error("stop was not called after a successful pairing")
	}
}

func TestPostSetupPair_WaitingForButton(t *testing.T) {
	pair := func(ctx context.Context, bridgeAddr string) (string, error) { return "", hue.ErrLinkButtonNotPressed }
	h, queries, stopped := newSetupTestSetup(t, &config.Config{HueBridgeHost: "192.168.1.10"}, pair)

	rec, c := newCtx(t, http.MethodPost, "/api/setup/pair", "")
	if err := h.PostSetupPair(c); err != nil {
		t.Fatalf("PostSetupPair: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusOK)
	}
	got := decodeJSON[pairResp](t, rec.Body.Bytes())
	if got.Paired {
		t.Error("Paired = true, want false")
	}
	if got.Reason != "waiting_for_button" {
		t.Errorf("Reason = %q, want %q", got.Reason, "waiting_for_button")
	}
	if *stopped {
		t.Error("stop must not be called while waiting for the button")
	}
	if _, err := queries.GetSetting(context.Background(), db.HueAppKeyKey); err == nil {
		t.Error("no key should have been stored")
	}
}

func TestPostSetupPair_HardError(t *testing.T) {
	pair := func(ctx context.Context, bridgeAddr string) (string, error) { return "", errBridgeUnreachable }
	h, _, stopped := newSetupTestSetup(t, &config.Config{HueBridgeHost: "192.168.1.10"}, pair)

	rec, c := newCtx(t, http.MethodPost, "/api/setup/pair", "")
	if err := h.PostSetupPair(c); err != nil {
		t.Fatalf("PostSetupPair: %v", err)
	}
	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadGateway)
	}
	if *stopped {
		t.Error("stop must not be called on a hard pairing error")
	}
}

func TestPostSetupPair_AlreadyConfiguredViaEnv(t *testing.T) {
	called := false
	pair := func(ctx context.Context, bridgeAddr string) (string, error) {
		called = true
		return "new-key", nil
	}
	h, _, stopped := newSetupTestSetup(t, &config.Config{HueBridgeHost: "192.168.1.10", HueAppKey: "env-key"}, pair)

	rec, c := newCtx(t, http.MethodPost, "/api/setup/pair", "")
	if err := h.PostSetupPair(c); err != nil {
		t.Fatalf("PostSetupPair: %v", err)
	}
	if rec.Code != http.StatusConflict {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusConflict)
	}
	if called {
		t.Error("pair function must not be invoked when already configured via env")
	}
	if *stopped {
		t.Error("stop must not be called")
	}
}
