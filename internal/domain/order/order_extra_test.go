package order_test

import (
	"testing"

	"pedidos/internal/domain/order"

	"github.com/google/uuid"
)

func TestOrderGettersAndRestore(t *testing.T) {
	productID := uuid.New()
	item, err := order.NewOrderItem(productID.String(), 2, 50.0)
	if err != nil {
		t.Fatalf("erro inesperado: %v", err)
	}

	// Testa getters do Item
	if item.ProductID() != productID.String() {
		t.Errorf("esperava %v, veio %v", productID, item.ProductID())
	}
	if item.Quantity() != 2 {
		t.Errorf("esperava 2, veio %d", item.Quantity())
	}
	if item.Price() != 50.0 {
		t.Errorf("esperava 50.0, veio %f", item.Price())
	}

	// Testa Restore e getters do Order
	orderID := uuid.New()
	clientID := uuid.New()
	ord := order.Restore(orderID, clientID, order.StatusPending, []order.OrderItem{item})

	if ord.ID() != orderID {
		t.Errorf("esperava %v, veio %v", orderID, ord.ID())
	}
	if ord.ClientID() != clientID {
		t.Errorf("esperava %v, veio %v", clientID, ord.ClientID())
	}
	if len(ord.Items()) != 1 {
		t.Errorf("esperava 1 item, veio %d", len(ord.Items()))
	}
}
