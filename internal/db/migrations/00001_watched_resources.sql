-- +goose Up
CREATE TABLE watched_resources (
    resource_id   TEXT PRIMARY KEY,
    resource_type TEXT NOT NULL,
    name          TEXT NOT NULL,
    notify        INTEGER NOT NULL DEFAULT 1
);

-- +goose Down
DROP TABLE watched_resources;
