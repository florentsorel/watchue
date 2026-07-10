package catalog_test

import (
	"context"
	"errors"
	"testing"

	"github.com/florentsorel/watchue/internal/catalog"
	"github.com/florentsorel/watchue/internal/hue"
)

// mockHue is a configurable catalog.HueClient for tests.
type mockHue struct {
	lightsFunc        func(ctx context.Context) ([]hue.Light, error)
	zonesFunc         func(ctx context.Context) ([]hue.Zone, error)
	roomsFunc         func(ctx context.Context) ([]hue.Room, error)
	groupedLightsFunc func(ctx context.Context) ([]hue.GroupedLight, error)
}

func (m *mockHue) Lights(ctx context.Context) ([]hue.Light, error) { return m.lightsFunc(ctx) }
func (m *mockHue) Zones(ctx context.Context) ([]hue.Zone, error)   { return m.zonesFunc(ctx) }
func (m *mockHue) Rooms(ctx context.Context) ([]hue.Room, error)   { return m.roomsFunc(ctx) }
func (m *mockHue) GroupedLights(ctx context.Context) ([]hue.GroupedLight, error) {
	return m.groupedLightsFunc(ctx)
}

var _ catalog.HueClient = (*mockHue)(nil)

func TestBuild(t *testing.T) {
	client := &mockHue{
		lightsFunc: func(ctx context.Context) ([]hue.Light, error) {
			return []hue.Light{
				{
					ID:       "light-1",
					Owner:    hue.ResourceIdentifier{RID: "device-1", RType: hue.ResourceDevice},
					Metadata: hue.Metadata{Name: "Lampe salon", Archetype: "sultan_bulb"},
					On:       hue.OnState{On: true},
				},
				{
					ID:       "light-2",
					Owner:    hue.ResourceIdentifier{RID: "device-2", RType: hue.ResourceDevice},
					Metadata: hue.Metadata{Name: "Suspension bureau"},
					On:       hue.OnState{On: false},
				},
			}, nil
		},
		zonesFunc: func(ctx context.Context) ([]hue.Zone, error) {
			return []hue.Zone{
				{
					ID:       "zone-1",
					Children: []hue.ResourceIdentifier{{RID: "light-1", RType: hue.ResourceLight}},
					Services: []hue.ResourceIdentifier{{RID: "grouped-1", RType: hue.ResourceGroupedLight}},
					Metadata: hue.Metadata{Name: "Salon"},
				},
			}, nil
		},
		roomsFunc: func(ctx context.Context) ([]hue.Room, error) {
			return []hue.Room{
				{
					ID:       "room-1",
					Children: []hue.ResourceIdentifier{{RID: "device-2", RType: hue.ResourceDevice}},
					Services: []hue.ResourceIdentifier{{RID: "grouped-2", RType: hue.ResourceGroupedLight}},
					Metadata: hue.Metadata{Name: "Bureau"},
				},
			}, nil
		},
		groupedLightsFunc: func(ctx context.Context) ([]hue.GroupedLight, error) {
			return []hue.GroupedLight{
				{ID: "grouped-1", On: hue.OnState{On: true}},
				{ID: "grouped-2", On: hue.OnState{On: false}},
			}, nil
		},
	}

	cat, err := catalog.Build(context.Background(), client)
	if err != nil {
		t.Fatalf("Build() error = %v", err)
	}

	if len(cat.Zones) != 1 {
		t.Fatalf("got %d zones, want 1", len(cat.Zones))
	}
	zone := cat.Zones[0]
	if zone.ID != "zone-1" || zone.Name != "Salon" || !zone.On {
		t.Errorf("unexpected zone: %+v", zone)
	}
	if len(zone.Lights) != 1 || zone.Lights[0].ID != "light-1" || zone.Lights[0].Name != "Lampe salon" || !zone.Lights[0].On {
		t.Errorf("unexpected zone lights: %+v", zone.Lights)
	}
	if zone.Lights[0].Archetype != "sultan_bulb" {
		t.Errorf("Archetype = %q, want %q", zone.Lights[0].Archetype, "sultan_bulb")
	}

	if len(cat.Rooms) != 1 {
		t.Fatalf("got %d rooms, want 1", len(cat.Rooms))
	}
	room := cat.Rooms[0]
	if room.ID != "room-1" || room.Name != "Bureau" || room.On {
		t.Errorf("unexpected room: %+v", room)
	}
	if len(room.Lights) != 1 || room.Lights[0].ID != "light-2" || room.Lights[0].Name != "Suspension bureau" || room.Lights[0].On {
		t.Errorf("unexpected room lights: %+v", room.Lights)
	}
}

