// Package watch matches incoming Hue eventstream data against the set of
// resources the user has chosen to watch.
package watch

import (
	"context"
	"database/sql"
	"errors"

	"github.com/florentsorel/watchue/internal/db"
	"github.com/florentsorel/watchue/internal/hue"
)

// Queries is the subset of db.Queries operations Match needs.
type Queries interface {
	GetWatchedResource(ctx context.Context, resourceID string) (db.WatchedResource, error)
}

// Change describes a watched resource whose on/off state just changed.
type Change struct {
	ResourceID   string
	ResourceType string
	Name         string
	On           bool
	Notify       bool // false if the user muted this resource; still record it, just don't send
}

// ResolveResourceID returns the watched_resources id an on/off event belongs
// to — a light's own id, or (for a grouped_light) the owning zone/room's id
// — and false if ev isn't an on/off change at all.
func ResolveResourceID(ev hue.EventData) (resourceID string, ok bool) {
	if ev.On == nil {
		return "", false
	}
	if ev.Type == hue.ResourceGroupedLight {
		if ev.Owner == nil {
			return "", false
		}
		return ev.Owner.RID, true
	}
	return ev.ID, true
}

// Match reports whether ev is an on/off change for a watched resource. ok is
// false for a non-match — the expected outcome for most events, not an error.
func Match(ctx context.Context, queries Queries, ev hue.EventData) (change Change, ok bool, err error) {
	resourceID, ok := ResolveResourceID(ev)
	if !ok {
		return Change{}, false, nil
	}

	watched, err := queries.GetWatchedResource(ctx, resourceID)
	if errors.Is(err, sql.ErrNoRows) {
		return Change{}, false, nil
	}
	if err != nil {
		return Change{}, false, err
	}

	return Change{
		ResourceID:   watched.ResourceID,
		ResourceType: watched.ResourceType,
		Name:         watched.Name,
		On:           ev.On.On,
		Notify:       watched.Notify != 0,
	}, true, nil
}
