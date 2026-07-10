package handler

import (
	"fmt"
	"net/http"

	"github.com/labstack/echo/v5"
)

// GetStream relays real-time resource/event updates over SSE until the
// client disconnects. See internal/stream for the message shapes.
func (h *Handler) GetStream(c *echo.Context) error {
	resp := c.Response()
	resp.Header().Set("Content-Type", "text/event-stream")
	resp.Header().Set("Cache-Control", "no-cache")
	resp.Header().Set("Connection", "keep-alive")
	resp.WriteHeader(http.StatusOK)

	flusher := http.NewResponseController(resp)
	flusher.Flush()

	ch, unsubscribe := h.hub.Subscribe()
	defer unsubscribe()

	ctx := c.Request().Context()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return nil
			}
			if _, err := fmt.Fprintf(resp, "data: %s\n\n", msg); err != nil {
				return nil
			}
			flusher.Flush()
		case <-ctx.Done():
			return nil
		}
	}
}
