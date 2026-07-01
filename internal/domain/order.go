package domain

// 1. Criando o "Enum" de Status
// Definimos que "StatusPedido" é, na base, uma string
type StatusPedido string

// constantes que usam esse novo tipo
const (
	StatusPendente  StatusPedido = "PENDENTE"
	StatusPago      StatusPedido = "PAGO"
	StatusCancelado StatusPedido = "CANCELADO"
)

// 2. O Item do Pedido (A ponte entre o Pedido e o Produto)
type ItemPedido struct {
	Produto    *Produto // O ponteiro liga este item diretamente à "struct" de "Produto" que esta no "package domain"
	Preco      float64  // Guardamos o preço aqui para "congelar" o "valor" no momento da compra
	Quantidade int
}

// 3. O Pedido Principal
type Pedido struct {
	ID      string
	Cliente string
	Itens   []ItemPedido // Um slice (lista) de itens
	Status  StatusPedido
}

// "CalcularTotal" deve percorrer os "Itens do pedido", "multiplicar" o "Preco" pela "Quantidade"
// de cada item e "retornar" a "soma total".
func (p *Pedido) CalcularTotal() float64 {
	var somaTotal float64

	// Usamos "p.Itens" porque o receiver (p) já possui a lista de itens do pedido
	for _, item := range p.Itens {
		// Usamos += para ir somando o valor de cada item no acumulador
		somaTotal += item.Preco * float64(item.Quantidade)
	}

	// O return fica fora do 'for', devolvendo o valor "só depois" de somar tudo
	return somaTotal
}

// "Pagar" deve mudar o status do pedido para "StatusPago".
// Regra do desafio: Só pode pagar se o status atual for "StatusPendente".
// Se já estiver "pago" ou "cancelado", deve retornar o erro "ErrMudancaDeStatusInvalida".
func (p *Pedido) Pagar() error {
	// 'if' para checar p.Status
	if p.Status != StatusPendente {
		return ErrMudancaDeStatusInvalida
	}

	p.Status = StatusPago
	return nil
}

// "Cancelar" deve mudar o status do pedido para "StatusCancelado".
// Regra do desafio: Só pode cancelar se o status atual for "StatusPendente".
// Se já estiver "pago" ou "cancelado", deve retornar o mesmo erro "ErrMudancaDeStatusInvalida".
func (p *Pedido) Cancelar() error {
	// 'if' para checar p.Status
	if p.Status != StatusPendente {
		return ErrMudancaDeStatusInvalida
	}

	p.Status = StatusCancelado
	return nil
}
