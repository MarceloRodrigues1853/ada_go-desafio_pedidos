package service

import (
	"context"
	"log/slog"
	"time"

	"pedidos/internal/events"

	"github.com/google/uuid"
)

// LoggingOrderService é um wrapper que adiciona logging ao OrderService
type LoggingOrderService struct {
	inner  OrderServiceInterface // Service interno que será encapsulado
	logger *slog.Logger          // logger para não resgistrar as operações
}

// NewloggingOrderService cria um novo LoggingOrderService
func NewLoggingOrderService(inner OrderServiceInterface, logger *slog.Logger) *LoggingOrderService {
	return &LoggingOrderService{
		inner:  inner,
		logger: logger.With("service", "orders", "component", "order_service"),
	}
}

// Garante que LoggingOrderService implementa OrderServiceInterface
var _ OrderServiceInterface = (*LoggingOrderService)(nil)

// Pay implementa o método Pay do OrderServiceInterface
func (s *LoggingOrderService) Pay(ctx context.Context, pedidoID uuid.UUID) error {
	start := time.Now()
	err := s.inner.Pay(ctx, pedidoID)
	if err != nil {
		s.logger.Warn("pay failed",
			"pedido_id", pedidoID,
			"duration_ms", time.Since(start).Milliseconds(),
			"erro", err.Error())
		return err
	}
	s.logger.Info("pay ok",
		"pedido_id", pedidoID,
		"duration_ms",
		time.Since(start).Milliseconds())
	return nil
}

// Create implementa o método Create do OrderServiceInterface
func (s *LoggingOrderService) Create(ctx context.Context, clienteID uuid.UUID, itens []OrderItemInput) (*OrderOutput, error) {
	start := time.Now()
	pedido, err := s.inner.Create(ctx, clienteID, itens)
	if err != nil {
		s.logger.Warn("create failed",
			"cliente_id", clienteID,
			"itens", len(itens),
			"duration_ms", time.Since(start).Milliseconds(),
			"erro", err.Error())
		return nil, err
	}
	s.logger.Info("create ok",
		"pedido_id", pedido.ID,
		"cliente_id", clienteID,
		"duration_ms", time.Since(start).Milliseconds())
	return pedido, nil
}

// GetByID implementa o método GetByID do OrderServiceInterface
func (s *LoggingOrderService) GetByID(ctx context.Context, pedidoID uuid.UUID) (*OrderOutput, error) {
	start := time.Now()
	pedido, err := s.inner.GetByID(ctx, pedidoID)
	if err != nil {
		s.logger.Warn("get_by_id failed",
			"pedido_id", pedidoID,
			"duration_ms", time.Since(start).Milliseconds(),
			"erro", err.Error())
		return nil, err
	}
	s.logger.Info("get_by_id ok",
		"pedido_id", pedidoID,
		"duration_ms", time.Since(start).Milliseconds())
	return pedido, nil
}

// ListPaginado implementa o método ListPaginado do OrderServiceInterface
func (s *LoggingOrderService) ListPaginado(ctx context.Context, limit, offset int32) ([]OrderOutput, error) {
	start := time.Now()
	pedido, err := s.inner.ListPaginado(ctx, limit, offset)
	if err != nil {
		s.logger.Warn("list_painado failed",
			"limit", limit,
			"offset", offset,
			"duration_ms", time.Since(start).Milliseconds(),
			"erro", err.Error())
		return nil, err
	}
	s.logger.Info("list_pagindando ok",
		"limit", limit,
		"offset", offset,
		"count", len(pedido),
		"duration_ms", time.Since(start).Milliseconds())
	return pedido, nil
}

// Cancel implementa o método Cancel do OrderServiceInterface
func (s *LoggingOrderService) Cancel(ctx context.Context, pedidoID uuid.UUID) error {
	start := time.Now()
	err := s.inner.Cancel(ctx, pedidoID)
	if err != nil {
		s.logger.Warn("cancel failed",
			"pedido_id", pedidoID,
			"duration_ms", time.Since(start).Milliseconds(),
			"erro", err.Error())
		return err
	}
	s.logger.Info("cancel ok",
		"pedido_id", pedidoID,
		"duration_ms",
		time.Since(start).Milliseconds())
	return nil
}

// ProcessPaymentResult implementa o método ProcessPaymentResult do OrderServiceInterface
func (s *LoggingOrderService) ProcessPaymentResult(ctx context.Context, event events.PaymentProcessedEvent) error {
	start := time.Now()
	err := s.inner.ProcessPaymentResult(ctx, event)
	if err != nil {
		s.logger.Warn("process_payment_result failed",
			"saga_id", event.SagaID,
			"order_id", event.OrderID,
			"status", event.Status,
			"duration_ms", time.Since(start).Milliseconds(),
			"erro", err.Error())
		return err
	}
	s.logger.Info("process_payment_result ok",
		"saga_id", event.SagaID,
		"order_id", event.OrderID,
		"status", event.Status,
		"duration_ms", time.Since(start).Milliseconds())
	return nil
}
