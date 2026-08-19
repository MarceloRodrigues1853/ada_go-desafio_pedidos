package service

import (
	"context"
	"log/slog"
	"time"

	"pedidos/internal/domain"
	"pedidos/internal/domain/order"
	"pedidos/internal/events"
	"pedidos/internal/infra/metrics"
	"pedidos/internal/repository"

	"github.com/google/uuid"
)

// OrderItemInput é a entrada de um item recebida pelo caso de uso de criação de pedido.
type OrderItemInput struct {
	ProductID string
	Quantity  int
}

// OrderOutput é a representação do pedido devolvida pela API.
type OrderOutput struct {
	ID        uuid.UUID `json:"id"`
	ClienteID uuid.UUID `json:"cliente_id"`
	Status    string    `json:"status"`
	CreatedAt time.Time `json:"created_at"`
}

// OrderServiceInterface define o contrato dos casos de uso de pedidos.
// É implementada pelo OrderService e pelo decorator LoggingOrderService.
type OrderServiceInterface interface {
	Create(context.Context, uuid.UUID, []OrderItemInput) (*OrderOutput, error)
	GetByID(context.Context, uuid.UUID) (*OrderOutput, error)
	ListPaginado(context.Context, int32, int32) ([]OrderOutput, error)
	Pay(context.Context, uuid.UUID) error
	Cancel(context.Context, uuid.UUID) error
	ProcessPaymentResult(context.Context, events.PaymentProcessedEvent) error
	ProcessPaymentFailure(context.Context, events.PaymentFailedEvent) error
}

// OrderService implementa os casos de uso de pedidos orquestrando as portas
// de repositório (clientes, produtos e pedidos) e publicando eventos da SAGA.
type OrderService struct {
	clients        repository.ClientRepository  // Porta de persistência de clientes
	products       repository.ProductRepository // Porta de persistência de produtos
	orders         repository.OrderRepository   // Porta de persistência de pedidos
	eventPublisher events.EventPublisher        // Publicador de eventos (RabbitMQ)
}

// NewOrderService cria uma nova instância do serviço de pedidos.
func NewOrderService(clients repository.ClientRepository, products repository.ProductRepository, orders repository.OrderRepository, eventPublisher events.EventPublisher) *OrderService {
	return &OrderService{clients, products, orders, eventPublisher}
}

// Create valida o cliente e os produtos, monta o agregado Order e persiste
// o pedido reservando o estoque dentro de uma única transação. Ao final,
// publica o evento order.created para iniciar a SAGA de pagamento.
func (s *OrderService) Create(ctx context.Context, clientID uuid.UUID, inputs []OrderItemInput) (*OrderOutput, error) {
	start := time.Now()
	defer func() {
		metrics.ObserveOrderProcessing(time.Since(start).Seconds())
	}()

	// 1. O cliente precisa existir.
	if _, err := s.clients.GetByID(ctx, clientID); err != nil {
		return nil, err
	}

	// 2. Converte cada input em um OrderItem usando o preço atual do catálogo.
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

	// 3. Cria o agregado (invariantes de status e itens são validadas aqui).
	aggregate, err := order.NewOrder(clientID, items)
	if err != nil {
		return nil, err
	}

	// 4. Persiste pedido + reserva de estoque em uma única transação.
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

	// 5. Publica o evento para o microsserviço de pagamentos (início da SAGA).
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

	metrics.IncOrdersCreated()
	return orderOutput(record), nil
}

// GetByID busca um pedido pelo seu identificador.
func (s *OrderService) GetByID(ctx context.Context, id uuid.UUID) (*OrderOutput, error) {
	record, err := s.orders.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	return orderOutput(record), nil
}

// ListPaginado lista pedidos com paginação (limit/offset).
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

// Pay marca um pedido PENDING como PAID.
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

// Cancel cancela um pedido PENDING e devolve o estoque na mesma transação.
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

// ProcessPaymentResult é o callback da SAGA quando o pagamento é processado:
//   - PAID: marca o pedido como pago e encerra a SAGA;
//   - caso contrário (rejeitado): executa a compensação cancelando o pedido
//     e estornando o estoque.
func (s *OrderService) ProcessPaymentResult(ctx context.Context, event events.PaymentProcessedEvent) error {
	record, err := s.orders.GetByID(ctx, event.OrderID)
	if err != nil {
		slog.Error("Pedido não encontrado para processar resultado de pagamento",
			"saga_id", event.SagaID,
			"order_id", event.OrderID,
			"erro", err.Error(),
		)
		return err
	}

	if event.Status == "PAID" {
		if err := record.Order.Pay(); err != nil {
			slog.Warn("Tentativa de pagar pedido com status inválido",
				"saga_id", event.SagaID,
				"order_id", event.OrderID,
				"status_atual", record.Order.Status(),
				"erro", err.Error(),
			)
			return err
		}
		if err := s.orders.UpdateStatus(ctx, event.OrderID, record.Order.Status()); err != nil {
			return err
		}
		slog.Info("Saga concluída: Pedido pago com sucesso",
			"saga_id", event.SagaID,
			"order_id", event.OrderID,
			"status", record.Order.Status(),
		)
		return nil
	}

	// Caso FAILED ou rejeitado -> compensação (cancelar pedido + devolver estoque)
	if err := record.Order.Cancel(); err != nil {
		slog.Warn("Tentativa de cancelar pedido em estado inválido na compensação",
			"saga_id", event.SagaID,
			"order_id", event.OrderID,
			"status_atual", record.Order.Status(),
			"erro", err.Error(),
		)
	}

	return s.orders.WithinTransaction(ctx, func(tx repository.OrderTransaction) error {
		for _, item := range record.Order.Items() {
			if err := tx.ReturnStock(ctx, item.ProductID(), item.Quantity()); err != nil {
				return err
			}
		}
		if err := tx.UpdateStatus(ctx, event.OrderID, record.Order.Status()); err != nil {
			return err
		}
		slog.Info("Saga compensada: Pagamento falhou, pedido cancelado e estoque estornado",
			"saga_id", event.SagaID,
			"order_id", event.OrderID,
			"status", record.Order.Status(),
		)
		return nil
	})
}

