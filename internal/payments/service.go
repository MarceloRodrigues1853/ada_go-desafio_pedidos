package payments

import (
	"context"
	"time"

	"pedidos/internal/events"
)

type PaymentService struct {
	publisher events.EventPublisher
}

func NewPaymentService(publisher events.EventPublisher) *PaymentService {
	return &PaymentService{publisher: publisher}
}

func (s *PaymentService) ProcessPayment(ctx context.Context, orderCreated events.OrderCreatedEvent) (*events.PaymentProcessedEvent, error) {
	status := "PAID"
	if orderCreated.TotalAmount <= 0 {
		status = "FAILED"
	}

	processedEvent := &events.PaymentProcessedEvent{
		SagaID:      orderCreated.SagaID,
		OrderID:     orderCreated.OrderID,
		ClientID:    orderCreated.ClientID,
		TotalAmount: orderCreated.TotalAmount,
		Status:      status,
		ProcessedAt: time.Now(),
	}

	if s.publisher != nil {
		if err := s.publisher.Publish(ctx, events.TopicPaymentProcessed, processedEvent); err != nil {
			return nil, err
		}
	}

	return processedEvent, nil
}
