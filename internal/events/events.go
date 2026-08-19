package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// Tópicos (filas) do RabbitMQ usados na orquestração da SAGA de pagamentos.
const (
	TopicOrderCreated     = "order.created"     // Pedido criado -> aguardando pagamento
	TopicPaymentProcessed = "payment.processed" // Pagamento aprovado -> concluir SAGA
	TopicPaymentFailed    = "payment.failed"    // Pagamento recusado -> compensação da SAGA
)

// OrderCreatedEvent é emitido quando um pedido é criado com sucesso e aguarda processamento de pagamento.
type OrderCreatedEvent struct {
	SagaID      uuid.UUID `json:"saga_id"`
	OrderID     uuid.UUID `json:"order_id"`
	ClientID    uuid.UUID `json:"client_id"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	CreatedAt   time.Time `json:"created_at"`
}

// PaymentProcessedEvent é emitido quando o serviço de pagamentos finaliza a tentativa de cobrança com sucesso.
type PaymentProcessedEvent struct {
	SagaID      uuid.UUID `json:"saga_id"`
	OrderID     uuid.UUID `json:"order_id"`
	ClientID    uuid.UUID `json:"client_id"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	ProcessedAt time.Time `json:"processed_at"`
}

// PaymentFailedEvent é emitido quando o serviço de pagamentos falha ao processar o pagamento.
type PaymentFailedEvent struct {
	SagaID      uuid.UUID `json:"saga_id"`
	OrderID     uuid.UUID `json:"order_id"`
	ClientID    uuid.UUID `json:"client_id"`
	TotalAmount float64   `json:"total_amount"`
	Reason      string    `json:"reason"`
	FailedAt    time.Time `json:"failed_at"`
}

// EventPublisher define o contrato para publicação de eventos
type EventPublisher interface {
	Publish(ctx context.Context, topic string, event any) error
}
