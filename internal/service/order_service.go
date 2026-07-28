package service

import (
	"context"
	"errors"

	"pedidos/internal/domain/order"
	"pedidos/internal/repository/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// OrderService coordena os casos de uso de pedidos e depende das queries do sqlc e do pool para transações
type OrderService struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

// NewOrderService injeta a conexão e o pool de conexões do banco de dados na camada de serviço
func NewOrderService(queries *db.Queries, pool *pgxpool.Pool) *OrderService {
	return &OrderService{
		queries: queries,
		pool:    pool,
	}
}

// Create processa a intenção de compra usando TRANSAÇÃO para garantir atomicidade no estoque e itens
func (s *OrderService) Create(ctx context.Context, clienteID uuid.UUID, itens []db.CreateItemPedidoParams) (*db.Pedido, error) {
	// 1. Valida se o cliente existe no banco de dados
	_, err := s.queries.GetCliente(ctx, clienteID)
	if err != nil {
		return nil, errors.New("cliente não encontrado")
	}

	// 2. Valida se o pedido contém ao menos 1 item (Invariante)
	if len(itens) == 0 {
		return nil, errors.New("o pedido precisa ter pelo menos um item")
	}

	// 3. Inicia a Transação no PostgreSQL (REQUISITO OBRIGATÓRIO DO PROFESSOR)
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, errors.New("erro ao iniciar transação no banco de dados")
	}
	defer tx.Rollback(ctx) // Garante que se houver erro no meio do caminho, nada é salvo no banco

	// Instancia as queries associadas à transação atual
	qtx := s.queries.WithTx(tx)

	// 4. Valida produtos e reduz estoque individualmente dentro da transação
	for _, item := range itens {
		produtoUUID, err := uuid.Parse(item.ProdutoID)
		if err != nil {
			return nil, errors.New("ID do produto inválido")
		}

		// Busca produto para conferir estoque
		produto, err := qtx.GetProduto(ctx, produtoUUID)
		if err != nil {
			return nil, errors.New("produto não encontrado")
		}

		if produto.Estoque < item.Quantidade {
			return nil, errors.New("estoque insuficiente para o produto")
		}

		// Reduz o estoque do produto no banco
		err = qtx.ReduzirEstoque(ctx, db.ReduzirEstoqueParams{
			ID:         produtoUUID,
			Quantidade: item.Quantidade,
		})
		if err != nil {
			return nil, errors.New("erro ao atualizar estoque do produto")
		}
	}

	// 5. Cria a capa do pedido no banco
	pedido, err := qtx.CreatePedido(ctx, clienteID)
	if err != nil {
		return nil, errors.New("erro ao criar a capa do pedido")
	}

	// 6. Insere os itens vinculados ao pedido criado
	for _, item := range itens {
		item.PedidoID = pedido.ID
		_, err := qtx.CreateItemPedido(ctx, item)
		if err != nil {
			return nil, errors.New("erro ao adicionar item ao pedido")
		}
	}

	// 7. Confirma e efetiva a transação no banco de dados (Commit)
	if err := tx.Commit(ctx); err != nil {
		return nil, errors.New("erro ao finalizar transação do pedido")
	}

	return &pedido, nil
}

// GetByID busca um pedido único pelo ID (NOVO - Requisito do enunciado)
func (s *OrderService) GetByID(ctx context.Context, pedidoID uuid.UUID) (*db.Pedido, error) {
	pedido, err := s.queries.GetPedidoByID(ctx, pedidoID)
	if err != nil {
		return nil, errors.New("pedido não encontrado")
	}
	return &pedido, nil
}

// ListPaginado retorna lista de pedidos com limit e offset (NOVO - Requisito do enunciado)
func (s *OrderService) ListPaginado(ctx context.Context, limit, offset int32) ([]db.Pedido, error) {
	return s.queries.ListPedidosPaginado(ctx, db.ListPedidosPaginadoParams{
		Limit:  limit,
		Offset: offset,
	})
}

// Pay altera o status do pedido para 'PAID'
func (s *OrderService) Pay(ctx context.Context, pedidoID uuid.UUID) error {
	// 1. Busca os dados brutos no banco
	dbPedido, err := s.queries.GetPedidoByID(ctx, pedidoID)
	if err != nil {
		return errors.New("pedido não encontrado")
	}

	// 2. Transforma em entidade pura do domínio
	domainOrder := order.Restore(
		dbPedido.ID,
		dbPedido.ClienteID,
		order.Status(dbPedido.Status),
		[]order.OrderItem{},
	)

	// 3. Aplica regra de negócio de pagamento
	if err := domainOrder.Pay(); err != nil {
		return err
	}

	// 4. Atualiza o status no banco de dados
	return s.queries.UpdatePedidoStatus(ctx, db.UpdatePedidoStatusParams{
		ID:     pedidoID,
		Status: string(domainOrder.Status()),
	})
}

// Cancel altera o status para 'CANCELED' e devolve os itens ao estoque
func (s *OrderService) Cancel(ctx context.Context, pedidoID uuid.UUID) error {
	// 1. Verifica se o pedido existe
	dbPedido, err := s.queries.GetPedidoByID(ctx, pedidoID)
	if err != nil {
		return errors.New("pedido não encontrado")
	}

	// Valida se o pedido já está pago ou cancelado (Regra do professor)
	if dbPedido.Status == "PAID" {
		return errors.New("não é possível cancelar um pedido que já foi pago")
	}
	if dbPedido.Status == "CANCELED" {
		return errors.New("este pedido já está cancelado")
	}

	// 2. Abre transação para devolver o estoque e atualizar o status
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return errors.New("erro ao iniciar transação")
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	// 3. Busca os itens do pedido para realizar a devolução do estoque
	itens, err := qtx.GetItensPedido(ctx, pedidoID)
	if err == nil {
		for _, item := range itens {
			produtoUUID, err := uuid.Parse(item.ProdutoID)
			if err == nil {
				// Devolve a quantidade para o estoque no banco de dados
				_ = qtx.DevolverEstoque(ctx, db.DevolverEstoqueParams{
					ID:         produtoUUID,
					Quantidade: item.Quantidade,
				})
			}
		}
	}

	// 4. Atualiza o status para CANCELED
	err = qtx.UpdatePedidoStatus(ctx, db.UpdatePedidoStatusParams{
		ID:     pedidoID,
		Status: "CANCELED",
	})
	if err != nil {
		return errors.New("erro ao atualizar status do pedido")
	}

	// 5. Confirma as alterações no banco
	return tx.Commit(ctx)
}
