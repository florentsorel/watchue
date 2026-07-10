package watch_test

import (
	"context"
	"database/sql"
	"errors"
	"testing"

	"github.com/florentsorel/watchue/internal/db"
	"github.com/florentsorel/watchue/internal/hue"
	"github.com/florentsorel/watchue/internal/watch"
)

func TestResolveResourceID(t *testing.T) {
	tests := []struct {
		name   string
		ev     hue.EventData
		wantID string
		wantOK bool
	}{
		{"no on state", hue.EventData{ID: "light-1", Type: hue.ResourceLight}, "", false},
		{"light", hue.EventData{ID: "light-1", Type: hue.ResourceLight, On: &hue.OnState{On: true}}, "light-1", true},
		{
			"grouped_light with owner",
			hue.EventData{ID: "grouped-1", Type: hue.ResourceGroupedLight, Owner: &hue.ResourceIdentifier{RID: "zone-1"}, On: &hue.OnState{On: true}},
			"zone-1", true,
		},
		{
			"grouped_light without owner",
			hue.EventData{ID: "grouped-1", Type: hue.ResourceGroupedLight, On: &hue.OnState{On: true}},
			"", false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotID, gotOK := watch.ResolveResourceID(tt.ev)
			if gotID != tt.wantID || gotOK != tt.wantOK {
				t.Errorf("ResolveResourceID(%+v) = (%q, %v), want (%q, %v)", tt.ev, gotID, gotOK, tt.wantID, tt.wantOK)
			}
		})
	}
}

type mockQueries struct {
	getWatchedResourceFunc func(ctx context.Context, resourceID string) (db.WatchedResource, error)
}

func (m *mockQueries) GetWatchedResource(ctx context.Context, resourceID string) (db.WatchedResource, error) {
	return m.getWatchedResourceFunc(ctx, resourceID)
}

var _ watch.Queries = (*mockQueries)(nil)

func TestMatch_IgnoresNonOnOffEvents(t *testing.T) {
	queries := &mockQueries{
		getWatchedResourceFunc: func(ctx context.Context, resourceID string) (db.WatchedResource, error) {
			t.Fatal("GetWatchedResource should not be called for an event with no On state")
			return db.WatchedResource{}, nil
		},
	}

	_, ok, err := watch.Match(context.Background(), queries, hue.EventData{ID: "light-1", Type: hue.ResourceLight})
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if ok {
		t.Error("ok = true, want false for an event with no On state")
	}
}

func TestMatch_WatchedLight(t *testing.T) {
	var gotID string
	queries := &mockQueries{
		getWatchedResourceFunc: func(ctx context.Context, resourceID string) (db.WatchedResource, error) {
			gotID = resourceID
			return db.WatchedResource{ResourceID: "light-1", ResourceType: "light", Name: "Lampe salon", Notify: 1}, nil
		},
	}

	change, ok, err := watch.Match(context.Background(), queries, hue.EventData{
		ID:   "light-1",
		Type: hue.ResourceLight,
		On:   &hue.OnState{On: true},
	})
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if !ok {
		t.Fatal("ok = false, want true for a watched light")
	}
	if gotID != "light-1" {
		t.Errorf("GetWatchedResource called with %q, want %q", gotID, "light-1")
	}
	want := watch.Change{ResourceID: "light-1", ResourceType: "light", Name: "Lampe salon", On: true, Notify: true}
	if change != want {
		t.Errorf("change = %+v, want %+v", change, want)
	}
}

func TestMatch_UnwatchedLight(t *testing.T) {
	queries := &mockQueries{
		getWatchedResourceFunc: func(ctx context.Context, resourceID string) (db.WatchedResource, error) {
			return db.WatchedResource{}, sql.ErrNoRows
		},
	}

	_, ok, err := watch.Match(context.Background(), queries, hue.EventData{
		ID:   "light-2",
		Type: hue.ResourceLight,
		On:   &hue.OnState{On: true},
	})
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if ok {
		t.Error("ok = true, want false for an unwatched light")
	}
}

func TestMatch_WatchedZoneViaGroupedLight(t *testing.T) {
	var gotID string
	queries := &mockQueries{
		getWatchedResourceFunc: func(ctx context.Context, resourceID string) (db.WatchedResource, error) {
			gotID = resourceID
			return db.WatchedResource{ResourceID: "zone-1", ResourceType: "zone", Name: "Salon", Notify: 1}, nil
		},
	}

	change, ok, err := watch.Match(context.Background(), queries, hue.EventData{
		ID:    "grouped-1",
		Type:  hue.ResourceGroupedLight,
		Owner: &hue.ResourceIdentifier{RID: "zone-1", RType: hue.ResourceZone},
		On:    &hue.OnState{On: false},
	})
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if !ok {
		t.Fatal("ok = false, want true for a watched zone")
	}
	if gotID != "zone-1" {
		t.Errorf("GetWatchedResource called with %q, want the zone id %q (not the grouped_light id)", gotID, "zone-1")
	}
	want := watch.Change{ResourceID: "zone-1", ResourceType: "zone", Name: "Salon", On: false, Notify: true}
	if change != want {
		t.Errorf("change = %+v, want %+v", change, want)
	}
}

func TestMatch_MutedResourceStillMatchesButNotifyIsFalse(t *testing.T) {
	queries := &mockQueries{
		getWatchedResourceFunc: func(ctx context.Context, resourceID string) (db.WatchedResource, error) {
			return db.WatchedResource{ResourceID: "light-1", ResourceType: "light", Name: "Lampe salon", Notify: 0}, nil
		},
	}

	change, ok, err := watch.Match(context.Background(), queries, hue.EventData{
		ID:   "light-1",
		Type: hue.ResourceLight,
		On:   &hue.OnState{On: true},
	})
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if !ok {
		t.Fatal("ok = false, want true — a muted resource is still matched (and its history recorded)")
	}
	if change.Notify {
		t.Error("Notify = true, want false for a muted resource")
	}
}

func TestMatch_GroupedLightWithoutOwner(t *testing.T) {
	queries := &mockQueries{
		getWatchedResourceFunc: func(ctx context.Context, resourceID string) (db.WatchedResource, error) {
			t.Fatal("GetWatchedResource should not be called when Owner is nil")
			return db.WatchedResource{}, nil
		},
	}

	_, ok, err := watch.Match(context.Background(), queries, hue.EventData{
		ID:   "grouped-1",
		Type: hue.ResourceGroupedLight,
		On:   &hue.OnState{On: true},
	})
	if err != nil {
		t.Fatalf("Match() error = %v", err)
	}
	if ok {
		t.Error("ok = true, want false when a grouped_light event has no Owner")
	}
}

func TestMatch_PropagatesQueryError(t *testing.T) {
	wantErr := errors.New("db unavailable")
	queries := &mockQueries{
		getWatchedResourceFunc: func(ctx context.Context, resourceID string) (db.WatchedResource, error) {
			return db.WatchedResource{}, wantErr
		},
	}

	_, ok, err := watch.Match(context.Background(), queries, hue.EventData{
		ID:   "light-1",
		Type: hue.ResourceLight,
		On:   &hue.OnState{On: true},
	})
	if !errors.Is(err, wantErr) {
		t.Errorf("Match() error = %v, want %v", err, wantErr)
	}
	if ok {
		t.Error("ok = true, want false on query error")
	}
}
