package telegram

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

// newTestClient points a Client at an httptest server instead of the real
// Telegram API. White-box (package telegram, not telegram_test) so it can
// set the unexported baseURL directly.
func newTestClient(t *testing.T, handler http.HandlerFunc) *Client {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	return &Client{
		botToken:   "test-bot-token",
		chatID:     "test-chat-id",
		httpClient: server.Client(),
		baseURL:    server.URL,
	}
}

func TestSend_Success(t *testing.T) {
	var gotPath string
	var gotBody sendMessageRequest

	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiResponse{OK: true})
	})

	if err := client.Send(context.Background(), "Salon is now off"); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if want := "/bottest-bot-token/sendMessage"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if gotBody.ChatID != "test-chat-id" || gotBody.Text != "Salon is now off" || gotBody.ParseMode != "HTML" {
		t.Errorf("unexpected request body: %+v", gotBody)
	}
}

func TestSend_APIError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiResponse{OK: false, Description: "chat not found"})
	})

	err := client.Send(context.Background(), "hello")
	if err == nil {
		t.Fatal("expected an error for an ok=false response")
	}
	if got := err.Error(); !strings.Contains(got, "chat not found") {
		t.Errorf("error = %q, want it to mention %q", got, "chat not found")
	}
}
