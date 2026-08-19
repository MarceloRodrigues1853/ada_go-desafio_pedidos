// Package repository contains persistence ports and their PostgreSQL adapters.
package repository

import (
	"context"
	"errors"
	"time"

	"pedidos/internal/domain"
	"pedidos/internal/domain/order"
	"pedidos/internal/repository/db"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	ErrNotFound = errors.New("repository: not found") // Registro não existe (traduzido de pgx.ErrNoRows)
	ErrConflict = errors.New("repository: conflict")  // Violação de unicidade ou atualização concorrente
)

// ClientRepository define a porta de persistência da entidade Cliente.
type ClientRepository interface {
	Create(context.Context, *domain.Cliente) (*domain.Cliente, error)
	List(context.Context) ([]domain.Cliente, error)
	GetByID(context.Context, uuid.UUID) (*domain.Cliente, error)
}

// ProductRepository define a porta de persistência da entidade Produto.
type ProductRepository interface {
	Create(context.Context, *domain.Produto) (*domain.Produto, error)
	List(context.Context) ([]domain.Produto, error)
	GetByID(context.Context, string) (*domain.Produto, error)
}

// OrderRecord representa um pedido persistido junto com sua data de criação.
type OrderRecord struct {
	Order     *order.Order
	CreatedAt time.Time
}

// OrderRepository define a porta de persistência do agregado Order.
type OrderRepository interface {
	GetByID(context.Context, uuid.UUID) (*OrderRecord, error)
	List(context.Context, int32, int32) ([]OrderRecord, error)
	UpdateStatus(context.Context, uuid.UUID, order.Status) error
	WithinTransaction(context.Context, func(OrderTransaction) error) error
}

// PaymentRepository define a porta de persistência da idempotência de pagamentos.
type PaymentRepository interface {
	IsProcessed(context.Context, uuid.UUID) (bool, error)
	MarkAsProcessed(context.Context, uuid.UUID, uuid.UUID, string) error
}

// OrderTransaction expõe as operações atômicas executadas dentro de uma transação.
type OrderTransaction interface {
	ReserveStock(context.Context, string, int) error
	ReturnStock(context.Context, string, int) error
	Create(context.Context, *order.Order) (*OrderRecord, error)
	UpdateStatus(context.Context, uuid.UUID, order.Status) error
}

// ClientPostgresRepository é o adaptador PostgreSQL da porta ClientRepository (via sqlc).
type ClientPostgresRepository struct{ queries *db.Queries }

