package handler_test

import (
	"context"
	"encoding/json"
	"io"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/florentsorel/watchue/internal/config"
	"github.com/florentsorel/watchue/internal/db"
	"github.com/florentsorel/watchue/internal/handler"
	"github.com/florentsorel/watchue/internal/hue"
	"github.com/florentsorel/watchue/internal/stream"
	"github.com/labstack/echo/v5"
)

// newTestSetup wraps a mock Hue client and a real in-memory sqlite db in a
// handler.Handler for tests, returning the queries too so tests can seed/
// assert on DB state directly.
func newTestSetup(t *testing.T, hueClient handler.HueClient) (*handler.Handler, *db.Queries) {
	t.Helper()
	return newTestSetupWithConfig(t, hueClient, &config.Config{})
}

func newTestSetupWithConfig(t *testing.T, hueClient handler.HueClient, cfg *config.Config) (*handler.Handler, *db.Queries) {
	t.Helper()
	return newTestSetupWithHub(t, hueClient, cfg, stream.NewHub())
}

func newTestSetupWithHub(t *testing.T, hueClient handler.HueClient, cfg *config.Config, hub *stream.Hub) (*handler.Handler, *db.Queries) {
	t.Helper()
	bridgeOnline := &atomic.Bool{}
	bridgeOnline.Store(true)
	return newTestSetupWithBridgeOnline(t, hueClient, cfg, hub, bridgeOnline)
}

func newTestSetupWithBridgeOnline(t *testing.T, hueClient handler.HueClient, cfg *config.Config, hub *stream.Hub, bridgeOnline *atomic.Bool) (*handler.Handler, *db.Queries) {
	t.Helper()
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("db.Open: %v", err)
	}
	t.Cleanup(func() { conn.Close() })
	queries := db.New(conn)
	noopPair := func(ctx context.Context, bridgeAddr string) (string, error) { return "", nil }
	noopFactory := func(cfg handler.NotifyConfig) (handler.Notifier, error) { return &mockNotifier{}, nil }
	return handler.New(hueClient, queries, cfg, hub, bridgeOnline, "test", func() {}, noopPair, noopFactory, handler.NewNotifierStore()), queries
}

// newCtx creates an Echo context backed by a response recorder.
func newCtx(t *testing.T, method, path, body string) (*httptest.ResponseRecorder, *echo.Context) {
	t.Helper()
	e := echo.New()
	var r io.Reader
	if body != "" {
		r = strings.NewReader(body)
	}
	req := httptest.NewRequest(method, path, r)
	if body != "" {
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	}
	rec := httptest.NewRecorder()
	return rec, e.NewContext(req, rec)
}

func decodeJSON[T any](t *testing.T, body []byte) T {
	t.Helper()
	var v T
	if err := json.Unmarshal(body, &v); err != nil {
		t.Fatalf("decodeJSON: %v\nbody: %s", err, body)
	}
	return v
}

// mockHue is a configurable handler.HueClient for tests.
type mockHue struct {
	lightsFunc        func(ctx context.Context) ([]hue.Light, error)
	zonesFunc         func(ctx context.Context) ([]hue.Zone, error)
	roomsFunc         func(ctx context.Context) ([]hue.Room, error)
	groupedLightsFunc func(ctx context.Context) ([]hue.GroupedLight, error)
}

func (m *mockHue) Lights(ctx context.Context) ([]hue.Light, error) {
	if m.lightsFunc != nil {
		return m.lightsFunc(ctx)
	}
	return nil, nil
}
func (m *mockHue) Zones(ctx context.Context) ([]hue.Zone, error) {
	if m.zonesFunc != nil {
		return m.zonesFunc(ctx)
	}
	return nil, nil
}
func (m *mockHue) Rooms(ctx context.Context) ([]hue.Room, error) {
	if m.roomsFunc != nil {
		return m.roomsFunc(ctx)
	}
	return nil, nil
}
func (m *mockHue) GroupedLights(ctx context.Context) ([]hue.GroupedLight, error) {
	if m.groupedLightsFunc != nil {
		return m.groupedLightsFunc(ctx)
	}
	return nil, nil
}

var _ handler.HueClient = (*mockHue)(nil)

// mockNotifier is a configurable handler.Notifier for tests.
type mockNotifier struct {
	sendErr     error
	sendTestErr error
	sendCalls   int
	testCalls   int
}

func (m *mockNotifier) Send(ctx context.Context, resourceName string, on bool) error {
	m.sendCalls++
	return m.sendErr
}

func (m *mockNotifier) SendTest(ctx context.Context) error {
	m.testCalls++
	return m.sendTestErr
}

var _ handler.Notifier = (*mockNotifier)(nil)
