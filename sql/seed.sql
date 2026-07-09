-- Seed de Produtos Iniciais
INSERT INTO produtos (id, nome, preco, estoque) VALUES 
('P001', 'Notebook Dell', 4500.00, 10),
('P002', 'Mouse Sem Fio Logitech', 150.00, 50),
('P003', 'Monitor LG Ultrawide', 1200.00, 20),
('P004', 'Teclado Mecânico Keychron', 650.00, 15)
ON CONFLICT (id) DO NOTHING;