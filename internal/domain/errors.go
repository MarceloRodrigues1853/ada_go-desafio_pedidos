package domain

import "errors"

// Usamos um bloco 'var' com parênteses para declarar várias variáveis de uma vez
var (
	// tipos de erro exigidos:
	ErrProdutoNaoEncontrado    = errors.New("produto não encontrado")            // Produto inexistente no catálogo
	ErrPedidoNaoEncontrado     = errors.New("pedido não encontrado")             // Pedido inexistente no banco
	ErrQuantidadeInvalida      = errors.New("quantidade inválida")               // Quantidade menor ou igual a zero
	ErrEstoqueInsuficiente     = errors.New("estoque insuficiente")              // Tentativa de reservar mais do que há em estoque
	ErrClienteInvalido         = errors.New("cliente inválido")                  // Cliente com dados inválidos
	ErrPedidoVazio             = errors.New("pedido vazio")                      // Pedido sem itens
	ErrMudancaDeStatusInvalida = errors.New("mudança de status inválida")        // Transição de status não permitida
	ErrProdutoInvalido         = errors.New("produto inválido")                  // Produto com dados inválidos
	ErrPedidoInvalido          = errors.New("id ou cliente do pedido inválidos") // Pedido com id ou cliente inválidos
)
