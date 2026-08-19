package payments_test

import (
	"context"
	"encoding/json"
	"errors"
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

type mockPaymentRepo struct {
	processed map[uuid.UUID]bool
}

func (m *mockPaymentRepo) IsProcessed(ctx context.Context, orderID uuid.UUID) (bool, error) {
	if m.processed == nil {
		return false, nil
	}
	return m.processed[orderID], nil
}

func (m *mockPaymentRepo) MarkAsProcessed(ctx context.Context, orderID uuid.UUID, sagaID uuid.UUID, status string) error {
	if m.processed == nil {
		m.processed = make(map[uuid.UUID]bool)
	}
	m.processed[orderID] = true
	return nil
}

func TestPaymentHandler_ProcessPayment(t *testing.T) {
	t.Run("Pagamento Aprovado (Valor Positivo)", func(t *testing.T) {
		pub := &mockPublisher{}
		svc := payments.NewPaymentService(pub)
		repo := &mockPaymentRepo{}
		handler := payments.NewPaymentHandler(svc, repo)

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

	t.Run("Pagamento Rejeitado (Valor Zero ou Negativo) - PaymentFailedEvent", func(t *testing.T) {
		pub := &mockPublisher{}
		svc := payments.NewPaymentService(pub)
		repo := &mockPaymentRepo{}
		handler := payments.NewPaymentHandler(svc, repo)

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

		if pub.publishedTopic != events.TopicPaymentFailed {
			t.Errorf("esperava tópico %s, obteve %s", events.TopicPaymentFailed, pub.publishedTopic)
		}

		failed, ok := pub.publishedEvent.(*events.PaymentFailedEvent)
		if !ok {
			t.Fatalf("esperava PaymentFailedEvent, obteve %T", pub.publishedEvent)
		}

		if failed.OrderID != orderEvent.OrderID {
			t.Errorf("esperava order_id %s, obteve %s", orderEvent.OrderID, failed.OrderID)
		}
	})

	t.Run("Idempotencia - Mensagem Duplicada (Persistente)", func(t *testing.T) {
		pub := &mockPublisher{}
		svc := payments.NewPaymentService(pub)
		repo := &mockPaymentRepo{}
		handler := payments.NewPaymentHandler(svc, repo)

		orderEvent := events.OrderCreatedEvent{
			SagaID:      uuid.New(),
			OrderID:     uuid.New(),
			ClientID:    uuid.New(),
			TotalAmount: 100.00,
			Status:      "PENDING",
			CreatedAt:   time.Now(),
		}

		body, _ := json.Marshal(orderEvent)

		// First message
		err := handler.HandleMessage(body)
		if err != nil {
			t.Fatalf("primeira mensagem falhou: %v", err)
		}

		// Duplicate message with same OrderID
		pub.publishedEvent = nil
		pub.publishedTopic = ""
		err = handler.HandleMessage(body)
		if err != nil {
			t.Fatalf("segunda mensagem (duplicada) retornou erro: %v", err)
		}

		// Should not publish again due to idempotency
		if pub.publishedEvent != nil {
			t.Errorf("esperava que o evento não fosse publicado novamente devido à idempotência")
		}
	})

	t.Run("Erro na publicação", func(t *testing.T) {
		pub := &mockPublisher{publishErr: errors.New("erro de conexao rabbitmq")}
		svc := payments.NewPaymentService(pub)
		repo := &mockPaymentRepo{}
		handler := payments.NewPaymentHandler(svc, repo)

		orderEvent := events.OrderCreatedEvent{
			SagaID:      uuid.New(),
			OrderID:     uuid.New(),
			ClientID:    uuid.New(),
			TotalAmount: 100.00,
			Status:      "PENDING",
			CreatedAt:   time.Now(),
		}

		body, _ := json.Marshal(orderEvent)
		err := handler.HandleMessage(body)
		if err == nil {
			t.Fatalf("esperava erro de publicação, obteve nil")
		}
	})
}
