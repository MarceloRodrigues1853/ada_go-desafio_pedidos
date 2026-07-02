package domain

import "errors"

// Usamos um bloco 'var' com parênteses para declarar várias variáveis de uma vez
var (
	// tipos de erro exigidos:
	ErrProdutoNaoEncontrado    = errors.New("produto não encontrado")
	ErrPedidoNaoEncontrado     = errors.New("pedido não encontrado")
	ErrQuantidadeInvalida      = errors.New("quantidade inválida")
	ErrEstoqueInsuficiente     = errors.New("estoque insuficiente")
	ErrClienteInvalido         = errors.New("cliente inválido")
	ErrPedidoVazio             = errors.New("pedido vazio")
	ErrMudancaDeStatusInvalida = errors.New("mudança de status inválida")
	ErrProdutoInvalido         = errors.New("produto inválido")
)
