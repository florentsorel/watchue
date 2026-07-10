package hue

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

// ErrLinkButtonNotPressed indicates the bridge's link button hasn't been
// pressed within its ~30s pairing window yet — an expected state while
// polling, not a hard failure.
var ErrLinkButtonNotPressed = errors.New("hue: link button not pressed")

// pairResult mirrors one element of the array the bridge's legacy /api
// pairing endpoint responds with, e.g.
// [{"success":{"username":"..."}}] or [{"error":{"type":101,...}}].
type pairResult struct {
	Success *struct {
		Username string `json:"username"`
	} `json:"success"`
	Error *struct {
		Type        int    `json:"type"`
		Description string `json:"description"`
	} `json:"error"`
}

// linkButtonNotPressedErrorType is the Hue API's error type code for
// "link button not pressed".
const linkButtonNotPressedErrorType = 101

// Pair performs a single attempt at the bridge's unauthenticated link-button
// pairing exchange (POST /api) and returns the resulting application key.
// Unlike Client, it sends no application key header — none exists yet.
//
// Callers should treat ErrLinkButtonNotPressed as "keep polling"; any other
// error means the bridge itself couldn't be reached or returned something
// unexpected.
//
// If httpClient is nil, a client that skips TLS verification is used, same
// as NewClient — Hue bridges serve a self-signed certificate.
func Pair(ctx context.Context, bridgeAddr string, httpClient *http.Client) (string, error) {
	if httpClient == nil {
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		}
	}

	body, err := json.Marshal(map[string]string{"devicetype": "watchue#homelab"})
	if err != nil {
		return "", fmt.Errorf("hue: pair: encode request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, "https://"+bridgeAddr+"/api", bytes.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("hue: pair: build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("hue: pair: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("hue: pair: unexpected status %s", resp.Status)
	}

	var results []pairResult
	if err := json.NewDecoder(resp.Body).Decode(&results); err != nil {
		return "", fmt.Errorf("hue: pair: decode response: %w", err)
	}
	if len(results) == 0 {
		return "", errors.New("hue: pair: empty response")
	}

	result := results[0]
	if result.Success != nil {
		return result.Success.Username, nil
	}
	if result.Error != nil {
		if result.Error.Type == linkButtonNotPressedErrorType {
			return "", ErrLinkButtonNotPressed
		}
		return "", fmt.Errorf("hue: pair: %s", result.Error.Description)
	}
	return "", errors.New("hue: pair: unrecognized response")
}
