package payments

import (
	"context"
	"time"

	"pedidos/internal/events"
)

// PaymentService orquestra a lógica de negócio do processamento de pagamentos.
// Ele recebe um evento OrderCreatedEvent e decide se o pagamento foi aprovado
// ou recusado, publicando o resultado de volta nos tópicos da SAGA.
type PaymentService struct {
	publisher events.EventPublisher // Publicador de eventos no RabbitMQ
}

// NewPaymentService cria uma nova instância do serviço de pagamentos.
func NewPaymentService(publisher events.EventPublisher) *PaymentService {
	return &PaymentService{publisher: publisher}
}

// ProcessPayment analisa o evento de pedido criado e publica o resultado do pagamento.
// Retorna o evento gerado (PaymentProcessedEvent ou PaymentFailedEvent) para que o
// handler registre o status na tabela de idempotência.
func (s *PaymentService) ProcessPayment(ctx context.Context, orderCreated events.OrderCreatedEvent) (any, error) {
	// Regra de negócio: pedidos com valor menor ou igual a zero são recusados.
	if orderCreated.TotalAmount <= 0 {
		// Monta o evento de falha com o motivo da recusa.
		failedEvent := &events.PaymentFailedEvent{
			SagaID:      orderCreated.SagaID,
			OrderID:     orderCreated.OrderID,
			ClientID:    orderCreated.ClientID,
			TotalAmount: orderCreated.TotalAmount,
			Reason:      "Valor do pedido menor ou igual a zero",
			FailedAt:    time.Now(),
		}

		// Publica o evento de falha para disparar a compensação da SAGA no serviço de pedidos.
		if s.publisher != nil {
			if err := s.publisher.Publish(ctx, events.TopicPaymentFailed, failedEvent); err != nil {
				return nil, err
			}
		}

		return failedEvent, nil
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
