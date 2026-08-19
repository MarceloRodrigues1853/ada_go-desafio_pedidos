package service_test

import (
	"context"
	"testing"
	"time"

	"pedidos/internal/domain/order"
	"pedidos/internal/events"
	"pedidos/internal/repository"
	"pedidos/internal/service"

	"github.com/google/uuid"
)

type mockOrderRepo struct {
	repository.OrderRepository
	orderRecord *repository.OrderRecord
	getErr      error
	updateErr   error
	txErr       error
	returnedQty int
	returnedPID string
}

func (m *mockOrderRepo) GetByID(ctx context.Context, id uuid.UUID) (*repository.OrderRecord, error) {
	return m.orderRecord, m.getErr
}

func (m *mockOrderRepo) UpdateStatus(ctx context.Context, id uuid.UUID, status order.Status) error {
	return m.updateErr
}

func (m *mockOrderRepo) WithinTransaction(ctx context.Context, fn func(repository.OrderTransaction) error) error {
	if m.txErr != nil {
		return m.txErr
	}
	tx := &mockOrderTx{
		repo: m,
	}
	return fn(tx)
}

type mockOrderTx struct {
	repository.OrderTransaction
	repo *mockOrderRepo
}

func (tx *mockOrderTx) ReturnStock(ctx context.Context, productID string, quantity int) error {
	tx.repo.returnedPID = productID
	tx.repo.returnedQty = quantity
	return nil
}

func (tx *mockOrderTx) UpdateStatus(ctx context.Context, id uuid.UUID, status order.Status) error {
	return tx.repo.updateErr
}

func TestOrderService_ProcessPaymentResult(t *testing.T) {
	clientID := uuid.New()
	item, _ := order.NewOrderItem("SKU-001", 2, 50.0)
	ord, _ := order.NewOrder(clientID, []order.OrderItem{item})

	t.Run("Fluxo Feliz - Pagamento Aprovado (PAID)", func(t *testing.T) {
		repo := &mockOrderRepo{
			orderRecord: &repository.OrderRecord{
				Order:     ord,
				CreatedAt: time.Now(),
			},
		}

		svc := service.NewOrderService(nil, nil, repo, nil)

		event := events.PaymentProcessedEvent{
			SagaID:      uuid.New(),
			OrderID:     ord.ID(),
			ClientID:    clientID,
			TotalAmount: 100.0,
			Status:      "PAID",
			ProcessedAt: time.Now(),
		}

		err := svc.ProcessPaymentResult(context.Background(), event)
		if err != nil {
			t.Fatalf("esperava sucesso ao processar pagamento PAID, obteve erro: %v", err)
		}

		if ord.Status() != order.StatusPaid {
			t.Errorf("esperava status do pedido PAID, obteve %s", ord.Status())
		}
	})

	t.Run("Fluxo de Compensação - Pagamento Falhou (FAILED)", func(t *testing.T) {
		item2, _ := order.NewOrderItem("SKU-002", 3, 30.0)
		ord2, _ := order.NewOrder(clientID, []order.OrderItem{item2})

		repo := &mockOrderRepo{
			orderRecord: &repository.OrderRecord{
				Order:     ord2,
				CreatedAt: time.Now(),
			},
		}

		svc := service.NewOrderService(nil, nil, repo, nil)

		event := events.PaymentProcessedEvent{
			SagaID:      uuid.New(),
			OrderID:     ord2.ID(),
			ClientID:    clientID,
			TotalAmount: 90.0,
			Status:      "FAILED",
			ProcessedAt: time.Now(),
		}

		err := svc.ProcessPaymentResult(context.Background(), event)
		if err != nil {
			t.Fatalf("esperava sucesso no fluxo de compensação, obteve erro: %v", err)
		}

		if ord2.Status() != order.StatusCanceled {
			t.Errorf("esperava status do pedido CANCELED, obteve %s", ord2.Status())
		}

		if repo.returnedPID != "SKU-002" || repo.returnedQty != 3 {
			t.Errorf("esperava estorno de 3 unidades do SKU-002, obteve %d de %s", repo.returnedQty, repo.returnedPID)
		}
	})
}
