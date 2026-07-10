CREATE TABLE pedidos (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    cliente_id UUID NOT NULL REFERENCES clientes(id),
    status VARCHAR(50) NOT NULL DEFAULT 'PENDING',
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

CREATE TABLE itens_pedido (
    id UUID PRIMARY KEY DEFAULT uuidv7(),
    pedido_id UUID NOT NULL REFERENCES pedidos(id) ON DELETE CASCADE,
    produto_id VARCHAR(50) NOT NULL REFERENCES produtos(id),
    quantidade INT NOT NULL,
    preco_unitario DECIMAL(10,2) NOT NULL
);