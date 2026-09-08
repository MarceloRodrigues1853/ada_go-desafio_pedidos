package payments

import (
	"context"
	"errors"
	"time"

	"pedidos/internal/events"
)

// PaymentService orquestra a lógica de negócio do processamento de pagamentos.
// Ele recebe um evento OrderCreatedEvent e decide se o pagamento foi aprovado
// ou recusado, publicando o resultado de volta nos tópicos da SAGA.
type PaymentService struct {
	publisher events.EventPublisher // Publicador de eventos no RabbitMQ
	gateway   PaymentGateway
	method    PaymentMethod
}

// NewPaymentService preserva o comportamento atual da aplicação usando um
// gateway falso aprovado e cartão como método temporário padrão.
func NewPaymentService(publisher events.EventPublisher) *PaymentService {
	return NewPaymentServiceWithGateway(
		publisher,
		NewFakePaymentGateway(PaymentOutcomeApproved, ""),
		PaymentMethodCard,
	)
}

// NewPaymentServiceWithGateway permite injetar a decisão e o método usados no
// processamento. A seleção por pedido será conectada ao contrato da API em uma
// etapa posterior.
func NewPaymentServiceWithGateway(publisher events.EventPublisher, gateway PaymentGateway, method PaymentMethod) *PaymentService {
	return &PaymentService{
		publisher: publisher,
		gateway:   gateway,
		method:    method,
	}
}

// ProcessPayment analisa o evento de pedido criado e publica o resultado do pagamento.
// Retorna o evento gerado (PaymentProcessedEvent ou PaymentFailedEvent) para que o
// handler registre o status na tabela de idempotência.
func (s *PaymentService) ProcessPayment(ctx context.Context, orderCreated events.OrderCreatedEvent) (any, error) {
	if s.gateway == nil {
		return nil, ErrPaymentGatewayRequired
	}

	method := s.method
	if orderCreated.PaymentMethod != "" {
		method = PaymentMethod(orderCreated.PaymentMethod)
	}

	result, err := s.gateway.Process(ctx, PaymentInput{
		Method:            method,
		Amount:            orderCreated.TotalAmount,
		SimulationOutcome: PaymentOutcome(orderCreated.SimulationOutcome),
	})
	if err != nil {
		if errors.Is(err, ErrInvalidPaymentAmount) {
			return s.publishFailure(ctx, orderCreated, err.Error())
		}
		return nil, err
	}

	if result.Outcome == PaymentOutcomeDeclined {
		reason := result.Reason
		if reason == "" {
			reason = "Pagamento recusado pelo gateway"
		}
		return s.publishFailure(ctx, orderCreated, reason)
	}

	// Caminho feliz: o pagamento foi aprovado.
	processedEvent := &events.PaymentProcessedEvent{
		SagaID:      orderCreated.SagaID,
		OrderID:     orderCreated.OrderID,
		ClientID:    orderCreated.ClientID,
		TotalAmount: orderCreated.TotalAmount,
		Status:      "PAID",
		ProcessedAt: time.Now(),
	}

	// Publica o evento de aprovação para o serviço de pedidos concluir a SAGA.
	if s.publisher != nil {
		if err := s.publisher.Publish(ctx, events.TopicPaymentProcessed, processedEvent); err != nil {
			return nil, err
		}
	}

	return processedEvent, nil
}

func (s *PaymentService) publishFailure(ctx context.Context, orderCreated events.OrderCreatedEvent, reason string) (any, error) {
	failedEvent := &events.PaymentFailedEvent{
		SagaID:      orderCreated.SagaID,
		OrderID:     orderCreated.OrderID,
		ClientID:    orderCreated.ClientID,
		TotalAmount: orderCreated.TotalAmount,
		Reason:      reason,
		FailedAt:    time.Now(),
	}

	if s.publisher != nil {
		if err := s.publisher.Publish(ctx, events.TopicPaymentFailed, failedEvent); err != nil {
			return nil, err
		}
	}

	return failedEvent, nil
}