func TestCatalogResolve(t *testing.T) {
	cat := catalog.Catalog{
		Zones: []catalog.Group{{ID: "zone-1", Name: "Salon", Lights: []catalog.Light{{ID: "light-1", Name: "Lampe salon"}}}},
		Rooms: []catalog.Group{{ID: "room-1", Name: "Bureau", Lights: []catalog.Light{{ID: "light-2", Name: "Suspension bureau"}}}},
	}

	tests := []struct {
		name     string
		id       string
		wantType string
		wantName string
		wantOK   bool
	}{
		{"known zone", "zone-1", "zone", "Salon", true},
		{"unknown zone", "zone-999", "", "", false},
		{"known room", "room-1", "room", "Bureau", true},
		{"unknown room", "room-999", "", "", false},
		{"known light, found via its room", "light-2", "light", "Suspension bureau", true},
		{"unknown light", "light-999", "", "", false},
		// light-1 only appears under Zones in this fixture, not under any
		// Rooms entry — unrealistic on a real bridge (every light belongs to
		// exactly one room), but it pins down that Resolve checks Rooms for
		// light existence, not Zones.
		{"light only nested under a zone isn't found", "light-1", "", "", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotType, gotName, gotOK := cat.Resolve(tt.id)
			if gotType != tt.wantType || gotName != tt.wantName || gotOK != tt.wantOK {
				t.Errorf("Resolve(%q) = (%q, %q, %v), want (%q, %q, %v)",
					tt.id, gotType, gotName, gotOK, tt.wantType, tt.wantName, tt.wantOK)
			}
		})
	}
}

func TestBuild_PropagatesErrors(t *testing.T) {
	wantErr := errors.New("bridge unreachable")

	tests := []struct {
		name   string
		client *mockHue
	}{
		{
			name: "lights error",
			client: &mockHue{
				lightsFunc: func(ctx context.Context) ([]hue.Light, error) { return nil, wantErr },
			},
		},
		{
			name: "zones error",
			client: &mockHue{
				lightsFunc: func(ctx context.Context) ([]hue.Light, error) { return nil, nil },
				zonesFunc:  func(ctx context.Context) ([]hue.Zone, error) { return nil, wantErr },
			},
		},
		{
			name: "rooms error",
			client: &mockHue{
				lightsFunc: func(ctx context.Context) ([]hue.Light, error) { return nil, nil },
				zonesFunc:  func(ctx context.Context) ([]hue.Zone, error) { return nil, nil },
				roomsFunc:  func(ctx context.Context) ([]hue.Room, error) { return nil, wantErr },
			},
		},
		{
			name: "grouped lights error",
			client: &mockHue{
				lightsFunc:        func(ctx context.Context) ([]hue.Light, error) { return nil, nil },
				zonesFunc:         func(ctx context.Context) ([]hue.Zone, error) { return nil, nil },
				roomsFunc:         func(ctx context.Context) ([]hue.Room, error) { return nil, nil },
				groupedLightsFunc: func(ctx context.Context) ([]hue.GroupedLight, error) { return nil, wantErr },
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := catalog.Build(context.Background(), tt.client); !errors.Is(err, wantErr) {
				t.Errorf("Build() error = %v, want %v", err, wantErr)
			}
		})
	}
}
