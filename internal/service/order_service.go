package service

import (
	"context"
	"errors"

	"pedidos/internal/domain/order"
	"pedidos/internal/repository/db"

	"github.com/google/uuid"
)

// OrderService agora depende exclusivamente das consultas do PostgreSQL geradas pelo sqlc
type OrderService struct {
	queries *db.Queries
}

// NewOrderService injeta a conexão do banco de dados na camada de serviço
func NewOrderService(queries *db.Queries) *OrderService {
	return &OrderService{
		queries: queries,
	}
}

// Create processa a intenção de compra, valida o cliente/produto e salva o pedido
func (s *OrderService) Create(ctx context.Context, clienteID uuid.UUID, itens []db.CreateItemPedidoParams) (*db.Pedido, error) {
	// 1. Valida se o cliente existe no banco (Regra de Negócio)
	_, err := s.queries.GetCliente(ctx, clienteID)
	if err != nil {
		return nil, errors.New("cliente não encontrado")
	}

	// 2. Cria a "capa" do pedido
	pedido, err := s.queries.CreatePedido(ctx, clienteID)
	if err != nil {
		return nil, errors.New("erro ao criar o pedido")
	}

	// 3. Adiciona os itens no carrinho e salva no banco
	for _, item := range itens {
		item.PedidoID = pedido.ID
		_, err := s.queries.CreateItemPedido(ctx, item)
		if err != nil {
			return nil, errors.New("erro ao adicionar item ao pedido")
		}
	}

	return &pedido, nil
}

// Pay refatorado com DDD
func (s *OrderService) Pay(ctx context.Context, pedidoID uuid.UUID) error {
	// 1. Busca os dados brutos no banco
	dbPedido, err := s.queries.GetPedidoByID(ctx, pedidoID)
	if err != nil {
		return errors.New("pedido não encontrado")
	}

	// 2. Transforma o banco na Entidade Pura de Domínio
	// (Você pode criar um construtor RestaurarPedido em domain/order)
	domainOrder := order.Restore(
		dbPedido.ID,
		dbPedido.ClienteID,
		order.Status(dbPedido.Status),
		[]order.OrderItem{},
	)

	// 3. Executa a REGRA DE NEGÓCIO DO DOMÍNIO (O domínio que sabe se pode pagar!)
	if err := domainOrder.Pay(); err != nil {
		return err // Retorna "ErrCannotPay..." validado pelo próprio domínio
	}

	// 4. Salva a alteração no banco
	return s.queries.UpdatePedidoStatus(ctx, db.UpdatePedidoStatusParams{
		ID:     pedidoID,
		Status: string(domainOrder.Status()),
	})
}

// Cancel altera o status do pedido para 'CANCELED'
func (s *OrderService) Cancel(ctx context.Context, pedidoID uuid.UUID) error {
	// 1. Verifica se o pedido existe
	pedido, err := s.queries.GetPedidoByID(ctx, pedidoID)
	if err != nil {
		return errors.New("pedido não encontrado")
	}

	if pedido.Status == "PAID" {
		return errors.New("não é possível cancelar um pedido que já foi pago")
	}

	// 2. Atualiza o status no banco de dados
	return s.queries.UpdatePedidoStatus(ctx, db.UpdatePedidoStatusParams{
		ID:     pedidoID,
		Status: "CANCELED",
	})
}
