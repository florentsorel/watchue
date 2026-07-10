package handler_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/florentsorel/watchue/internal/hue"
	"github.com/labstack/echo/v5"
)

var errBridgeUnreachable = errors.New("bridge unreachable")

type watchedResp struct {
	ID     string `json:"id"`
	Type   string `json:"type"`
	Name   string `json:"name"`
	Notify bool   `json:"notify"`
}

func TestWatchedResourcesCRUD(t *testing.T) {
	zoneName := "Salon"
	mock := &mockHue{
		zonesFunc: func(ctx context.Context) ([]hue.Zone, error) {
			return []hue.Zone{{ID: "zone-1", Metadata: hue.Metadata{Name: zoneName}}}, nil
		},
	}
	h, _ := newTestSetup(t, mock)

	rec, c := newCtx(t, http.MethodGet, "/api/watched", "")
	if err := h.GetWatched(c); err != nil {
		t.Fatalf("GetWatched: %v", err)
	}
	if got := decodeJSON[[]watchedResp](t, rec.Body.Bytes()); len(got) != 0 {
		t.Fatalf("got %d watched resources, want 0", len(got))
	}

	// no body needed: type and name are resolved from the bridge (the mock)
	rec, c = newCtx(t, http.MethodPut, "/api/watched/zone-1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "zone-1"}})
	if err := h.PutWatched(c); err != nil {
		t.Fatalf("PutWatched: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	rec, c = newCtx(t, http.MethodGet, "/api/watched", "")
	if err := h.GetWatched(c); err != nil {
		t.Fatalf("GetWatched: %v", err)
	}
	got := decodeJSON[[]watchedResp](t, rec.Body.Bytes())
	if len(got) != 1 || got[0].ID != "zone-1" || got[0].Type != "zone" || got[0].Name != "Salon" || !got[0].Notify {
		t.Fatalf("unexpected watched resources: %+v", got)
	}

	// the bridge-side name changes; re-watching the same id refreshes the cache
	zoneName = "Salon renamed"
	_, c = newCtx(t, http.MethodPut, "/api/watched/zone-1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "zone-1"}})
	if err := h.PutWatched(c); err != nil {
		t.Fatalf("PutWatched (refresh): %v", err)
	}
	rec, c = newCtx(t, http.MethodGet, "/api/watched", "")
	if err := h.GetWatched(c); err != nil {
		t.Fatalf("GetWatched: %v", err)
	}
	got = decodeJSON[[]watchedResp](t, rec.Body.Bytes())
	if len(got) != 1 || got[0].Name != "Salon renamed" {
		t.Fatalf("unexpected watched resources after refresh: %+v", got)
	}

	rec, c = newCtx(t, http.MethodDelete, "/api/watched/zone-1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "zone-1"}})
	if err := h.DeleteWatched(c); err != nil {
		t.Fatalf("DeleteWatched: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	rec, c = newCtx(t, http.MethodGet, "/api/watched", "")
	if err := h.GetWatched(c); err != nil {
		t.Fatalf("GetWatched: %v", err)
	}
	if got := decodeJSON[[]watchedResp](t, rec.Body.Bytes()); len(got) != 0 {
		t.Fatalf("got %d watched resources after delete, want 0", len(got))
	}
}

func TestPutWatched_UnknownResource(t *testing.T) {
	mock := &mockHue{
		zonesFunc: func(ctx context.Context) ([]hue.Zone, error) {
			return []hue.Zone{{ID: "zone-1", Metadata: hue.Metadata{Name: "Salon"}}}, nil
		},
	}
	h, _ := newTestSetup(t, mock)

	rec, c := newCtx(t, http.MethodPut, "/api/watched/zone-does-not-exist", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "zone-does-not-exist"}})
	if err := h.PutWatched(c); err != nil {
		t.Fatalf("PutWatched: %v", err)
	}
	if rec.Code != http.StatusUnprocessableEntity {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnprocessableEntity)
	}
}

func TestPutWatched_MissingID(t *testing.T) {
	h, _ := newTestSetup(t, &mockHue{})

	rec, c := newCtx(t, http.MethodPut, "/api/watched/", "")
	if err := h.PutWatched(c); err != nil {
		t.Fatalf("PutWatched: %v", err)
	}
	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
}

func TestPatchWatched_TogglesNotify(t *testing.T) {
	mock := &mockHue{
		zonesFunc: func(ctx context.Context) ([]hue.Zone, error) {
			return []hue.Zone{{ID: "zone-1", Metadata: hue.Metadata{Name: "Salon"}}}, nil
		},
	}
	h, _ := newTestSetup(t, mock)

	_, c := newCtx(t, http.MethodPut, "/api/watched/zone-1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "zone-1"}})
	if err := h.PutWatched(c); err != nil {
		t.Fatalf("PutWatched: %v", err)
	}

	rec, c := newCtx(t, http.MethodPatch, "/api/watched/zone-1", `{"notify":false}`)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "zone-1"}})
	if err := h.PatchWatched(c); err != nil {
		t.Fatalf("PatchWatched: %v", err)
	}
	if rec.Code != http.StatusNoContent {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}

	rec, c = newCtx(t, http.MethodGet, "/api/watched", "")
	if err := h.GetWatched(c); err != nil {
		t.Fatalf("GetWatched: %v", err)
	}
	got := decodeJSON[[]watchedResp](t, rec.Body.Bytes())
	if len(got) != 1 || got[0].Notify {
		t.Fatalf("Notify = %v, want false after muting", got)
	}

	_, c = newCtx(t, http.MethodPatch, "/api/watched/zone-1", `{"notify":true}`)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "zone-1"}})
	if err := h.PatchWatched(c); err != nil {
		t.Fatalf("PatchWatched (unmute): %v", err)
	}
	rec, c = newCtx(t, http.MethodGet, "/api/watched", "")
	if err := h.GetWatched(c); err != nil {
		t.Fatalf("GetWatched: %v", err)
	}
	got = decodeJSON[[]watchedResp](t, rec.Body.Bytes())
	if len(got) != 1 || !got[0].Notify {
		t.Fatalf("Notify = %v, want true after unmuting", got)
	}
}