// NewClientPostgresRepository cria o repositório de clientes com as queries geradas pelo sqlc.
func NewClientPostgresRepository(queries *db.Queries) *ClientPostgresRepository {
	return &ClientPostgresRepository{queries}
}
func (r *ClientPostgresRepository) Create(ctx context.Context, client *domain.Cliente) (*domain.Cliente, error) {
	row, err := r.queries.CreateCliente(ctx, db.CreateClienteParams{Name: client.Name, Email: client.Email, PasswordHash: client.PasswordHash})
	if err != nil {
		return nil, translateError(err)
	}
	return clientFromDB(row), nil
}
func (r *ClientPostgresRepository) List(ctx context.Context) ([]domain.Cliente, error) {
	rows, err := r.queries.ListClientes(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	result := make([]domain.Cliente, 0, len(rows))
	for _, row := range rows {
		result = append(result, *clientFromDB(row))
	}
	return result, nil
}
func (r *ClientPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cliente, error) {
	row, err := r.queries.GetCliente(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return clientFromDB(row), nil
}

// ProductPostgresRepository é o adaptador PostgreSQL da porta ProductRepository (via sqlc).
type ProductPostgresRepository struct{ queries *db.Queries }

// NewProductPostgresRepository cria o repositório de produtos com as queries geradas pelo sqlc.
func NewProductPostgresRepository(queries *db.Queries) *ProductPostgresRepository {
	return &ProductPostgresRepository{queries}
}
func (r *ProductPostgresRepository) Create(ctx context.Context, product *domain.Produto) (*domain.Produto, error) {
	row, err := r.queries.CreateProduto(ctx, db.CreateProdutoParams{ID: product.ID, Nome: product.Nome, Preco: product.Preco, Estoque: int32(product.Estoque)})
	if err != nil {
		return nil, translateError(err)
	}
	return productFromDB(row), nil
}
func (r *ProductPostgresRepository) List(ctx context.Context) ([]domain.Produto, error) {
	rows, err := r.queries.ListProdutos(ctx)
	if err != nil {
		return nil, translateError(err)
	}
	result := make([]domain.Produto, 0, len(rows))
	for _, row := range rows {
		result = append(result, *productFromDB(row))
	}
	return result, nil
}
func (r *ProductPostgresRepository) GetByID(ctx context.Context, id string) (*domain.Produto, error) {
	row, err := r.queries.GetProduto(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return productFromDB(row), nil
}

// OrderPostgresRepository é o adaptador PostgreSQL do agregado Order.
// Usa o pool diretamente para transações e o sqlc para as queries simples.
type OrderPostgresRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

// NewOrderPostgresRepository cria o repositório de pedidos.
func NewOrderPostgresRepository(queries *db.Queries, pool *pgxpool.Pool) *OrderPostgresRepository {
	return &OrderPostgresRepository{queries, pool}
}
func (r *OrderPostgresRepository) GetByID(ctx context.Context, id uuid.UUID) (*OrderRecord, error) {
	row, err := r.queries.GetPedidoByID(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	items, err := r.queries.GetItensPedido(ctx, id)
	if err != nil {
		return nil, translateError(err)
	}
	return orderRecordFromDB(row, items)
}
func (r *OrderPostgresRepository) List(ctx context.Context, limit, offset int32) ([]OrderRecord, error) {
	rows, err := r.queries.ListPedidosPaginado(ctx, db.ListPedidosPaginadoParams{Limit: limit, Offset: offset})
	if err != nil {
		return nil, translateError(err)
	}
	result := make([]OrderRecord, 0, len(rows))
	for _, row := range rows {
		record, err := orderRecordFromDB(row, nil)
		if err != nil {
			return nil, err
		}
		result = append(result, *record)
	}
	return result, nil
}
func (r *OrderPostgresRepository) UpdateStatus(ctx context.Context, id uuid.UUID, status order.Status) error {
	command, err := r.pool.Exec(ctx, "UPDATE pedidos SET status = $2 WHERE id = $1 AND status = 'PENDING'", id, string(status))
	if err != nil {
		return translateError(err)
	}
	if command.RowsAffected() == 0 {
		return ErrConflict
	}
	return nil
}
func (r *OrderPostgresRepository) WithinTransaction(ctx context.Context, fn func(OrderTransaction) error) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if err := fn(&postgresOrderTransaction{tx, r.queries.WithTx(tx)}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// postgresOrderTransaction implementa OrderTransaction sobre uma transação pgx real.
type postgresOrderTransaction struct {
	tx      pgx.Tx
	queries *db.Queries // Queries apontando para a transação (via WithTx)
}

func (t *postgresOrderTransaction) ReserveStock(ctx context.Context, id string, quantity int) error {
	if quantity <= 0 {
		return domain.ErrQuantidadeInvalida
	}
	command, err := t.tx.Exec(ctx, "UPDATE produtos SET estoque = estoque - $2 WHERE id = $1 AND estoque >= $2", id, quantity)
	if err != nil {
		return err
	}
	if command.RowsAffected() == 0 {
		return domain.ErrEstoqueInsuficiente
	}
	return nil
}
func (t *postgresOrderTransaction) ReturnStock(ctx context.Context, id string, quantity int) error {
	if quantity <= 0 {
		return domain.ErrQuantidadeInvalida
	}
	return t.queries.DevolverEstoque(ctx, db.DevolverEstoqueParams{ID: id, Estoque: int32(quantity)})
}
func (t *postgresOrderTransaction) Create(ctx context.Context, aggregate *order.Order) (*OrderRecord, error) {
	row, err := t.queries.CreatePedido(ctx, aggregate.ClientID())
	if err != nil {
		return nil, translateError(err)
	}
	for _, item := range aggregate.Items() {
		if _, err := t.queries.CreateItemPedido(ctx, db.CreateItemPedidoParams{PedidoID: row.ID, ProdutoID: item.ProductID(), Quantidade: int32(item.Quantity()), PrecoUnitario: item.Price()}); err != nil {
			return nil, translateError(err)
		}
	}
	items, err := t.queries.GetItensPedido(ctx, row.ID)
	if err != nil {
		return nil, translateError(err)
	}
	return orderRecordFromDB(row, items)
}
func (t *postgresOrderTransaction) UpdateStatus(ctx context.Context, id uuid.UUID, status order.Status) error {
	command, err := t.tx.Exec(ctx, "UPDATE pedidos SET status = $2 WHERE id = $1 AND status = 'PENDING'", id, string(status))
	if err != nil {
		return translateError(err)
	}
	if command.RowsAffected() == 0 {
		return ErrConflict
	}
	return nil
}

// clientFromDB converte a linha do banco (sqlc) na entidade de domínio Cliente.
func clientFromDB(row db.Cliente) *domain.Cliente {
	return &domain.Cliente{ID: row.ID, Name: row.Name, Email: row.Email, PasswordHash: row.PasswordHash, CreatedAt: row.CreatedAt.Time}
}

// productFromDB converte a linha do banco (sqlc) na entidade de domínio Produto.
func productFromDB(row db.Produto) *domain.Produto {
	return &domain.Produto{ID: row.ID, Nome: row.Nome, Preco: row.Preco, Estoque: int(row.Estoque)}
}

// orderRecordFromDB reconstrói o agregado Order a partir das linhas persistidas.
func orderRecordFromDB(row db.Pedido, persisted []db.ItensPedido) (*OrderRecord, error) {
	items := make([]order.OrderItem, 0, len(persisted))
	for _, item := range persisted {
		domainItem, err := order.NewOrderItem(item.ProdutoID, int(item.Quantidade), item.PrecoUnitario)
		if err != nil {
			return nil, err
		}
		items = append(items, domainItem)
	}
	return &OrderRecord{Order: order.Restore(row.ID, row.ClienteID, order.Status(row.Status), items), CreatedAt: row.CreatedAt.Time}, nil
}

// translateError converte erros do pgx/pgconn nos erros de domínio do repositório.
func translateError(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" {
		return ErrConflict
	}
	return err
}

// PaymentPostgresRepository é o adaptador PostgreSQL da idempotência de pagamentos.
type PaymentPostgresRepository struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

// NewPaymentPostgresRepository cria o repositório de idempotência de pagamentos.
func NewPaymentPostgresRepository(queries *db.Queries, pool *pgxpool.Pool) *PaymentPostgresRepository {
	return &PaymentPostgresRepository{queries, pool}
}

// IsProcessed verifica se o pedido já teve o pagamento processado anteriormente.
func (r *PaymentPostgresRepository) IsProcessed(ctx context.Context, orderID uuid.UUID) (bool, error) {
	_, err := r.queries.GetProcessedEvent(ctx, orderID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, translateError(err)
	}
	return true, nil
}

// MarkAsProcessed registra o pedido como processado (tabela de idempotência).
func (r *PaymentPostgresRepository) MarkAsProcessed(ctx context.Context, orderID uuid.UUID, sagaID uuid.UUID, status string) error {
	err := r.queries.CreateProcessedEvent(ctx, db.CreateProcessedEventParams{
		OrderID: orderID,
		SagaID:  sagaID,
		Status:  status,
	})
	if err != nil {
		return translateError(err)
	}
	return nil
}
