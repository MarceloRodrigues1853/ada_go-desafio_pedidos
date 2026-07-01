package repository

import "pedidos/internal/domain"

// OrderRepository define o contrato que qualquer banco de dados de pedidos deve seguir
type OrderRepository interface {
	Salvar(pedido *domain.Pedido) error
	BuscarPorID(id string) (*domain.Pedido, error)
	Listar() ([]*domain.Pedido, error)
}
