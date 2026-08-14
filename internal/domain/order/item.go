package order

import "errors"

// Erros de validação do item do pedido.
var (
	ErrInvalidQuantity = errors.New("quantity must be greater than zero")
	ErrInvalidPrice    = errors.New("price must be greater than zero")
)

// OrderItem é a entidade de linha do agregado Order.
// Quantidade e preço ficam encapsulados e só nascem válidos via NewOrderItem.
type OrderItem struct {
	productID string
	quantity  int
	price     float64
}

// NewOrderItem cria um item validando quantidade > 0 e preço > 0.

func NewOrderItem(productID string, quantity int, price float64) (OrderItem, error) {
	if productID == "" {
		return OrderItem{}, errors.New("product id is required")
	}
	if quantity <= 0 {
		return OrderItem{}, ErrInvalidQuantity
	}
	if price <= 0 {
		return OrderItem{}, ErrInvalidPrice
	}

	return OrderItem{
		productID: productID,
		quantity:  quantity,
		price:     price,
	}, nil
}

// ProductID retorna o identificador do produto.
func (i OrderItem) ProductID() string {
	return i.productID
}

// Quantity retorna a quantidade do item.
func (i OrderItem) Quantity() int {
	return i.quantity
}

// Price retorna o preço unitário congelado no momento da compra.
func (i OrderItem) Price() float64 {
	return i.price
}

// Subtotal retorna preço × quantidade.
func (i OrderItem) Subtotal() float64 {
	return i.price * float64(i.quantity)
}
