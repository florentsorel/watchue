// Unlike internal/telegram, no white-box test seam is needed here: a
// Discord webhook URL is the entire endpoint (no separate base+token to
// stitch together), so tests just point NewClient straight at an
// httptest.Server's URL.
package discord_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/florentsorel/watchue/internal/discord"
)

type executeWebhookRequest struct {
	Content         string `json:"content"`
	AllowedMentions struct {
		Parse []string `json:"parse"`
	} `json:"allowed_mentions"`
}

func TestSend_Success(t *testing.T) {
	var gotBody executeWebhookRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	client := discord.NewClient(server.URL, server.Client())

	if err := client.Send(context.Background(), "Salon", true); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if want := "🔆 **Salon** was turned **on**."; gotBody.Content != want {
		t.Errorf("Content = %q, want %q", gotBody.Content, want)
	}
	if gotBody.AllowedMentions.Parse == nil || len(gotBody.AllowedMentions.Parse) != 0 {
		t.Errorf("AllowedMentions.Parse = %v, want an empty (non-nil) list", gotBody.AllowedMentions.Parse)
	}
}

func TestSend_EscapesMarkdown(t *testing.T) {
	var gotBody executeWebhookRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	client := discord.NewClient(server.URL, server.Client())

	if err := client.Send(context.Background(), "*Office* `desk`", false); err != nil {
		t.Fatalf("Send() error = %v", err)
	}

	if want := "🔆 **\\*Office\\* \\`desk\\`** was turned **off**."; gotBody.Content != want {
		t.Errorf("Content = %q, want %q", gotBody.Content, want)
	}
}

func TestSend_ErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{"message": "Invalid Webhook Token", "code": 50027})
	}))
	t.Cleanup(server.Close)
	client := discord.NewClient(server.URL, server.Client())

	err := client.Send(context.Background(), "Salon", true)
	if err == nil {
		t.Fatal("expected an error for a non-2xx response")
	}
	if got := err.Error(); !strings.Contains(got, "Invalid Webhook Token") {
		t.Errorf("error = %q, want it to mention %q", got, "Invalid Webhook Token")
	}
}

func TestSend_NonJSONErrorResponse(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
		w.Write([]byte("<html>bad gateway</html>"))
	}))
	t.Cleanup(server.Close)
	client := discord.NewClient(server.URL, server.Client())

	err := client.Send(context.Background(), "Salon", true)
	if err == nil {
		t.Fatal("expected an error for a non-JSON error body")
	}
	if got := err.Error(); !strings.Contains(got, "502") {
		t.Errorf("error = %q, want it to mention the status code", got)
	}
}

func TestSendTest_Success(t *testing.T) {
	var gotBody executeWebhookRequest
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&gotBody); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	t.Cleanup(server.Close)
	client := discord.NewClient(server.URL, server.Client())

	if err := client.SendTest(context.Background()); err != nil {
		t.Fatalf("SendTest() error = %v", err)
	}
	if !strings.Contains(gotBody.Content, "Watchue") {
		t.Errorf("Content = %q, want it to mention Watchue", gotBody.Content)
	}
}
