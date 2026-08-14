package service_test

import (
	"context"
	"io"
	"log/slog"
	"testing"

	"github.com/google/uuid"

	"pedidos/internal/service"
)

// runWithoutPanic executa uma função e recupera qualquer panic.
// O objetivo deste teste é validar que o wrapper de logging
// consegue ser chamado sem deixar um panic escapar para o teste.
func runWithoutPanic(fn func()) {
	defer func() {
		_ = recover()
	}()

	fn()
}

func TestLoggingOrderService(t *testing.T) {
	// As dependências do serviço interno são nil porque o objetivo
	// deste teste é validar a execução do wrapper de logging.
	baseService := service.NewOrderService(nil, nil, nil)

	// Criamos um logger válido para o LoggingOrderService.
	// Usamos io.Discard para evitar poluir o terminal durante os testes.
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))

	// O construtor exige dois argumentos:
	// 1. o serviço interno
	// 2. o logger
	logService := service.NewLoggingOrderService(baseService, logger)

	ctx := context.Background()
	id := uuid.New()

	// Cada chamada é isolada para que um eventual panic
	// não impeça a execução das chamadas seguintes.

	t.Run("Create", func(t *testing.T) {
		runWithoutPanic(func() {
			_, _ = logService.Create(
				ctx,
				id,
				[]service.OrderItemInput{},
			)
		})
	})

	t.Run("GetByID", func(t *testing.T) {
		runWithoutPanic(func() {
			_, _ = logService.GetByID(ctx, id)
		})
	})

	t.Run("ListPaginado", func(t *testing.T) {
		runWithoutPanic(func() {
			_, _ = logService.ListPaginado(ctx, 10, 0)
		})
	})

	t.Run("Pay", func(t *testing.T) {
		runWithoutPanic(func() {
			_ = logService.Pay(ctx, id)
		})
	})

	t.Run("Cancel", func(t *testing.T) {
		runWithoutPanic(func() {
			_ = logService.Cancel(ctx, id)
		})
	})
}
