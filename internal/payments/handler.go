package payments

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"sync"

	"pedidos/internal/events"
	internalLogger "pedidos/internal/infra/logger"
	"pedidos/internal/infra/metrics"
	"pedidos/internal/repository"
)

// PaymentHandler é o consumidor da fila order.created.
// Ele desserializa o evento, aplica idempotência (persistente ou em memória),
// processa o pagamento via PaymentService e registra o resultado.
type PaymentHandler struct {
	service         *PaymentService
	repo            repository.PaymentRepository
	processedOrders sync.Map // Fallback de idempotência em memória quando não há banco
}

// NewPaymentHandler cria um novo handler de pagamentos.
func NewPaymentHandler(service *PaymentService, repo repository.PaymentRepository) *PaymentHandler {
	return &PaymentHandler{
		service: service,
		repo:    repo,
	}
}

// HandleMessage processa uma mensagem recebida do RabbitMQ usando um contexto de fundo.
func (h *PaymentHandler) HandleMessage(body []byte) error {
	return h.HandleMessageWithContext(context.Background(), body)
}

// HandleMessageWithContext processa uma mensagem da fila order.created:
//  1. Desserializa o evento
//  2. Verifica idempotência (mesmo pedido não pode ser pago duas vezes)
//  3. Processa o pagamento
//  4. Marca o evento como processado
func (h *PaymentHandler) HandleMessageWithContext(ctx context.Context, body []byte) error {
	var event events.OrderCreatedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		slog.ErrorContext(ctx, "Erro ao desserializar OrderCreatedEvent", "erro", err.Error())
		return fmt.Errorf("falha ao desserializar evento: %w", err)
	}

	if ctx == nil {
		ctx = context.Background()
	}
	// Propaga o saga_id para todos os logs gerados a partir daqui.
	ctx = internalLogger.WithSagaID(ctx, event.SagaID)

	// Idempotência persistente ou em memória (fallback)
	if h.repo != nil {
		processed, err := h.repo.IsProcessed(ctx, event.OrderID)
		if err != nil {
			slog.ErrorContext(ctx, "Erro ao verificar idempotência no banco",
				"saga_id", event.SagaID,
				"order_id", event.OrderID,
				"erro", err.Error(),
			)
			return err
		}
		if processed {
			slog.WarnContext(ctx, "Pedido já processado anteriormente (idempotência)",
				"saga_id", event.SagaID,
				"order_id", event.OrderID,
			)
			return nil
		}
	} else {
		if _, loaded := h.processedOrders.LoadOrStore(event.OrderID, true); loaded {
			slog.WarnContext(ctx, "Pedido já processado anteriormente (idempotência)",
				"saga_id", event.SagaID,
				"order_id", event.OrderID,
			)
			return nil
		}
	}

	result, err := h.service.ProcessPayment(ctx, event)
	if err != nil {
		slog.ErrorContext(ctx, "Erro ao processar pagamento",
			"saga_id", event.SagaID,
			"order_id", event.OrderID,
			"erro", err.Error(),
		)
		return err
	}

	var status string
	switch ev := result.(type) {
	case *events.PaymentProcessedEvent:
		status = ev.Status
	case *events.PaymentFailedEvent:
		status = "FAILED"
	}

	// Incrementa a métrica Prometheus de pagamentos processados por status.
	metrics.IncPaymentsProcessed(status)

	// Registra o pedido como processado (tabela de idempotência), se o repositório existir.
	if h.repo != nil {
		if err := h.repo.MarkAsProcessed(ctx, event.OrderID, event.SagaID, status); err != nil {
			if !errors.Is(err, repository.ErrConflict) {
				slog.ErrorContext(ctx, "Erro ao marcar evento como processado",
					"saga_id", event.SagaID,
					"order_id", event.OrderID,
					"erro", err.Error(),
				)
				return err
			}
		}
	}

	// Log de conclusão: pagamento processado com sucesso.
	slog.InfoContext(ctx, "Pagamento processado com sucesso",
		"saga_id", event.SagaID,
		"order_id", event.OrderID,
		"status", status,
	)

	return nil
}
