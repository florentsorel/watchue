package handler_test

import (
	"context"
	"database/sql"
	"net/http"
	"sync/atomic"
	"testing"

	"github.com/florentsorel/watchue/internal/config"
	"github.com/florentsorel/watchue/internal/db"
	"github.com/florentsorel/watchue/internal/handler"
	"github.com/florentsorel/watchue/internal/stream"
)

// newStatsSetup is newTestSetup plus the raw connection, needed to backdate
// seeded events into distinct sessions.
func newStatsSetup(t *testing.T) (*handler.Handler, *db.Queries, *sql.DB) {
	t.Helper()
	online := &atomic.Bool{}
	online.Store(true)
	return newTestSetupWithConn(t, &mockHue{}, &config.Config{}, stream.NewHub(), online)
}

type sessionResp struct {
	ResourceID   string  `json:"resource_id"`
	ResourceType string  `json:"resource_type"`
	Name         string  `json:"name"`
	Start        string  `json:"start"`
	End          *string `json:"end"`
}

type statsResp struct {
	From     string        `json:"from"`
	Days     int64         `json:"days"`
	Sessions []sessionResp `json:"sessions"`
}

// seedEvent inserts a change with an explicit timestamp — InsertEvent leaves
// created_at to CURRENT_TIMESTAMP, which would put every seeded row in the same
// second and make session pairing untestable.
func seedEvent(t *testing.T, queries *db.Queries, conn *sql.DB, resourceID, name string, on int64, at string) {
	t.Helper()
	row, err := queries.InsertEvent(context.Background(), db.InsertEventParams{
		ResourceID: resourceID, ResourceType: "room", Name: name, OnState: on, Outcome: "sent",
	})
	if err != nil {
		t.Fatalf("InsertEvent: %v", err)
	}
	if _, err := conn.ExecContext(context.Background(),
		"UPDATE events SET created_at = ? WHERE id = ?", at, row.ID); err != nil {
		t.Fatalf("backdate event: %v", err)
	}
}

// watch marks a resource as still being listened to, which is what keeps an
// open session ongoing rather than dropped.
func watch(t *testing.T, queries *db.Queries, resourceID, name string) {
	t.Helper()
	if err := queries.UpsertWatchedResource(context.Background(), db.UpsertWatchedResourceParams{
		ResourceID: resourceID, ResourceType: "room", Name: name,
	}); err != nil {
		t.Fatalf("UpsertWatchedResource: %v", err)
	}
}

func TestGetStats(t *testing.T) {
	h, queries, conn := newStatsSetup(t)

	// Two closed sessions for one room, plus one still running.
	seedEvent(t, queries, conn, "room-1", "Salon", 1, "2026-08-20 18:00:00")
	seedEvent(t, queries, conn, "room-1", "Salon", 0, "2026-08-20 22:30:00")
	seedEvent(t, queries, conn, "room-1", "Salon", 1, "2026-08-21 19:00:00")
	seedEvent(t, queries, conn, "room-1", "Salon", 0, "2026-08-21 20:00:00")
	seedEvent(t, queries, conn, "room-1", "Salon", 1, "2026-08-22 19:00:00")
	// A different resource, so pairing must not cross resources.
	seedEvent(t, queries, conn, "room-2", "Bureau", 1, "2026-08-20 19:00:00")
	seedEvent(t, queries, conn, "room-2", "Bureau", 0, "2026-08-23 09:00:00")

	watch(t, queries, "room-1", "Salon")

	rec, c := newCtx(t, http.MethodGet, "/api/stats?days=3650", "")
	if err := h.GetStats(c); err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	got := decodeJSON[statsResp](t, rec.Body.Bytes())
	if got.Days != 3650 {
		t.Errorf("days = %d, want 3650", got.Days)
	}
	if len(got.Sessions) != 4 {
		t.Fatalf("got %d sessions, want 4 (three for room-1, one for room-2)", len(got.Sessions))
	}

	// Oldest first.
	if got.Sessions[0].Start != "2026-08-20 18:00:00" {
		t.Errorf("sessions[0].start = %q, want the oldest turn-on", got.Sessions[0].Start)
	}
	if got.Sessions[0].End == nil || *got.Sessions[0].End != "2026-08-20 22:30:00" {
		t.Errorf("sessions[0].end = %v, want 2026-08-20 22:30:00", got.Sessions[0].End)
	}
	// room-2's turn-off is later than room-1's, so pairing by resource matters.
	var bureau *sessionResp
	for i := range got.Sessions {
		if got.Sessions[i].ResourceID == "room-2" {
			bureau = &got.Sessions[i]
		}
	}
	if bureau == nil {
		t.Fatal("no session for room-2")
	}
	if bureau.End == nil || *bureau.End != "2026-08-23 09:00:00" {
		t.Errorf("room-2 end = %v, want 2026-08-23 09:00:00", bureau.End)
	}
	if bureau.Name != "Bureau" {
		t.Errorf("room-2 name = %q, want Bureau", bureau.Name)
	}

	// The still-running session reports a null end rather than an empty string.
	last := got.Sessions[len(got.Sessions)-1]
	if last.Start != "2026-08-22 19:00:00" {
		t.Fatalf("last session start = %q, want the still-running one", last.Start)
	}
	if last.End != nil {
		t.Errorf("last session end = %q, want null (still on)", *last.End)
	}
}

