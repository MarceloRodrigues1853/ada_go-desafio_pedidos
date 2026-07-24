package order

import (
	"errors"
	"testing"

	"github.com/google/uuid"
)

// mustItem cria um OrderItem válido para uso nos testes.
func mustItem(t *testing.T, productID uuid.UUID, quantity int, price float64) OrderItem {
	t.Helper()
	item, err := NewOrderItem(productID, quantity, price)
	if err != nil {
		t.Fatalf("NewOrderItem() retornou erro inesperado: %v", err)
	}
	return item
}

// TestNewOrder_StatusPending garante que um pedido recém-criado
// inicia com o status PENDING.
func TestNewOrder_StatusPending(t *testing.T) {
	clientID := uuid.New()
	items := []OrderItem{mustItem(t, uuid.New(), 1, 10.0)}

	pedido, err := NewOrder(clientID, items)
	if err != nil {
		t.Fatalf("NewOrder() retornou erro inesperado: %v", err)
	}

	if pedido.Status() != StatusPending {
		t.Errorf("status = %q, quero %q", pedido.Status(), StatusPending)
	}
}

// TestNewOrder_EmptyItems_ReturnsErrEmptyOrder garante que um pedido
// sem itens não pode ser criado e retorna ErrEmptyOrder.
func TestNewOrder_EmptyItems_ReturnsErrEmptyOrder(t *testing.T) {
	_, err := NewOrder(uuid.New(), nil)
	if !errors.Is(err, ErrEmptyOrder) {
		t.Errorf("NewOrder() erro = %v, quero %v", err, ErrEmptyOrder)
	}

	_, err = NewOrder(uuid.New(), []OrderItem{})
	if !errors.Is(err, ErrEmptyOrder) {
		t.Errorf("NewOrder() com slice vazio erro = %v, quero %v", err, ErrEmptyOrder)
	}
}

// TestOrder_Pay_ChangesStatusToPaid verifica que Pay() aprova o pagamento
// e altera o status do pedido de PENDING para PAID.
func TestOrder_Pay_ChangesStatusToPaid(t *testing.T) {
	items := []OrderItem{mustItem(t, uuid.New(), 1, 10.0)}

	pedido, err := NewOrder(uuid.New(), items)
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
	items := []OrderItem{mustItem(t, uuid.New(), 1, 10.0)}

	pedido, err := NewOrder(uuid.New(), items)
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

// TestOrder_CalculateTotal verifica que CalculateTotal() soma corretamente
// os subtotais (preço × quantidade) de todos os itens do pedido.
func TestOrder_CalculateTotal(t *testing.T) {
	items := []OrderItem{
		mustItem(t, uuid.New(), 2, 10.0), // subtotal: 20.0
		mustItem(t, uuid.New(), 3, 5.50), // subtotal: 16.5
	}

	pedido, err := NewOrder(uuid.New(), items)
	if err != nil {
		t.Fatalf("NewOrder() retornou erro inesperado: %v", err)
	}

	const want = 36.5
	got := pedido.CalculateTotal()
	if got != want {
		t.Errorf("CalculateTotal() = %v, quero %v", got, want)
	}
}
