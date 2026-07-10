package hue_test

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/florentsorel/watchue/internal/hue"
)

const eventStreamTimeout = 2 * time.Second

func TestClientSubscribeDecodesEvents(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/eventstream/clip/v2", func(w http.ResponseWriter, r *http.Request) {
		flusher := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)

		fmt.Fprint(w, ": keep-alive\n\n") // must be ignored, not decoded as an event
		flusher.Flush()

		fmt.Fprint(w, `data: [{"type":"update","data":[{"id":"light-1","type":"light","on":{"on":true}}]}]`+"\n\n")
		flusher.Flush()

		<-r.Context().Done() // hold the connection open until the client disconnects
	})

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)
	client := hue.NewClient(strings.TrimPrefix(server.URL, "https://"), "test-app-key", server.Client())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events, errs := client.Subscribe(ctx)

	select {
	case ev := <-events:
		if len(ev.Data) != 1 {
			t.Fatalf("got %d event data entries, want 1", len(ev.Data))
		}
		data := ev.Data[0]
		if data.ID != "light-1" || data.On == nil || !data.On.On {
			t.Errorf("unexpected event data: %+v", data)
		}
	case err := <-errs:
		t.Fatalf("unexpected error: %v", err)
	case <-time.After(eventStreamTimeout):
		t.Fatal("timed out waiting for event")
	}

	cancel()

	select {
	case _, ok := <-events:
		if ok {
			t.Error("expected events channel to close after context cancellation")
		}
	case <-time.After(eventStreamTimeout):
		t.Fatal("timed out waiting for events channel to close")
	}
}

func TestClientSubscribeDecodeError(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/eventstream/clip/v2", func(w http.ResponseWriter, r *http.Request) {
		flusher := w.(http.Flusher)
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "data: {not valid json\n\n")
		flusher.Flush()
		<-r.Context().Done()
	})

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)
	client := hue.NewClient(strings.TrimPrefix(server.URL, "https://"), "test-app-key", server.Client())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events, errs := client.Subscribe(ctx)

	select {
	case err := <-errs:
		if err == nil {
			t.Fatal("expected a non-nil decode error")
		}
	case ev := <-events:
		t.Fatalf("expected an error, got event: %+v", ev)
	case <-time.After(eventStreamTimeout):
		t.Fatal("timed out waiting for the decode error")
	}
}

func TestClientSubscribeUnexpectedStatus(t *testing.T) {
	mux := http.NewServeMux()
	mux.HandleFunc("/eventstream/clip/v2", func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	})

	server := httptest.NewTLSServer(mux)
	t.Cleanup(server.Close)
	client := hue.NewClient(strings.TrimPrefix(server.URL, "https://"), "test-app-key", server.Client())

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	events, errs := client.Subscribe(ctx)

	select {
	case err := <-errs:
		if err == nil {
			t.Fatal("expected a non-nil error")
		}
	case ev := <-events:
		t.Fatalf("expected an error, got event: %+v", ev)
	case <-time.After(eventStreamTimeout):
		t.Fatal("timed out waiting for the status error")
	}
}
