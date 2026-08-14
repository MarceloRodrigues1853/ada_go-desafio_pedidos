ALTER TABLE produtos
    ADD CONSTRAINT produtos_preco_positivo CHECK (preco > 0),
    ADD CONSTRAINT produtos_estoque_nao_negativo CHECK (estoque >= 0);

ALTER TABLE pedidos
    ADD CONSTRAINT pedidos_status_valido CHECK (status IN ('PENDING', 'PAID', 'CANCELED'));

ALTER TABLE itens_pedido
    ADD CONSTRAINT itens_pedido_quantidade_positiva CHECK (quantidade > 0),
    ADD CONSTRAINT itens_pedido_preco_positivo CHECK (preco_unitario > 0);
