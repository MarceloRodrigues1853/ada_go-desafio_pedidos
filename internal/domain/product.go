package domain

// Produto representa a entidade de domínio principal
type Produto struct {
	ID      string
	Nome    string
	Preco   float64
	Estoque int
}

// novoProduto cria um novo produto
func NovoProduto(id, nome string, preco float64, estoque int) (*Produto, error) {
	produto := &Produto{
		ID:      id,
		Nome:    nome,
		Preco:   preco,
		Estoque: estoque,
	}

	// Validamos o produto
	if err := produto.validarProduto(); err != nil {
		return nil, err
	}

	// Retornamos o produto validado
	return produto, nil
}

// DevolverEstoque repõe itens no estoque (ex: quando um pedido é cancelado)
// usamos `p.Estoque += quantidade` para SOMAR, e não substituir (=)
func (p *Produto) DevolverEstoque(quantidade int) {
	p.Estoque += quantidade
}

// ReduzirEstoque tira itens do estoque, mas barra a operação se não houver saldo.
// devolvendo um `error` como resposta.
func (p *Produto) ReduzirEstoque(quantidade int) error {
	// 1. A regra de negócio: Tentar tirar mais do que tem
	if p.Estoque < quantidade {
		// Retornamos errors caso não tenha saldo
		return ErrEstoqueInsuficiente
	}

	// 2. Se passou pelo if (caminho feliz), subtraímos o estoque
	p.Estoque -= quantidade

	// 3. Devolvemos nil (nulo) para avisar que nenhum erro aconteceu
	return nil
}

func (p *Produto) validarProduto() error {
	if p.ID == "" || p.Nome == "" || p.Preco <= 0 || p.Estoque < 0 {
		return ErrProdutoInvalido
	}
	return nil

}
