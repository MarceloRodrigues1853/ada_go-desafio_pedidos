package events

import (
	"context"
	"time"

	"github.com/google/uuid"
)

const (
	TopicOrderCreated     = "order.created"
	TopicPaymentProcessed = "payment.processed"
	TopicPaymentFailed    = "payment.failed"
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

// PaymentProcessedEvent é emitido quando o serviço de pagamentos finaliza a tentativa de cobrança (sucesso ou falha).
type PaymentProcessedEvent struct {
	SagaID      uuid.UUID `json:"saga_id"`
	OrderID     uuid.UUID `json:"order_id"`
	ClientID    uuid.UUID `json:"client_id"`
	TotalAmount float64   `json:"total_amount"`
	Status      string    `json:"status"`
	ProcessedAt time.Time `json:"processed_at"`
}

// EventPublisher define o contrato para publicação de eventos
type EventPublisher interface {
	Publish(ctx context.Context, topic string, event any) error
}
