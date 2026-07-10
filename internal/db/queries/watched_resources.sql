-- name: ListWatchedResources :many
SELECT resource_id, resource_type, name, notify FROM watched_resources;

-- name: GetWatchedResource :one
SELECT resource_id, resource_type, name, notify FROM watched_resources WHERE resource_id = ?;

-- name: UpsertWatchedResource :exec
INSERT INTO watched_resources (resource_id, resource_type, name)
VALUES (?, ?, ?)
ON CONFLICT (resource_id) DO UPDATE SET resource_type = excluded.resource_type, name = excluded.name;
-- notify excluded on purpose: re-watching must not un-mute

-- name: SetWatchedResourceNotify :execrows
UPDATE watched_resources SET notify = ? WHERE resource_id = ?;

-- name: DeleteWatchedResource :exec
DELETE FROM watched_resources WHERE resource_id = ?;
