package order

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// TestNewOrder_StatusPending garante que um pedido recém-criado
// inicia com o status PENDING.
func TestNewOrder_StatusPending(t *testing.T) {
	clientID := uuid.New()

	pedido, err := NewOrder(clientID)
	if err != nil {
		t.Fatalf("NewOrder() retornou erro inesperado: %v", err)
	}

	if pedido.Status() != StatusPending {
		t.Errorf("status = %q, quero %q", pedido.Status(), StatusPending)
	}
}

// TestOrder_Pay_ChangesStatusToPaid verifica que Pay() aprova o pagamento
// e altera o status do pedido de PENDING para PAID.
func TestOrder_Pay_ChangesStatusToPaid(t *testing.T) {
	pedido, err := NewOrder(uuid.New())
	if err != nil {
		t.Fatalf("NewOrder() retornou erro inesperado: %v", err)
	}

	if err := pedido.Pay(); err != nil {
		t.Fatalf("Pay() retornou erro inesperado: %v", err)
	}

	if pedido.Status() != StatusPaid {
		t.Errorf("status = %q, quero %q", pedido.Status(), StatusPaid)
	}
}

// TestOrder_Cancel_PaidOrder_ReturnsErrCannotCancelPaidOrder garante a regra
// de negócio: um pedido já pago não pode ser cancelado e deve retornar
// ErrCannotCancelPaidOrder.
func TestOrder_Cancel_PaidOrder_ReturnsErrCannotCancelPaidOrder(t *testing.T) {
	pedido, err := NewOrder(uuid.New())
	if err != nil {
		t.Fatalf("NewOrder() retornou erro inesperado: %v", err)
	}

	// Coloca o pedido em PAID antes de tentar cancelar.
	if err := pedido.Pay(); err != nil {
		t.Fatalf("Pay() retornou erro inesperado: %v", err)
	}

	err = pedido.Cancel()
	if !errors.Is(err, ErrCannotCancelPaidOrder) {
		t.Errorf("Cancel() erro = %v, quero %v", err, ErrCannotCancelPaidOrder)
	}
}
