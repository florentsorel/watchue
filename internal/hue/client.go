// Package hue wraps the Philips Hue Bridge CLIP v2 API: fetching resources
// (lights, zones, grouped lights) and subscribing to its server-sent event
// stream.
package hue

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// Client talks to a single Hue Bridge on the local network.
type Client struct {
	bridgeAddr string // host or host:port, e.g. "192.168.1.10"
	appKey     string
	httpClient *http.Client
}

// NewClient builds a Client for the bridge at bridgeAddr, authenticated with
// appKey (obtained once via the bridge's link-button pairing flow).
//
// If httpClient is nil, a client that skips TLS verification is used: Hue
// bridges serve a self-signed certificate and are only reachable on the LAN.
// Pass a custom httpClient (e.g. pinning the bridge's certificate) to avoid
// that. Do not set a Timeout on a client you pass in — it would also cut off
// the long-lived event stream in Subscribe.
func NewClient(bridgeAddr, appKey string, httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = &http.Client{
			Transport: &http.Transport{
				TLSClientConfig: &tls.Config{InsecureSkipVerify: true}, //nolint:gosec
			},
		}
	}
	return &Client{bridgeAddr: bridgeAddr, appKey: appKey, httpClient: httpClient}
}

// get performs an authenticated GET against the bridge's CLIP v2 API and
// decodes the JSON response body into out. It applies its own timeout so
// callers don't need to, independent of the long-lived Subscribe stream.
func (c *Client) get(ctx context.Context, path string, out any) error {
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+c.bridgeAddr+path, nil)
	if err != nil {
		return fmt.Errorf("hue: build request: %w", err)
	}
	req.Header.Set("hue-application-key", c.appKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("hue: GET %s: %w", path, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("hue: GET %s: unexpected status %s", path, resp.Status)
	}
	if err := json.NewDecoder(resp.Body).Decode(out); err != nil {
		return fmt.Errorf("hue: decode %s: %w", path, err)
	}
	return nil
}
