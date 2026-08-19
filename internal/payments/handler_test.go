package payments_test

import (
	"context"
	"encoding/json"
	"testing"
	"time"

	"pedidos/internal/events"
	"pedidos/internal/payments"

	"github.com/google/uuid"
)

type mockPublisher struct {
	publishedTopic string
	publishedEvent any
	publishErr     error
}

func (m *mockPublisher) Publish(ctx context.Context, topic string, event any) error {
	m.publishedTopic = topic
	m.publishedEvent = event
	return m.publishErr
}

func TestPaymentHandler_ProcessPayment(t *testing.T) {
	t.Run("Pagamento Aprovado (Valor Positivo)", func(t *testing.T) {
		pub := &mockPublisher{}
		svc := payments.NewPaymentService(pub)
		handler := payments.NewPaymentHandler(svc)

		orderEvent := events.OrderCreatedEvent{
			SagaID:      uuid.New(),
			OrderID:     uuid.New(),
			ClientID:    uuid.New(),
			TotalAmount: 150.00,
			Status:      "PENDING",
			CreatedAt:   time.Now(),
		}

		body, _ := json.Marshal(orderEvent)
		err := handler.HandleMessage(body)
		if err != nil {
			t.Fatalf("esperava sucesso, obteve erro: %v", err)
		}

		if pub.publishedTopic != events.TopicPaymentProcessed {
			t.Errorf("esperava tópico %s, obteve %s", events.TopicPaymentProcessed, pub.publishedTopic)
		}

		processed, ok := pub.publishedEvent.(*events.PaymentProcessedEvent)
		if !ok {
			t.Fatalf("esperava PaymentProcessedEvent, obteve %T", pub.publishedEvent)
		}

		if processed.Status != "PAID" {
			t.Errorf("esperava status PAID, obteve %s", processed.Status)
		}
		if processed.SagaID != orderEvent.SagaID {
			t.Errorf("esperava saga_id %s, obteve %s", orderEvent.SagaID, processed.SagaID)
		}
	})

	t.Run("Pagamento Rejeitado (Valor Zero ou Negativo)", func(t *testing.T) {
		pub := &mockPublisher{}
		svc := payments.NewPaymentService(pub)
		handler := payments.NewPaymentHandler(svc)

		orderEvent := events.OrderCreatedEvent{
			SagaID:      uuid.New(),
			OrderID:     uuid.New(),
			ClientID:    uuid.New(),
			TotalAmount: 0.00,
			Status:      "PENDING",
			CreatedAt:   time.Now(),
		}

		body, _ := json.Marshal(orderEvent)
		err := handler.HandleMessage(body)
		if err != nil {
			t.Fatalf("esperava sucesso ao processar rejeição, obteve erro: %v", err)
		}

		processed, ok := pub.publishedEvent.(*events.PaymentProcessedEvent)
		if !ok {
			t.Fatalf("esperava PaymentProcessedEvent, obteve %T", pub.publishedEvent)
		}

		if processed.Status != "FAILED" {
			t.Errorf("esperava status FAILED, obteve %s", processed.Status)
		}
	})
}
