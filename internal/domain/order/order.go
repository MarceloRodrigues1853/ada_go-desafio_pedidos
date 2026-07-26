package order

import (
	"errors"

	"github.com/google/uuid"
)

// Status representa o ciclo de vida do pedido.
type Status string

const (
	StatusPending  Status = "PENDING"
	StatusPaid     Status = "PAID"
	StatusCanceled Status = "CANCELED"
)

var (
	// ErrCannotCancelPaidOrder é retornado ao tentar cancelar um pedido já pago.
	ErrCannotCancelPaidOrder = errors.New("cannot cancel a paid order")
	// ErrEmptyOrder é retornado ao tentar criar um pedido sem itens.
	ErrEmptyOrder = errors.New("order must contain at least one item")
)

// Order é o agregado raiz do pedido.
// O status só muda via métodos de intenção de negócio (Pay, Cancel).
type Order struct {
	id       uuid.UUID
	clientID uuid.UUID
	status   Status
	items    []OrderItem
}

// NewOrder cria um pedido com status PENDING e exige pelo menos um item.
func NewOrder(clientID uuid.UUID, items []OrderItem) (*Order, error) {
	if len(items) == 0 {
		return nil, ErrEmptyOrder
	}

	// Cópia defensiva: o agregado não compartilha o slice externo.
	copied := make([]OrderItem, len(items))
	copy(copied, items)

	return &Order{
		id:       uuid.New(),
		clientID: clientID,
		status:   StatusPending,
		items:    copied,
	}, nil
}

// ID retorna o identificador do pedido.
func (o *Order) ID() uuid.UUID {
	return o.id
}

// ClientID retorna o identificador do cliente.
func (o *Order) ClientID() uuid.UUID {
	return o.clientID
}

// Status retorna o status atual do pedido.
func (o *Order) Status() Status {
	return o.status
}

// Items retorna uma cópia dos itens do pedido.
func (o *Order) Items() []OrderItem {
	copied := make([]OrderItem, len(o.items))
	copy(copied, o.items)
	return copied
}

// CalculateTotal soma os subtotais de todos os itens do pedido.
func (o *Order) CalculateTotal() float64 {
	var total float64
	for _, item := range o.items {
		total += item.Subtotal()
	}
	return total
}

// Pay aprova o pagamento e altera o status para PAID.
func (o *Order) Pay() error {
	o.status = StatusPaid
	return nil
}

// Cancel cancela o pedido. Pedidos pagos não podem ser cancelados.
func (o *Order) Cancel() error {
	if o.status == StatusPaid {
		return ErrCannotCancelPaidOrder
	}

	o.status = StatusCanceled
	return nil
}

// Restore reconstitui um pedido existente vindo do banco de dados
func Restore(id uuid.UUID, clientID uuid.UUID, status Status, items []OrderItem) *Order {
	return &Order{
		id:       id,
		clientID: clientID,
		status:   status,
		items:    items,
	}
}
