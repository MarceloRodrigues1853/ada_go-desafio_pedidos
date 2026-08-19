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
	"pedidos/internal/repository"
)

type PaymentHandler struct {
	service         *PaymentService
	repo            repository.PaymentRepository
	processedOrders sync.Map
}

func NewPaymentHandler(service *PaymentService, repo repository.PaymentRepository) *PaymentHandler {
	return &PaymentHandler{
		service: service,
		repo:    repo,
	}
}

func (h *PaymentHandler) HandleMessage(body []byte) error {
	return h.HandleMessageWithContext(context.Background(), body)
}

func (h *PaymentHandler) HandleMessageWithContext(ctx context.Context, body []byte) error {
	var event events.OrderCreatedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		slog.ErrorContext(ctx, "Erro ao desserializar OrderCreatedEvent", "erro", err.Error())
		return fmt.Errorf("falha ao desserializar evento: %w", err)
	}

	if ctx == nil {
		ctx = context.Background()
	}
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

	slog.InfoContext(ctx, "Pagamento processado com sucesso",
		"saga_id", event.SagaID,
		"order_id", event.OrderID,
		"status", status,
	)

	return nil
}
