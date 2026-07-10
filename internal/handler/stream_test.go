package handler_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/florentsorel/watchue/internal/config"
	"github.com/florentsorel/watchue/internal/stream"
	"github.com/labstack/echo/v5"
)

func TestGetStream_RelaysPublishedMessages(t *testing.T) {
	hub := stream.NewHub()
	h, _ := newTestSetupWithHub(t, &mockHue{}, &config.Config{}, hub)

	ctx, cancel := context.WithCancel(context.Background())
	req := httptest.NewRequest(http.MethodGet, "/api/stream", nil).WithContext(ctx)
	rec := httptest.NewRecorder()
	c := echo.New().NewContext(req, rec)

	done := make(chan error, 1)
	go func() { done <- h.GetStream(c) }()

	// let the handler subscribe before publishing
	time.Sleep(50 * time.Millisecond)
	hub.Publish([]byte(`{"kind":"resource","id":"zone-1","on":true}`))
	time.Sleep(50 * time.Millisecond)

	cancel()

	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("GetStream did not return after its context was cancelled")
	}

	body := rec.Body.String()
	if !strings.Contains(body, `data: {"kind":"resource","id":"zone-1","on":true}`) {
		t.Errorf("body = %q, want it to contain the published SSE message", body)
	}
	if ct := rec.Header().Get("Content-Type"); ct != "text/event-stream" {
		t.Errorf("Content-Type = %q, want text/event-stream", ct)
	}
}
