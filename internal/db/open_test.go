package db_test

import (
	"context"
	"testing"

	"github.com/florentsorel/watchue/internal/db"
)

func TestOpen_RunsMigrations(t *testing.T) {
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)

	if _, err := q.ListWatchedResources(context.Background()); err != nil {
		t.Errorf("ListWatchedResources: %v", err)
	}
	if _, err := q.ListEvents(context.Background(), 50); err != nil {
		t.Errorf("ListEvents: %v", err)
	}
}

func TestUpsertWatchedResource(t *testing.T) {
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)
	ctx := context.Background()

	err = q.UpsertWatchedResource(ctx, db.UpsertWatchedResourceParams{
		ResourceID:   "zone-1",
		ResourceType: "zone",
		Name:         "Bureau",
	})
	if err != nil {
		t.Fatalf("UpsertWatchedResource: %v", err)
	}

	got, err := q.GetWatchedResource(ctx, "zone-1")
	if err != nil {
		t.Fatalf("GetWatchedResource: %v", err)
	}
	if got.ResourceType != "zone" || got.Name != "Bureau" {
		t.Errorf("unexpected watched resource: %+v", got)
	}
	if got.Notify != 1 {
		t.Errorf("Notify = %d, want 1 (default) on first watch", got.Notify)
	}

	// upsert on the same id updates in place rather than duplicating
	err = q.UpsertWatchedResource(ctx, db.UpsertWatchedResourceParams{
		ResourceID:   "zone-1",
		ResourceType: "zone",
		Name:         "Bureau renamed",
	})
	if err != nil {
		t.Fatalf("UpsertWatchedResource (update): %v", err)
	}

	all, err := q.ListWatchedResources(ctx)
	if err != nil {
		t.Fatalf("ListWatchedResources: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("got %d watched resources, want 1", len(all))
	}
	if all[0].Name != "Bureau renamed" {
		t.Errorf("Name = %q, want %q", all[0].Name, "Bureau renamed")
	}

	// muting, then re-watching (upsert) must not reset notify back to 1
	rows, err := q.SetWatchedResourceNotify(ctx, db.SetWatchedResourceNotifyParams{ResourceID: "zone-1", Notify: 0})
	if err != nil {
		t.Fatalf("SetWatchedResourceNotify: %v", err)
	}
	if rows != 1 {
		t.Fatalf("SetWatchedResourceNotify rows affected = %d, want 1", rows)
	}
	err = q.UpsertWatchedResource(ctx, db.UpsertWatchedResourceParams{
		ResourceID:   "zone-1",
		ResourceType: "zone",
		Name:         "Bureau renamed again",
	})
	if err != nil {
		t.Fatalf("UpsertWatchedResource (re-watch after mute): %v", err)
	}
	got, err = q.GetWatchedResource(ctx, "zone-1")
	if err != nil {
		t.Fatalf("GetWatchedResource: %v", err)
	}
	if got.Notify != 0 {
		t.Errorf("Notify = %d, want 0 (re-watching must not un-mute)", got.Notify)
	}

	// setting notify on an id that isn't watched affects zero rows
	rows, err = q.SetWatchedResourceNotify(ctx, db.SetWatchedResourceNotifyParams{ResourceID: "does-not-exist", Notify: 1})
	if err != nil {
		t.Fatalf("SetWatchedResourceNotify (unknown id): %v", err)
	}
	if rows != 0 {
		t.Errorf("SetWatchedResourceNotify rows affected = %d, want 0 for an unknown id", rows)
	}

	if err := q.DeleteWatchedResource(ctx, "zone-1"); err != nil {
		t.Fatalf("DeleteWatchedResource: %v", err)
	}
	all, err = q.ListWatchedResources(ctx)
	if err != nil {
		t.Fatalf("ListWatchedResources: %v", err)
	}
	if len(all) != 0 {
		t.Errorf("got %d watched resources after delete, want 0", len(all))
	}
}

func TestInsertAndListEvents(t *testing.T) {
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)
	ctx := context.Background()

	events := []db.InsertEventParams{
		{ResourceID: "zone-1", ResourceType: "zone", Name: "Salon", OnState: 1, Outcome: "sent"},
		{ResourceID: "light-1", ResourceType: "light", Name: "Lampe salon", OnState: 0, Outcome: "muted"},
		{ResourceID: "zone-1", ResourceType: "zone", Name: "Salon", OnState: 0, Outcome: "channel_off"},
	}
	for _, e := range events {
		inserted, err := q.InsertEvent(ctx, e)
		if err != nil {
			t.Fatalf("InsertEvent(%+v): %v", e, err)
		}
		if inserted.ID == 0 || inserted.CreatedAt == "" {
			t.Errorf("InsertEvent(%+v) returned = %+v, want a populated id/created_at", e, inserted)
		}
	}

	all, err := q.ListEvents(ctx, 50)
	if err != nil {
		t.Fatalf("ListEvents: %v", err)
	}
	if len(all) != 3 {
		t.Fatalf("got %d events, want 3", len(all))
	}
	// most recent first
	if all[0].ResourceID != "zone-1" || all[0].OnState != 0 || all[0].Outcome != "channel_off" {
		t.Errorf("most recent event = %+v, want zone-1 off/channel_off", all[0])
	}
	if all[2].ResourceID != "zone-1" || all[2].OnState != 1 || all[2].Outcome != "sent" {
		t.Errorf("oldest event = %+v, want zone-1 on/sent", all[2])
	}

	limited, err := q.ListEvents(ctx, 2)
	if err != nil {
		t.Fatalf("ListEvents (limit 2): %v", err)
	}
	if len(limited) != 2 {
		t.Errorf("got %d events with limit 2, want 2", len(limited))
	}
}

func TestBoolSetting(t *testing.T) {
	conn, err := db.Open(":memory:")
	if err != nil {
		t.Fatalf("Open: %v", err)
	}
	defer conn.Close()

	q := db.New(conn)
	ctx := context.Background()

	// never set: falls back to the given default
	got, err := q.GetBoolSetting(ctx, db.TelegramEnabledKey, true)
	if err != nil {
		t.Fatalf("GetBoolSetting: %v", err)
	}
	if !got {
		t.Error("GetBoolSetting = false, want the default (true) when unset")
	}

	if err := q.SetBoolSetting(ctx, db.TelegramEnabledKey, false); err != nil {
		t.Fatalf("SetBoolSetting: %v", err)
	}
	got, err = q.GetBoolSetting(ctx, db.TelegramEnabledKey, true)
	if err != nil {
		t.Fatalf("GetBoolSetting: %v", err)
	}
	if got {
		t.Error("GetBoolSetting = true, want false after SetBoolSetting(false)")
	}

	// setting again (not just inserting) updates in place
	if err := q.SetBoolSetting(ctx, db.TelegramEnabledKey, true); err != nil {
		t.Fatalf("SetBoolSetting (update): %v", err)
	}
	got, err = q.GetBoolSetting(ctx, db.TelegramEnabledKey, false)
	if err != nil {
		t.Fatalf("GetBoolSetting: %v", err)
	}
	if !got {
		t.Error("GetBoolSetting = false, want true after SetBoolSetting(true)")
	}
}
