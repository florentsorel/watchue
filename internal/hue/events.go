package hue

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
)

// EventType is the kind of change an Event reports.
type EventType string

const (
	EventAdd    EventType = "add"
	EventUpdate EventType = "update"
	EventDelete EventType = "delete"
)

// Event is one entry of a Hue eventstream message. A single SSE "data:" line
// carries a JSON array of these.
type Event struct {
	Type EventType   `json:"type"`
	Data []EventData `json:"data"`
}

// EventData is one changed resource within an Event. Only on/off-relevant
// fields are decoded; the bridge sends more (dimming, color, effects, ...).
type EventData struct {
	ID    string              `json:"id"`
	IDV1  string              `json:"id_v1,omitempty"`
	Type  ResourceType        `json:"type"`
	Owner *ResourceIdentifier `json:"owner,omitempty"`
	On    *OnState            `json:"on,omitempty"`
}

// Subscribe streams decoded events until ctx is cancelled or the connection
// drops, at which point both returned channels close (errs gets at most one
// error first).
func (c *Client) Subscribe(ctx context.Context) (<-chan Event, <-chan error) {
	events := make(chan Event)
	errs := make(chan error, 1)

	go func() {
		defer close(events)
		defer close(errs)

		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://"+c.bridgeAddr+"/eventstream/clip/v2", nil)
		if err != nil {
			errs <- fmt.Errorf("hue: build eventstream request: %w", err)
			return
		}
		req.Header.Set("hue-application-key", c.appKey)
		req.Header.Set("Accept", "text/event-stream")

		resp, err := c.httpClient.Do(req)
		if err != nil {
			errs <- fmt.Errorf("hue: eventstream: %w", err)
			return
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			errs <- fmt.Errorf("hue: eventstream: unexpected status %s", resp.Status)
			return
		}

		scanner := bufio.NewScanner(resp.Body)
		scanner.Buffer(make([]byte, 0, 64*1024), 1024*1024)
		for scanner.Scan() {
			data, ok := strings.CutPrefix(scanner.Text(), "data: ")
			if !ok {
				continue // blank lines and ":" keep-alive comments
			}

			var batch []Event
			if err := json.Unmarshal([]byte(data), &batch); err != nil {
				errs <- fmt.Errorf("hue: decode event: %w", err)
				continue
			}

			for _, ev := range batch {
				select {
				case events <- ev:
				case <-ctx.Done():
					return
				}
			}
		}
		if err := scanner.Err(); err != nil && ctx.Err() == nil {
			errs <- fmt.Errorf("hue: eventstream read: %w", err)
		}
	}()

	return events, errs
}
