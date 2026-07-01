package repository

import "pedidos/internal/domain"

// ProductRepository define o contrato que qualquer banco de dados de produtos deve seguir
type ProductRepository interface {
	Salvar(produto *domain.Produto) error
	BuscarPorID(id string) (*domain.Produto, error)
	Listar() ([]*domain.Produto, error)
}
