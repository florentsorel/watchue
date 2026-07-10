-- name: InsertEvent :one
INSERT INTO events (resource_id, resource_type, name, on_state, outcome)
VALUES (?, ?, ?, ?, ?)
RETURNING id, resource_id, resource_type, name, on_state, outcome, created_at;

-- name: ListEvents :many
SELECT id, resource_id, resource_type, name, on_state, outcome, created_at
FROM events
ORDER BY id DESC
LIMIT ?;
