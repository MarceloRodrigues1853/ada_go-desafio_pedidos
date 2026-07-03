package service

import (
	"pedidos/internal/domain"
	"pedidos/internal/repository"
)

// OrderService é o serviço de pedidos
type OrderService struct {
	produtoRepo repository.ProductRepository
	pedidoRepo  repository.OrderRepository
}

// NewOrderService injeta as dependências (os bancos de dados) no serviço
func NewOrderService(pr repository.ProductRepository, or repository.OrderRepository) *OrderService {
	return &OrderService{
		produtoRepo: pr,
		pedidoRepo:  or,
	}
}

// CriarPedido recebe o ID do cliente, o ID do produto e a quantidade.
// (Para simplificar o desafio, estamos criando pedido de 1 produto só por vez)
func (s *OrderService) CriarPedido(idPedido, cliente, idProduto string, quantidade int) (*domain.Pedido, error) {

	// 1. Cria a base do pedido usando a Factory Function.
	// Todo pedido nasce com o status "PENDENTE".
	pedido, err := domain.NovoPedido(idPedido, cliente, domain.StatusPendente)
	if err != nil {
		return nil, err // Se o ID for vazio, a missão é abortada aqui mesmo
	}

	// 2. Vamos ao banco de dados (repositório) ver se o produto solicitado realmente existe.
	produto, err := s.produtoRepo.BuscarPorID(idProduto)
	if err != nil {
		// Se der erro (ex: ProdutoNaoEncontrado), aborta a missão e devolve o erro.
		return nil, err
	}

	// 3. O coração do negócio: pedimos para o produto reduzir o próprio estoque.
	// retorna um erro se faltar saldo
	if err := produto.ReduzirEstoque(quantidade); err != nil {
		return nil, err // Aborta se o estoque for insuficiente
	}

	// 4. Se chegou até aqui, temos estoque! Vamos montar o item do carrinho.
	// Guardamos o preço atual para garantir que o cliente pague o valor congelado de hoje.
	item := domain.ItemPedido{
		Produto:    produto,
		Preco:      produto.Preco,
		Quantidade: quantidade,
	}

	// Usamos o append() nativo do Go para adicionar esse item à lista vazia do pedido.
	pedido.Itens = append(pedido.Itens, item)

	// 5. Persistência de Dados: O pedido foi montado e o estoque do produto mudou.
	// Precisamos salvar os dois estados novos nos seus respectivos bancos de dados.
	if err := s.produtoRepo.Salvar(produto); err != nil {
		return nil, err
	}

	if err := s.pedidoRepo.Salvar(pedido); err != nil {
		return nil, err
	}

	// Sucesso absoluto! Retornamos o ponteiro do pedido montado e "nil" para o erro.
	return pedido, nil
}

// "PagarPedido" busca o pedido, muda o status e salva novamente
func (s *OrderService) PagarPedido(idPedido string) error {

	// PASSO 1: Buscar o pedido
	pedido, err := s.pedidoRepo.BuscarPorID(idPedido)
	if err != nil {
		return err
	}

	// PASSO 2: Alterar a regra de negócio
	// Apenas mandamos o pedido se pagar. Se a regra lá no domínio falhar,
	// ele nos devolve o erro e nós abortamos.
	if err := pedido.Pagar(); err != nil {
		return err
	}

	// PASSO 3: Salvar a alteração no banco de dados
	if err := s.pedidoRepo.Salvar(pedido); err != nil {
		return err
	}

	return nil
}

// CancelarPedido busca o pedido, muda o status, devolve o estoque do produto e salva
func (s *OrderService) CancelarPedido(idPedido string) error {

	// 1. Busca o pedido
	pedido, err := s.pedidoRepo.BuscarPorID(idPedido)
	if err != nil {
		return err
	}

	// 2. tenta cancelar (o dominio valida se pode)
	if err := pedido.Cancelar(); err != nil {
		return err
	}

	// 3. Devolve os itens para o estoque do produto
	// Como o pedido tem uma lista de itens, fazemos um laço 'for'
	for _, item := range pedido.Itens {

		item.Produto.DevolverEstoque(item.Quantidade)

		// Salvamos o produto atualizado no banco
		if err := s.produtoRepo.Salvar(item.Produto); err != nil {
			return err
		}
	}

	// 4. Salva o pedido cancelado no banco
	if err := s.pedidoRepo.Salvar(pedido); err != nil {
		return err
	}

	return nil
}
