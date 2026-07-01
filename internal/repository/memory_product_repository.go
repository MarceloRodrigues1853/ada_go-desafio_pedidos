package repository

import "pedidos/internal/domain"

// "MemoryProductRepository" é uma implementação de "ProductRepository" que armazena os produtos em memória
// A struct começa com letra minúscula pois não queremos que outras pastas
// instanciem ela diretamente, forçando o uso do "construtor" abaixo.
type memoryProductRepository struct {
	produtos map[string]*domain.Produto
}

// "NewMemoryProductRepository" atua como a função construtora.
// É aqui que nós inicializamos o map na memória para ele não dar erro de "nil map".
func NewMemoryProductRepository() ProductRepository {
	return &memoryProductRepository{
		produtos: make(map[string]*domain.Produto),
	}
}

// "Salvar" é o método que salva um produto no map.
func (r *memoryProductRepository) Salvar(produto *domain.Produto) error {
	// A chave do map é o ID (string) e o valor é a struct inteira (o ponteiro)
	r.produtos[produto.ID] = produto

	// Retorna nil pois não há erro ao salvar o produto
	return nil
}

// "BuscarPorID" é o método que busca um produto no map pelo ID.
func (r *memoryProductRepository) BuscarPorID(id string) (*domain.Produto, error) {
	// Busca o produto no map pelo ID
	produtoEncontrado, existe := r.produtos[id]

	// Se o produto não for encontrado, retorna o erro "ErrProdutoNaoEncontrado"
	if !existe {
		return nil, domain.ErrProdutoNaoEncontrado
	}

	// Retorna o produto encontrado e nil pois não há erro ao buscar o produto
	return produtoEncontrado, nil
}

// "Listar" é o método que lista todos os produtos no map.
func (r *memoryProductRepository) Listar() ([]*domain.Produto, error) {
	// Cria um slice vazio com capacidade para o número de produtos no map
	produtos := make([]*domain.Produto, 0, len(r.produtos))

	// Percorre o map e adiciona cada produto ao slice
	for _, produto := range r.produtos {
		// Adiciona o produto ao slice
		produtos = append(produtos, produto)
	}

	// Retorna o slice de produtos e nil pois não há erro ao listar os produtos
	return produtos, nil
}
