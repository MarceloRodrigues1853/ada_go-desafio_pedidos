-- name: GetProcessedEvent :one
SELECT order_id, saga_id, status, processed_at
FROM processed_events
WHERE order_id = $1 LIMIT 1;

-- name: CreateProcessedEvent :exec
INSERT INTO processed_events (order_id, saga_id, status)
VALUES ($1, $2, $3);
