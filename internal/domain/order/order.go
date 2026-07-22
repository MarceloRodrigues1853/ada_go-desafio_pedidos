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

// ErrCannotCancelPaidOrder é retornado ao tentar cancelar um pedido já pago.
var ErrCannotCancelPaidOrder = errors.New("cannot cancel a paid order")

// Order é a entidade de domínio do pedido.
// O status só muda via métodos de intenção de negócio (Pay, Cancel).
type Order struct {
	id       uuid.UUID
	clientID uuid.UUID
	status   Status
}

// NewOrder cria um pedido com status PENDING.
func NewOrder(clientID uuid.UUID) (*Order, error) {
	return &Order{
		id:       uuid.New(),
		clientID: clientID,
		status:   StatusPending,
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
