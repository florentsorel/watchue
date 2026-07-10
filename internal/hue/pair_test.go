package hue_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/florentsorel/watchue/internal/hue"
)

func pairAddr(server *httptest.Server) string {
	return strings.TrimPrefix(server.URL, "https://")
}

func TestPair_Success(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("hue-application-key") != "" {
			t.Error("pairing request must not send an application key header")
		}
		fmt.Fprint(w, `[{"success":{"username":"new-app-key"}}]`)
	}))
	t.Cleanup(server.Close)

	key, err := hue.Pair(context.Background(), pairAddr(server), server.Client())
	if err != nil {
		t.Fatalf("Pair: %v", err)
	}
	if key != "new-app-key" {
		t.Errorf("key = %q, want %q", key, "new-app-key")
	}
}

func TestPair_LinkButtonNotPressed(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"error":{"type":101,"address":"","description":"link button not pressed"}}]`)
	}))
	t.Cleanup(server.Close)

	_, err := hue.Pair(context.Background(), pairAddr(server), server.Client())
	if !errors.Is(err, hue.ErrLinkButtonNotPressed) {
		t.Fatalf("err = %v, want ErrLinkButtonNotPressed", err)
	}
}

func TestPair_OtherBridgeError(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[{"error":{"type":1,"address":"","description":"unauthorized user"}}]`)
	}))
	t.Cleanup(server.Close)

	_, err := hue.Pair(context.Background(), pairAddr(server), server.Client())
	if err == nil {
		t.Fatal("expected an error")
	}
	if errors.Is(err, hue.ErrLinkButtonNotPressed) {
		t.Error("a non-101 error must not be classified as ErrLinkButtonNotPressed")
	}
}

func TestPair_MalformedResponse(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `not json`)
	}))
	t.Cleanup(server.Close)

	if _, err := hue.Pair(context.Background(), pairAddr(server), server.Client()); err == nil {
		t.Fatal("expected an error for malformed JSON")
	}
}

func TestPair_EmptyResponse(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `[]`)
	}))
	t.Cleanup(server.Close)

	if _, err := hue.Pair(context.Background(), pairAddr(server), server.Client()); err == nil {
		t.Fatal("expected an error for an empty response array")
	}
}

func TestPair_NonOKStatus(t *testing.T) {
	server := httptest.NewTLSServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(server.Close)

	if _, err := hue.Pair(context.Background(), pairAddr(server), server.Client()); err == nil {
		t.Fatal("expected an error for a non-200 status")
	}
}
