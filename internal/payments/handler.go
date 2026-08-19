package payments

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	"pedidos/internal/events"
)

type PaymentHandler struct {
	service *PaymentService
}

func NewPaymentHandler(service *PaymentService) *PaymentHandler {
	return &PaymentHandler{service: service}
}

func (h *PaymentHandler) HandleMessage(body []byte) error {
	var event events.OrderCreatedEvent
	if err := json.Unmarshal(body, &event); err != nil {
		slog.Error("Erro ao desserializar OrderCreatedEvent", "erro", err.Error())
		return fmt.Errorf("falha ao desserializar evento: %w", err)
	}

	ctx := context.Background()
	processed, err := h.service.ProcessPayment(ctx, event)
	if err != nil {
		slog.Error("Erro ao processar pagamento",
			"saga_id", event.SagaID,
			"order_id", event.OrderID,
			"erro", err.Error(),
		)
		return err
	}

	slog.Info("Pagamento processado com sucesso",
		"saga_id", processed.SagaID,
		"order_id", processed.OrderID,
		"status", processed.Status,
	)

	return nil
}
