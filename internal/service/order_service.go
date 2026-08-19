package service

import (
	"context"
	"time"

	"pedidos/internal/domain"
	"pedidos/internal/domain/order"
	"pedidos/internal/events"
	"pedidos/internal/repository"

	"github.com/google/uuid"
)

type OrderItemInput struct {
	ProductID string
	Quantity  int
}
type OrderOutput struct {
	ID        uuid.UUID `json:"id"`
	ClienteID uuid.UUID `json:"cliente_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}
type OrderServiceInterface interface {
	Create(context.Context, uuid.UUID, []OrderItemInput) (*OrderOutput, error)
	GetByID(context.Context, uuid.UUID) (*OrderOutput, error)
	ListPaginado(context.Context, int32, int32) ([]OrderOutput, error)
	Pay(context.Context, uuid.UUID) error
	Cancel(context.Context, uuid.UUID) error
}

type OrderService struct {
	clients        repository.ClientRepository
	products       repository.ProductRepository
	orders         repository.OrderRepository
	eventPublisher events.EventPublisher
}

func NewOrderService(clients repository.ClientRepository, products repository.ProductRepository, orders repository.OrderRepository, eventPublisher events.EventPublisher) *OrderService {
	return &OrderService{clients, products, orders, eventPublisher}
}

func (s *OrderService) Create(ctx context.Context, clientID uuid.UUID, inputs []OrderItemInput) (*OrderOutput, error) {
	if _, err := s.clients.GetByID(ctx, clientID); err != nil {
		return nil, err
	}
	items := make([]order.OrderItem, 0, len(inputs))
	for _, input := range inputs {
		product, err := s.products.GetByID(ctx, input.ProductID)
		if err != nil {
			return nil, err
		}
		item, err := order.NewOrderItem(product.ID, input.Quantity, product.Preco)
		if err != nil {
			return nil, err
		}
		items = append(items, item)
	}
	aggregate, err := order.NewOrder(clientID, items)
	if err != nil {
		return nil, err
	}
	var record *repository.OrderRecord
	err = s.orders.WithinTransaction(ctx, func(tx repository.OrderTransaction) error {
		for _, item := range aggregate.Items() {
			if err := tx.ReserveStock(ctx, item.ProductID(), item.Quantity()); err != nil {
				return err
			}
		}
		created, err := tx.Create(ctx, aggregate)
		if err != nil {
			return err
		}
		record = created
		return nil
	})
	if err != nil {
		return nil, err
	}

	if s.eventPublisher != nil {
		event := events.OrderCreatedEvent{
			SagaID:      uuid.New(),
			OrderID:     record.Order.ID(),
			ClientID:    record.Order.ClientID(),
			TotalAmount: record.Order.CalculateTotal(),
			Status:      string(record.Order.Status()),
			CreatedAt:   record.CreatedAt,
		}
		if err := s.eventPublisher.Publish(ctx, events.TopicOrderCreated, event); err != nil {
			return nil, err
		}
	}

	return orderOutput(record), nil
}
func (s *OrderService) GetByID(ctx context.Context, id uuid.UUID) (*OrderOutput, error) {
	record, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return orderOutput(record), nil
}
func (s *OrderService) ListPaginado(ctx context.Context, limit, offset int32) ([]OrderOutput, error) {
	records, err := s.orders.List(ctx, limit, offset)
	if err != nil {
		return nil, err
	}
	output := make([]OrderOutput, 0, len(records))
	for i := range records {
		output = append(output, *orderOutput(&records[i]))
	}
	return output, nil
}
func (s *OrderService) Pay(ctx context.Context, id uuid.UUID) error {
	record, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := record.Order.Pay(); err != nil {
		return err
	}
	return s.orders.UpdateStatus(ctx, id, record.Order.Status())
}
func (s *OrderService) Cancel(ctx context.Context, id uuid.UUID) error {
	record, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := record.Order.Cancel(); err != nil {
		return err
	}
	return s.orders.WithinTransaction(ctx, func(tx repository.OrderTransaction) error {
		for _, item := range record.Order.Items() {
			if err := tx.ReturnStock(ctx, item.ProductID(), item.Quantity()); err != nil {
				return err
			}
		}
		return tx.UpdateStatus(ctx, id, record.Order.Status())
	})
}
func orderOutput(record *repository.OrderRecord) *OrderOutput {
	return &OrderOutput{ID: record.Order.ID(), ClienteID: record.Order.ClientID(), Status: string(record.Order.Status()), CreatedAt: record.CreatedAt}
}

type ClientService struct{ repository repository.ClientRepository }

func NewClientService(repository repository.ClientRepository) *ClientService {
	return &ClientService{repository}
}
func (s *ClientService) Create(ctx context.Context, name, email, password string) (*domain.Cliente, error) {
	client, err := domain.NovoCliente(name, email, password)
	if err != nil {
		return nil, err
	}
	return s.repository.Create(ctx, client)
}
func (s *ClientService) List(ctx context.Context) ([]domain.Cliente, error) {
	return s.repository.List(ctx)
}
func (s *ClientService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cliente, error) {
	return s.repository.GetByID(ctx, id)
}

type ProductService struct{ repository repository.ProductRepository }

func NewProductService(repository repository.ProductRepository) *ProductService {
	return &ProductService{repository}
}
func (s *ProductService) Create(ctx context.Context, id, name string, price float64, stock int) (*domain.Produto, error) {
	product, err := domain.NovoProduto(id, name, price, stock)
	if err != nil {
		return nil, err
	}
	return s.repository.Create(ctx, product)
}
func (s *ProductService) List(ctx context.Context) ([]domain.Produto, error) {
	return s.repository.List(ctx)
}
func (s *ProductService) GetByID(ctx context.Context, id string) (*domain.Produto, error) {
	if id == "" {
		return nil, domain.ErrProdutoInvalido
	}
	return s.repository.GetByID(ctx, id)
}