// ProcessPaymentFailure é o callback da SAGA quando o pagamento falha.
// Executa a compensação: cancela o pedido e estorna o estoque.
func (s *OrderService) ProcessPaymentFailure(ctx context.Context, event events.PaymentFailedEvent) error {
	record, err := s.orders.GetByID(ctx, event.OrderID)
	if err != nil {
		slog.ErrorContext(ctx, "Pedido não encontrado para processar falha de pagamento",
			"saga_id", event.SagaID,
			"order_id", event.OrderID,
			"erro", err.Error(),
		)
		return err
	}

	if err := record.Order.Cancel(); err != nil {
		slog.WarnContext(ctx, "Tentativa de cancelar pedido em estado inválido na compensação",
			"saga_id", event.SagaID,
			"order_id", event.OrderID,
			"status_atual", record.Order.Status(),
			"erro", err.Error(),
		)
	}

	return s.orders.WithinTransaction(ctx, func(tx repository.OrderTransaction) error {
		for _, item := range record.Order.Items() {
			if err := tx.ReturnStock(ctx, item.ProductID(), item.Quantity()); err != nil {
				return err
			}
		}
		if err := tx.UpdateStatus(ctx, event.OrderID, record.Order.Status()); err != nil {
			return err
		}
		slog.InfoContext(ctx, "Saga compensada: Pagamento falhou, pedido cancelado e estoque estornado",
			"saga_id", event.SagaID,
			"order_id", event.OrderID,
			"status", record.Order.Status(),
			"reason", event.Reason,
		)
		return nil
	})
}

// orderOutput converte o registro persistido em uma resposta de API.
func orderOutput(record *repository.OrderRecord) *OrderOutput {
	return &OrderOutput{ID: record.Order.ID(), ClienteID: record.Order.ClientID(), Status: string(record.Order.Status()), CreatedAt: record.CreatedAt}
}

// ClientService implementa os casos de uso da entidade cliente.
type ClientService struct{ repository repository.ClientRepository }

// NewClientService cria uma nova instância do serviço de clientes.
func NewClientService(repository repository.ClientRepository) *ClientService {
	return &ClientService{repository}
}

// Create valida os dados e cria um novo cliente (senha é criptografada no domínio).
func (s *ClientService) Create(ctx context.Context, name, email, password string) (*domain.Cliente, error) {
	client, err := domain.NovoCliente(name, email, password)
	if err != nil {
		return nil, err
	}
	return s.repository.Create(ctx, client)
}

// List retorna todos os clientes cadastrados.
func (s *ClientService) List(ctx context.Context) ([]domain.Cliente, error) {
	return s.repository.List(ctx)
}

// GetByID busca um cliente pelo seu identificador.
func (s *ClientService) GetByID(ctx context.Context, id uuid.UUID) (*domain.Cliente, error) {
	return s.repository.GetByID(ctx, id)
}

// ProductService implementa os casos de uso da entidade produto.
type ProductService struct{ repository repository.ProductRepository }

// NewProductService cria uma nova instância do serviço de produtos.
func NewProductService(repository repository.ProductRepository) *ProductService {
	return &ProductService{repository}
}

// Create valida os dados e cria um novo produto.
func (s *ProductService) Create(ctx context.Context, id, name string, price float64, stock int) (*domain.Produto, error) {
	product, err := domain.NovoProduto(id, name, price, stock)
	if err != nil {
		return nil, err
	}
	return s.repository.Create(ctx, product)
}

// List retorna todos os produtos cadastrados.
func (s *ProductService) List(ctx context.Context) ([]domain.Produto, error) {
	return s.repository.List(ctx)
}

// GetByID busca um produto pelo seu identificador (SKU).
func (s *ProductService) GetByID(ctx context.Context, id string) (*domain.Produto, error) {
	if id == "" {
		return nil, domain.ErrProdutoInvalido
	}
	return s.repository.GetByID(ctx, id)
}
