package handler

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v5"
)

const (
	defaultEventsLimit = 50
	maxEventsLimit     = 200
)

type eventResponse struct {
	ID           int64  `json:"id"`
	ResourceID   string `json:"resource_id"`
	ResourceType string `json:"resource_type"`
	Name         string `json:"name"`
	On           bool   `json:"on"`
	Outcome      string `json:"outcome"` // "sent", "muted", or "channel_off"
	CreatedAt    string `json:"created_at"`
}

// GetEvents lists recent changes, most recent first, defaulting to 50 and capped at 200.
func (h *Handler) GetEvents(c *echo.Context) error {
	limit := int64(defaultEventsLimit)
	if v := c.QueryParam("limit"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 && n <= maxEventsLimit {
			limit = n
		}
	}

	rows, err := h.db.ListEvents(c.Request().Context(), limit)
	if err != nil {
		return jsonInternalError(c, err)
	}

	resp := make([]eventResponse, 0, len(rows))
	for _, r := range rows {
		resp = append(resp, eventResponse{
			ID:           r.ID,
			ResourceID:   r.ResourceID,
			ResourceType: r.ResourceType,
			Name:         r.Name,
			On:           r.OnState != 0,
			Outcome:      r.Outcome,
			CreatedAt:    r.CreatedAt,
		})
	}
	return c.JSON(http.StatusOK, resp)
}
