-- +goose Up
CREATE TABLE events (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    resource_id   TEXT    NOT NULL,
    resource_type TEXT    NOT NULL,
    name          TEXT    NOT NULL,
    on_state      INTEGER NOT NULL,
    -- 'sent', 'muted', or 'channel_off', fixed at insert time
    outcome       TEXT    NOT NULL DEFAULT 'sent',
    created_at    TEXT    NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- +goose Down
DROP TABLE events;
