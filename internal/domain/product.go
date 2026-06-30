package domain

// Produto representa a entidade de domínio principal
type Produto struct {
	ID      string
	Nome    string
	Preco   float64
	Estoque int
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
