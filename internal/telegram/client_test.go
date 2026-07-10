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

	if err := client.Send(context.Background(), "Salon", false); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if want := "/bottest-bot-token/sendMessage"; gotPath != want {
		t.Errorf("path = %q, want %q", gotPath, want)
	}
	if want := "🔆 <b>Salon</b> was turned <b>off</b>."; gotBody.Text != want {
		t.Errorf("Text = %q, want %q", gotBody.Text, want)
	}
	if gotBody.ChatID != "test-chat-id" || gotBody.ParseMode != "HTML" {
		t.Errorf("unexpected request body: %+v", gotBody)
	}
}

func TestSend_EscapesHTML(t *testing.T) {
	var gotBody sendMessageRequest
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiResponse{OK: true})
	})

	if err := client.Send(context.Background(), "<script>&", true); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if want := "🔆 <b>&lt;script&gt;&amp;</b> was turned <b>on</b>."; gotBody.Text != want {
		t.Errorf("Text = %q, want %q", gotBody.Text, want)
	}
}

func TestSend_APIError(t *testing.T) {
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(apiResponse{OK: false, Description: "chat not found"})
	})

	err := client.Send(context.Background(), "Salon", true)
	if err == nil {
		t.Fatal("expected an error for an ok=false response")
	}
	if got := err.Error(); !strings.Contains(got, "chat not found") {
		t.Errorf("error = %q, want it to mention %q", got, "chat not found")
	}
}

func TestSendTest_Success(t *testing.T) {
	var gotBody sendMessageRequest
	client := newTestClient(t, func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(apiResponse{OK: true})
	})

	if err := client.SendTest(context.Background()); err != nil {
		t.Fatalf("SendTest() error = %v", err)
	}
	if !strings.Contains(gotBody.Text, "Watchue") {
		t.Errorf("Text = %q, want it to mention Watchue", gotBody.Text)
	}
}