func TestGetStats_WindowKeepsOverlappingSessions(t *testing.T) {
	h, queries, conn := newStatsSetup(t)

	// Ended long before any window we ask for.
	seedEvent(t, queries, conn, "room-1", "Salon", 1, "2020-01-01 10:00:00")
	seedEvent(t, queries, conn, "room-1", "Salon", 0, "2020-01-01 12:00:00")
	// Started long ago and still on: must survive any window, since it is
	// contributing "on" time right now.
	seedEvent(t, queries, conn, "room-2", "Bureau", 1, "2020-06-01 10:00:00")

	watch(t, queries, "room-2", "Bureau")

	rec, c := newCtx(t, http.MethodGet, "/api/stats?days=7", "")
	if err := h.GetStats(c); err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	got := decodeJSON[statsResp](t, rec.Body.Bytes())
	if len(got.Sessions) != 1 {
		t.Fatalf("got %d sessions, want 1 (only the still-running one)", len(got.Sessions))
	}
	if got.Sessions[0].ResourceID != "room-2" || got.Sessions[0].End != nil {
		t.Errorf("sessions[0] = %+v, want the still-running room-2 session", got.Sessions[0])
	}
}

func TestGetStats_DaysParam(t *testing.T) {
	h, _, _ := newStatsSetup(t)

	for _, tc := range []struct {
		name  string
		query string
		want  int64
	}{
		{"defaults when absent", "/api/stats", 30},
		{"respects a valid value", "/api/stats?days=7", 7},
		{"falls back when out of range", "/api/stats?days=99999", 30},
		{"falls back when not a number", "/api/stats?days=week", 30},
		{"falls back when zero", "/api/stats?days=0", 30},
	} {
		t.Run(tc.name, func(t *testing.T) {
			rec, c := newCtx(t, http.MethodGet, tc.query, "")
			if err := h.GetStats(c); err != nil {
				t.Fatalf("GetStats: %v", err)
			}
			if got := decodeJSON[statsResp](t, rec.Body.Bytes()); got.Days != tc.want {
				t.Errorf("days = %d, want %d", got.Days, tc.want)
			}
		})
	}
}

func TestGetStats_DropsOpenSessionOfUnwatchedResource(t *testing.T) {
	h, queries, conn := newStatsSetup(t)

	// Watched, turned on, never turned off, then unwatched: the turn-off
	// happened after we stopped listening, so the duration is unknowable.
	seedEvent(t, queries, conn, "room-1", "Salon", 1, "2026-08-20 18:00:00")
	// Still watched and still on — must be kept.
	seedEvent(t, queries, conn, "room-2", "Bureau", 1, "2026-08-20 19:00:00")
	watch(t, queries, "room-2", "Bureau")

	rec, c := newCtx(t, http.MethodGet, "/api/stats?days=3650", "")
	if err := h.GetStats(c); err != nil {
		t.Fatalf("GetStats: %v", err)
	}
	got := decodeJSON[statsResp](t, rec.Body.Bytes())
	if len(got.Sessions) != 1 {
		t.Fatalf("got %d sessions, want 1 (the unwatched resource's dangling turn-on is dropped)", len(got.Sessions))
	}
	if got.Sessions[0].ResourceID != "room-2" {
		t.Errorf("sessions[0].resource_id = %q, want room-2", got.Sessions[0].ResourceID)
	}
}
