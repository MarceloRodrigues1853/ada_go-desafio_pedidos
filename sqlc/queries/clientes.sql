-- name: CreateCliente :one
INSERT INTO clientes (name, email, password_hash)
VALUES ($1, $2, $3)
RETURNING id, name, email, password_hash, created_at;

-- name: GetCliente :one
SELECT id, name, email, password_hash, created_at
FROM clientes
WHERE id = $1 LIMIT 1;

-- name: ListClientes :many
SELECT id, name, email, password_hash, created_at
FROM clientes
ORDER BY created_at DESC;