func TestPatchWatched_UnknownResource(t *testing.T) {
	h, _ := newTestSetup(t, &mockHue{})

	rec, c := newCtx(t, http.MethodPatch, "/api/watched/zone-does-not-exist", `{"notify":false}`)
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "zone-does-not-exist"}})
	if err := h.PatchWatched(c); err != nil {
		t.Fatalf("PatchWatched: %v", err)
	}
	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestPatchWatched_InvalidRequest(t *testing.T) {
	h, _ := newTestSetup(t, &mockHue{})

	t.Run("missing id", func(t *testing.T) {
		rec, c := newCtx(t, http.MethodPatch, "/api/watched/", `{"notify":false}`)
		if err := h.PatchWatched(c); err != nil {
			t.Fatalf("PatchWatched: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("missing notify field", func(t *testing.T) {
		rec, c := newCtx(t, http.MethodPatch, "/api/watched/zone-1", `{}`)
		c.SetPathValues(echo.PathValues{{Name: "id", Value: "zone-1"}})
		if err := h.PatchWatched(c); err != nil {
			t.Fatalf("PatchWatched: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("invalid body", func(t *testing.T) {
		rec, c := newCtx(t, http.MethodPatch, "/api/watched/zone-1", `not json`)
		c.SetPathValues(echo.PathValues{{Name: "id", Value: "zone-1"}})
		if err := h.PatchWatched(c); err != nil {
			t.Fatalf("PatchWatched: %v", err)
		}
		if rec.Code != http.StatusBadRequest {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
		}
	})
}

func TestPutWatched_BridgeUnreachable(t *testing.T) {
	mock := &mockHue{
		zonesFunc: func(ctx context.Context) ([]hue.Zone, error) {
			return nil, errBridgeUnreachable
		},
	}
	h, _ := newTestSetup(t, mock)

	rec, c := newCtx(t, http.MethodPut, "/api/watched/zone-1", "")
	c.SetPathValues(echo.PathValues{{Name: "id", Value: "zone-1"}})
	if err := h.PutWatched(c); err != nil {
		t.Fatalf("PutWatched: %v", err)
	}
	if rec.Code != http.StatusBadGateway {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadGateway)
	}
}
