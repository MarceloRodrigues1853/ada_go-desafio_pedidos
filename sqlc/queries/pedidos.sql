-- name: CreatePedido :one
-- Cria a "capa" do pedido vinculada a um cliente. O status padrão ('PENDING') e o ID são gerados pelo banco.
INSERT INTO pedidos (cliente_id)
VALUES ($1)
RETURNING id, cliente_id, status, created_at;

-- name: CreateItemPedido :one
-- Adiciona um produto dentro do carrinho do pedido
INSERT INTO itens_pedido (pedido_id, produto_id, quantidade, preco_unitario)
VALUES ($1, $2, $3, $4)
RETURNING id, pedido_id, produto_id, quantidade, preco_unitario;

-- name: UpdatePedidoStatus :exec
-- Usado pelas rotas de Pagar e Cancelar para mudar o status ('PAID' ou 'CANCELED')
UPDATE pedidos
SET status = $2
WHERE id = $1;

-- name: GetPedidoByID :one
-- Busca os detalhes da capa do pedido para validações
SELECT id, cliente_id, status, created_at
FROM pedidos
WHERE id = $1 LIMIT 1;

-- name: GetItensPedido :many
-- Busca todos os itens cadastrados de um pedido específico
SELECT id, pedido_id, produto_id, quantidade, preco_unitario
FROM itens_pedido
WHERE pedido_id = $1;

-- name: ListPedidosPaginado :many
-- Busca pedidos do banco com limite e deslocamento (paginação)
SELECT id, cliente_id, status, created_at
FROM pedidos
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;