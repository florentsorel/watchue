package handler_test

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/florentsorel/watchue/internal/hue"
)

func TestGetZones(t *testing.T) {
	type lightResp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		On   bool   `json:"on"`
	}
	type groupResp struct {
		ID     string      `json:"id"`
		Name   string      `json:"name"`
		On     bool        `json:"on"`
		Lights []lightResp `json:"lights"`
	}

	t.Run("returns zones with nested lights", func(t *testing.T) {
		mock := &mockHue{
			lightsFunc: func(ctx context.Context) ([]hue.Light, error) {
				return []hue.Light{{
					ID:       "light-1",
					Owner:    hue.ResourceIdentifier{RID: "device-1", RType: hue.ResourceDevice},
					Metadata: hue.Metadata{Name: "Lampe salon"},
					On:       hue.OnState{On: true},
				}}, nil
			},
			zonesFunc: func(ctx context.Context) ([]hue.Zone, error) {
				return []hue.Zone{{
					ID:       "zone-1",
					Children: []hue.ResourceIdentifier{{RID: "light-1", RType: hue.ResourceLight}},
					Services: []hue.ResourceIdentifier{{RID: "grouped-1", RType: hue.ResourceGroupedLight}},
					Metadata: hue.Metadata{Name: "Salon"},
				}}, nil
			},
			groupedLightsFunc: func(ctx context.Context) ([]hue.GroupedLight, error) {
				return []hue.GroupedLight{{ID: "grouped-1", On: hue.OnState{On: true}}}, nil
			},
		}

		h, _ := newTestSetup(t, mock)
		rec, c := newCtx(t, http.MethodGet, "/api/zones", "")

		if err := h.GetZones(c); err != nil {
			t.Fatalf("GetZones: %v", err)
		}
		if rec.Code != http.StatusOK {
			t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
		}

		zones := decodeJSON[[]groupResp](t, rec.Body.Bytes())
		if len(zones) != 1 {
			t.Fatalf("got %d zones, want 1", len(zones))
		}
		if zones[0].ID != "zone-1" || zones[0].Name != "Salon" || !zones[0].On {
			t.Errorf("unexpected zone: %+v", zones[0])
		}
		if len(zones[0].Lights) != 1 || zones[0].Lights[0].ID != "light-1" {
			t.Errorf("unexpected zone lights: %+v", zones[0].Lights)
		}
	})

	t.Run("bridge failure returns 502", func(t *testing.T) {
		mock := &mockHue{
			lightsFunc: func(ctx context.Context) ([]hue.Light, error) {
				return nil, errors.New("bridge unreachable")
			},
		}

		h, _ := newTestSetup(t, mock)
		rec, c := newCtx(t, http.MethodGet, "/api/zones", "")

		if err := h.GetZones(c); err != nil {
			t.Fatalf("GetZones: %v", err)
		}
		if rec.Code != http.StatusBadGateway {
			t.Errorf("status = %d, want %d", rec.Code, http.StatusBadGateway)
		}
	})
}

func TestGetRooms(t *testing.T) {
	type lightResp struct {
		ID   string `json:"id"`
		Name string `json:"name"`
		On   bool   `json:"on"`
	}
	type groupResp struct {
		ID     string      `json:"id"`
		Name   string      `json:"name"`
		On     bool        `json:"on"`
		Lights []lightResp `json:"lights"`
	}

	mock := &mockHue{
		lightsFunc: func(ctx context.Context) ([]hue.Light, error) {
			return []hue.Light{{
				ID:       "light-2",
				Owner:    hue.ResourceIdentifier{RID: "device-2", RType: hue.ResourceDevice},
				Metadata: hue.Metadata{Name: "Suspension bureau"},
				On:       hue.OnState{On: false},
			}}, nil
		},
		roomsFunc: func(ctx context.Context) ([]hue.Room, error) {
			return []hue.Room{{
				ID:       "room-1",
				Children: []hue.ResourceIdentifier{{RID: "device-2", RType: hue.ResourceDevice}},
				Services: []hue.ResourceIdentifier{{RID: "grouped-2", RType: hue.ResourceGroupedLight}},
				Metadata: hue.Metadata{Name: "Bureau"},
			}}, nil
		},
		groupedLightsFunc: func(ctx context.Context) ([]hue.GroupedLight, error) {
			return []hue.GroupedLight{{ID: "grouped-2", On: hue.OnState{On: false}}}, nil
		},
	}

	h, _ := newTestSetup(t, mock)
	rec, c := newCtx(t, http.MethodGet, "/api/rooms", "")

	if err := h.GetRooms(c); err != nil {
		t.Fatalf("GetRooms: %v", err)
	}
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", rec.Code, http.StatusOK)
	}

	rooms := decodeJSON[[]groupResp](t, rec.Body.Bytes())
	if len(rooms) != 1 {
		t.Fatalf("got %d rooms, want 1", len(rooms))
	}
	if rooms[0].ID != "room-1" || rooms[0].Name != "Bureau" || rooms[0].On {
		t.Errorf("unexpected room: %+v", rooms[0])
	}
	if len(rooms[0].Lights) != 1 || rooms[0].Lights[0].ID != "light-2" {
		t.Errorf("unexpected room lights: %+v", rooms[0].Lights)
	}
}
