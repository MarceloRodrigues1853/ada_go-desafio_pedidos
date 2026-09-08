package payments

import (
	"context"
	"errors"
	"testing"

	"pedidos/internal/events"

	"github.com/google/uuid"
)

type recordingPublisher struct {
	topic string
	event any
}

func (p *recordingPublisher) Publish(_ context.Context, topic string, event any) error {
	p.topic = topic
	p.event = event
	return nil
}

type failingGateway struct {
	err error
}

func (g failingGateway) Process(context.Context, PaymentInput) (PaymentResult, error) {
	return PaymentResult{}, g.err
}

func TestPaymentServicePublishesProcessedWhenGatewayApproves(t *testing.T) {
	publisher := &recordingPublisher{}
	service := NewPaymentServiceWithGateway(
		publisher,
		NewFakePaymentGateway(PaymentOutcomeApproved, ""),
		PaymentMethodPix,
	)
	event := newOrderCreatedEvent(120)

	result, err := service.ProcessPayment(context.Background(), event)
	if err != nil {
		t.Fatalf("ProcessPayment() retornou erro: %v", err)
	}
	if publisher.topic != events.TopicPaymentProcessed {
		t.Errorf("tópico = %q; esperado %q", publisher.topic, events.TopicPaymentProcessed)
	}
	processed, ok := result.(*events.PaymentProcessedEvent)
	if !ok {
		t.Fatalf("resultado = %T; esperado *events.PaymentProcessedEvent", result)
	}
	if processed.Status != "PAID" {
		t.Errorf("status = %q; esperado PAID", processed.Status)
	}
}

func TestPaymentServicePublishesFailedWhenGatewayDeclines(t *testing.T) {
	publisher := &recordingPublisher{}
	service := NewPaymentService(publisher)
	event := newOrderCreatedEvent(220)
	event.PaymentMethod = string(PaymentMethodBoleto)
	event.SimulationOutcome = string(PaymentOutcomeDeclined)

	result, err := service.ProcessPayment(context.Background(), event)
	if err != nil {
		t.Fatalf("ProcessPayment() retornou erro: %v", err)
	}
	if publisher.topic != events.TopicPaymentFailed {
		t.Errorf("tópico = %q; esperado %q", publisher.topic, events.TopicPaymentFailed)
	}
	failed, ok := result.(*events.PaymentFailedEvent)
	if !ok {
		t.Fatalf("resultado = %T; esperado *events.PaymentFailedEvent", result)
	}
	if failed.Reason != "Pagamento recusado pelo gateway" {
		t.Errorf("motivo = %q; esperado motivo padrão da recusa", failed.Reason)
	}
}

func TestPaymentServicePreservesInvalidAmountAsPaymentFailure(t *testing.T) {
	publisher := &recordingPublisher{}
	service := NewPaymentService(publisher)

	result, err := service.ProcessPayment(context.Background(), newOrderCreatedEvent(0))
	if err != nil {
		t.Fatalf("ProcessPayment() retornou erro: %v", err)
	}
	if publisher.topic != events.TopicPaymentFailed {
		t.Errorf("tópico = %q; esperado %q", publisher.topic, events.TopicPaymentFailed)
	}
	if _, ok := result.(*events.PaymentFailedEvent); !ok {
		t.Fatalf("resultado = %T; esperado *events.PaymentFailedEvent", result)
	}
}

func TestPaymentServicePropagatesGatewayError(t *testing.T) {
	expected := errors.New("gateway indisponível")
	service := NewPaymentServiceWithGateway(nil, failingGateway{err: expected}, PaymentMethodCard)

	_, err := service.ProcessPayment(context.Background(), newOrderCreatedEvent(90))
	if !errors.Is(err, expected) {
		t.Fatalf("erro = %v; esperado %v", err, expected)
	}
}

func newOrderCreatedEvent(amount float64) events.OrderCreatedEvent {
	return events.OrderCreatedEvent{
		SagaID:      uuid.New(),
		OrderID:     uuid.New(),
		ClientID:    uuid.New(),
		TotalAmount: amount,
		Status:      "PENDING",
	}
}
