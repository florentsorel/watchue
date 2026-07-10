package handler_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/florentsorel/watchue/internal/db"
)

func TestGetEvents(t *testing.T) {
	h, queries := newTestSetup(t, &mockHue{})
	ctx := context.Background()

	seed := []db.InsertEventParams{
		{ResourceID: "zone-1", ResourceType: "zone", Name: "Salon", OnState: 1, Outcome: "sent"},
		{ResourceID: "zone-1", ResourceType: "zone", Name: "Salon", OnState: 0, Outcome: "muted"},
	}
	for _, e := range seed {
		if _, err := queries.InsertEvent(ctx, e); err != nil {
			t.Fatalf("InsertEvent: %v", err)
		}
	}

	rec, c := newCtx(t, http.MethodGet, "/api/events", "")
	if err := h.GetEvents(c); err != nil {
		t.Fatalf("GetEvents: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	type eventResp struct {
		ID           int64  `json:"id"`
		ResourceID   string `json:"resource_id"`
		ResourceType string `json:"resource_type"`
		Name         string `json:"name"`
		On           bool   `json:"on"`
		Outcome      string `json:"outcome"`
		CreatedAt    string `json:"created_at"`
	}
	events := decodeJSON[[]eventResp](t, rec.Body.Bytes())
	if len(events) != 2 {
		t.Fatalf("got %d events, want 2", len(events))
	}
	// most recent first
	if events[0].On || events[0].Outcome != "muted" {
		t.Errorf("events[0] = %+v, want on=false outcome=muted (most recent change)", events[0])
	}
	if !events[1].On || events[1].Outcome != "sent" {
		t.Errorf("events[1] = %+v, want on=true outcome=sent (oldest change)", events[1])
	}
	if events[0].CreatedAt == "" {
		t.Error("CreatedAt is empty, want a timestamp")
	}
}

func TestGetEvents_LimitParam(t *testing.T) {
	h, queries := newTestSetup(t, &mockHue{})
	ctx := context.Background()

	for i := 0; i < 5; i++ {
		if _, err := queries.InsertEvent(ctx, db.InsertEventParams{
			ResourceID: "light-1", ResourceType: "light", Name: "Lampe", OnState: 1, Outcome: "sent",
		}); err != nil {
			t.Fatalf("InsertEvent: %v", err)
		}
	}

	type eventResp struct {
		ID int64 `json:"id"`
	}

	t.Run("respects a valid limit", func(t *testing.T) {
		rec, c := newCtx(t, http.MethodGet, "/api/events?limit=2", "")
		if err := h.GetEvents(c); err != nil {
			t.Fatalf("GetEvents: %v", err)
		}
		if got := decodeJSON[[]eventResp](t, rec.Body.Bytes()); len(got) != 2 {
			t.Errorf("got %d events, want 2", len(got))
		}
	})

	t.Run("ignores an out-of-range limit and falls back to the default", func(t *testing.T) {
		rec, c := newCtx(t, http.MethodGet, "/api/events?limit=9999", "")
		if err := h.GetEvents(c); err != nil {
			t.Fatalf("GetEvents: %v", err)
		}
		if got := decodeJSON[[]eventResp](t, rec.Body.Bytes()); len(got) != 5 {
			t.Errorf("got %d events, want 5 (all seeded events, under the default limit)", len(got))
		}
	})
}
