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

func (s *PaymentService) ProcessPayment(ctx context.Context, orderCreated events.OrderCreatedEvent) (any, error) {
	if orderCreated.TotalAmount <= 0 {
		failedEvent := &events.PaymentFailedEvent{
			SagaID:      orderCreated.SagaID,
			OrderID:     orderCreated.OrderID,
			ClientID:    orderCreated.ClientID,
			TotalAmount: orderCreated.TotalAmount,
			Reason:      "Valor do pedido menor ou igual a zero",
			FailedAt:    time.Now(),
		}

		if s.publisher != nil {
			if err := s.publisher.Publish(ctx, events.TopicPaymentFailed, failedEvent); err != nil {
				return nil, err
			}
		}

		return failedEvent, nil
	}

	processedEvent := &events.PaymentProcessedEvent{
		SagaID:      orderCreated.SagaID,
		OrderID:     orderCreated.OrderID,
		ClientID:    orderCreated.ClientID,
		TotalAmount: orderCreated.TotalAmount,
		Status:      "PAID",
		ProcessedAt: time.Now(),
	}

	if s.publisher != nil {
		if err := s.publisher.Publish(ctx, events.TopicPaymentProcessed, processedEvent); err != nil {
			return nil, err
		}
	}

	return processedEvent, nil
}
