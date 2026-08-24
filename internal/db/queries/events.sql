-- name: InsertEvent :one
INSERT INTO events (resource_id, resource_type, name, on_state, outcome)
VALUES (?, ?, ?, ?, ?)
RETURNING id, resource_id, resource_type, name, on_state, outcome, created_at;

-- name: ListEvents :many
SELECT id, resource_id, resource_type, name, on_state, outcome, created_at
FROM events
ORDER BY id DESC
LIMIT ?;


-- name: ListSessionsSince :many
-- Every "on" period: a turn-on paired with the first turn-off that follows it
-- for the same resource. end_at is the empty string while the resource is
-- still on: the correlated subquery yields NULL there, and sqlc types a
-- subquery column as non-null TEXT; COALESCE keeps the scan from failing and
-- the CAST keeps sqlc typing the column as string rather than interface{}.
-- The NOT EXISTS keeps any session that overlaps the window rather than only
-- those starting inside it, so a light switched on before `from` and off after
-- it still counts; the caller clips the overhang to its own window. Ordering on
-- id, not created_at, since CURRENT_TIMESTAMP only resolves to the second.
SELECT e.resource_id,
       e.resource_type,
       e.name,
       e.created_at AS start_at,
       CAST(COALESCE((SELECT o.created_at
                   FROM events o
                  WHERE o.resource_id = e.resource_id AND o.on_state = 0 AND o.id > e.id
                  ORDER BY o.id
                  LIMIT 1), '') AS TEXT) AS end_at
FROM events e
WHERE e.on_state = 1
  AND NOT EXISTS (SELECT 1
                    FROM events o
                   WHERE o.resource_id = e.resource_id
                     AND o.on_state = 0
                     AND o.id > e.id
                     AND o.created_at < ?)
ORDER BY e.created_at;
