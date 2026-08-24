package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/labstack/echo/v5"
)

const (
	defaultStatsDays = 30
	maxStatsDays     = 3650 // the UI's "All" — a ceiling, not a real range
)

// sqliteTimeLayout matches CURRENT_TIMESTAMP, which events.created_at defaults
// to: UTC, second resolution, no zone suffix.
const sqliteTimeLayout = "2006-01-02 15:04:05"

type sessionResponse struct {
	ResourceID   string  `json:"resource_id"`
	ResourceType string  `json:"resource_type"`
	Name         string  `json:"name"`
	Start        string  `json:"start"`
	End          *string `json:"end"` // null while the resource is still on
}

type statsResponse struct {
	From     string            `json:"from"`
	Days     int64             `json:"days"`
	Sessions []sessionResponse `json:"sessions"`
}

// GetStats lists the "on" periods overlapping the last `days` days, oldest
// first. Deliberately raw intervals rather than pre-binned counts: bucketing by
// day or hour is only correct in the viewer's own timezone (DST included), and
// only the browser knows it — see web/src/utils/insights.ts.
func (h *Handler) GetStats(c *echo.Context) error {
	days := int64(defaultStatsDays)
	if v := c.QueryParam("days"); v != "" {
		if n, err := strconv.ParseInt(v, 10, 64); err == nil && n > 0 && n <= maxStatsDays {
			days = n
		}
	}

	from := time.Now().UTC().AddDate(0, 0, -int(days)).Format(sqliteTimeLayout)
	rows, err := h.db.ListSessionsSince(c.Request().Context(), from)
	if err != nil {
		return jsonInternalError(c, err)
	}

	// An open session is only genuinely ongoing while we are still listening to
	// the resource. Unwatching one leaves its last turn-on with no turn-off ever
	// to follow, which would otherwise read as "on ever since" and swamp every
	// duration on the page — the real turn-off just happened off our radar.
	watched, err := h.db.ListWatchedResources(c.Request().Context())
	if err != nil {
		return jsonInternalError(c, err)
	}
	stillWatched := make(map[string]bool, len(watched))
	for _, w := range watched {
		stillWatched[w.ResourceID] = true
	}

	sessions := make([]sessionResponse, 0, len(rows))
	for _, r := range rows {
		if r.EndAt == "" && !stillWatched[r.ResourceID] {
			continue
		}
		s := sessionResponse{
			ResourceID:   r.ResourceID,
			ResourceType: r.ResourceType,
			Name:         r.Name,
			Start:        r.StartAt,
		}
		if r.EndAt != "" {
			end := r.EndAt
			s.End = &end
		}
		sessions = append(sessions, s)
	}
	return c.JSON(http.StatusOK, statsResponse{From: from, Days: days, Sessions: sessions})
}
