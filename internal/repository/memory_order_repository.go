package repository

import "pedidos/internal/domain"

// "memoryOrderRepository" é uma implementação de "OrderRepository" que armazena os pedidos em memória
// A struct começa com letra minúscula pois não queremos que outras pastas
// instanciem ela diretamente, forçando o uso do "construtor" abaixo.
type memoryOrderRepository struct {
	pedidos map[string]*domain.Pedido
}

// "NewMemoryOrderRepository" atua como a função construtora.
// É aqui que nós inicializamos o map na memória para ele não dar erro de "nil map".
func NewMemoryOrderRepository() OrderRepository {
	return &memoryOrderRepository{
		pedidos: make(map[string]*domain.Pedido),
	}
}

// "Salvar" é o método que salva um pedido no map.
func (r *memoryOrderRepository) Salvar(pedido *domain.Pedido) error {
	// A chave do map é o ID (string) e o valor é a struct inteira (o ponteiro)
	r.pedidos[pedido.ID] = pedido

	// Retorna nil pois não há erro ao salvar o pedido
	return nil
}

// "BuscarPorID" é o método que busca um pedido no map pelo ID.
func (r *memoryOrderRepository) BuscarPorID(id string) (*domain.Pedido, error) {
	// Busca o pedido no map pelo ID
	pedidoEncontrado, existe := r.pedidos[id]

	// Se o pedido não for encontrado, retorna o erro "ErrPedidoNaoEncontrado"
	if !existe {
		return nil, domain.ErrPedidoNaoEncontrado
	}

	// Retorna o pedido encontrado e nil pois não há erro ao buscar o pedido
	return pedidoEncontrado, nil
}

// "Listar" é o método que lista todos os pedidos no map.
func (r *memoryOrderRepository) Listar() ([]*domain.Pedido, error) {
	// Cria um slice vazio com capacidade para o número de pedidos no map
	pedidos := make([]*domain.Pedido, 0, len(r.pedidos))

	// Percorre o map e adiciona cada pedido ao slice
	for _, pedido := range r.pedidos {
		// Adiciona o pedido ao slice
		pedidos = append(pedidos, pedido)
	}

	// Retorna o slice de pedidos e nil pois não há erro ao listar os pedidos
	return pedidos, nil
}
