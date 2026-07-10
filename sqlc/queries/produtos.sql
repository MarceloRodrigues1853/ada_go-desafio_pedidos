-- name: CreateProduto :one
INSERT INTO produtos (id, nome, preco, estoque)
VALUES ($1, $2, $3, $4)
RETURNING id, nome, preco, estoque;

-- name: GetProdutoByID :one
SELECT id, nome, preco, estoque
FROM produtos
WHERE id = $1 LIMIT 1;

-- name: ListProdutos :many
SELECT id, nome, preco, estoque
FROM produtos
ORDER BY nome ASC;