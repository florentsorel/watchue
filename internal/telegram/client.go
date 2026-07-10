// Package telegram sends messages through a Telegram bot to a single chat.
package telegram

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

const defaultBaseURL = "https://api.telegram.org"

// Client sends messages via a Telegram bot to one configured chat.
type Client struct {
	botToken   string
	chatID     string
	httpClient *http.Client
	baseURL    string
}

// NewClient builds a Client for the given bot token and chat id (both
// obtained once from @BotFather / the target chat).
func NewClient(botToken, chatID string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{}
	}
	return &Client{botToken: botToken, chatID: chatID, httpClient: httpClient, baseURL: defaultBaseURL}
}

type sendMessageRequest struct {
	ChatID    string `json:"chat_id"`
	Text      string `json:"text"`
	ParseMode string `json:"parse_mode"`
}

// apiResponse's ok field is the authoritative success signal, not HTTP status.
type apiResponse struct {
	OK          bool   `json:"ok"`
	Description string `json:"description"`
}

// Send posts text as an HTML-formatted message (see Telegram's HTML message
// style: https://core.telegram.org/bots/api#html-style) to the configured chat.
func (c *Client) Send(ctx context.Context, text string) error {
	body, err := json.Marshal(sendMessageRequest{ChatID: c.chatID, Text: text, ParseMode: "HTML"})
	if err != nil {
		return fmt.Errorf("telegram: encode request: %w", err)
	}

	url := fmt.Sprintf("%s/bot%s/sendMessage", c.baseURL, c.botToken)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(body))
	if err != nil {
		return fmt.Errorf("telegram: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("telegram: send message: %w", err)
	}
	defer resp.Body.Close()

	var apiResp apiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		return fmt.Errorf("telegram: decode response: %w", err)
	}
	if !apiResp.OK {
		return fmt.Errorf("telegram: send message: %s", apiResp.Description)
	}
	return nil
}
