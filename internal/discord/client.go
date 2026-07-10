// Package discord sends messages through a Discord incoming webhook.
package discord

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// Client posts messages to a single configured Discord webhook.
type Client struct {
	webhookURL string
	httpClient *http.Client
}

// NewClient builds a Client for the given incoming webhook URL (created via
// a Discord channel's Integrations settings — no OAuth/bot token needed).
func NewClient(webhookURL string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &Client{webhookURL: webhookURL, httpClient: httpClient}
}

type executeWebhookRequest struct {
	Content         string          `json:"content"`
	AllowedMentions allowedMentions `json:"allowed_mentions"`
}

// allowedMentions with an empty parse list suppresses @everyone/@here/role
// mentions Discord would otherwise parse out of plain content — Hue
// light/room names are user-renameable, so this isn't optional hardening.
type allowedMentions struct {
	Parse []string `json:"parse"`
}

// errorResponse is Discord's JSON error envelope on non-2xx responses. Not
// guaranteed to be present (e.g. a Cloudflare-level 5xx isn't JSON at all).
type errorResponse struct {
	Message    string  `json:"message"`
	Code       int     `json:"code"`
	RetryAfter float64 `json:"retry_after"`
}

var markdownSpecialChars = strings.NewReplacer(
	"\\", "\\\\",
	"*", "\\*",
	"_", "\\_",
	"~", "\\~",
	"`", "\\`",
	"|", "\\|",
)

func escapeMarkdown(s string) string {
	return markdownSpecialChars.Replace(s)
}

// Send notifies that resourceName was turned on or off.
func (c *Client) Send(ctx context.Context, resourceName string, on bool) error {
	state := "off"
	if on {
		state = "on"
	}
	content := fmt.Sprintf("🔆 **%s** was turned **%s**.", escapeMarkdown(resourceName), state)
	return c.post(ctx, content)
}

// SendTest posts a fixed message, used by the /setup notification step to
// confirm the webhook actually works before saving it.
func (c *Client) SendTest(ctx context.Context) error {
	return c.post(ctx, "✅ **Watchue** is connected and will notify you here.")
}

func (c *Client) post(ctx context.Context, content string) error {
	body, err := json.Marshal(executeWebhookRequest{Content: content, AllowedMentions: allowedMentions{Parse: []string{}}})
	if err != nil {
		return fmt.Errorf("discord: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, c.webhookURL, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("discord: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("discord: send message: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		var errResp errorResponse
		if err := json.NewDecoder(resp.Body).Decode(&errResp); err == nil && errResp.Message != "" {
			if errResp.RetryAfter > 0 {
				return fmt.Errorf("discord: send message: %s (retry after %.1fs)", errResp.Message, errResp.RetryAfter)
			}
			return fmt.Errorf("discord: send message: %s", errResp.Message)
		}
		return fmt.Errorf("discord: send message: unexpected status %s", resp.Status)
	}
	return nil
}
