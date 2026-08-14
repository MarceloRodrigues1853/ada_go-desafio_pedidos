package domain_test

import (
	"testing"

	// Importa o pacote de domínio local
	"pedidos/internal/domain"
)

// TestNovoProduto_Sucesso valida a criação de um produto com dados corretos
func TestNovoProduto_Sucesso(t *testing.T) {
	// Dados de entrada válidos
	id := "P001"
	nome := "Teclado Mecânico"
	preco := 250.00
	estoque := 10

	// Executa o construtor do domínio
	p, err := domain.NovoProduto(id, nome, preco, estoque)

	// Valida se a criação ocorreu sem erros
	if err != nil {
		t.Fatalf("esperava sucesso ao criar produto, mas recebeu erro: %v", err)
	}

	// Valida se os atributos foram atribuídos corretamente
	if p.ID != id || p.Nome != nome || p.Preco != preco || p.Estoque != estoque {
		t.Errorf("dados do produto criados incorretamente. Recebeu: %+v", p)
	}
}

// TestNovoProduto_Validacao testa todas as invariantes da função validarProduto()
func TestNovoProduto_Validacao(t *testing.T) {
	// Subteste 1: ID vazio
	t.Run("ID vazio deve retornar ErrProdutoInvalido", func(t *testing.T) {
		_, err := domain.NovoProduto("", "Teclado", 100.0, 5)
		if err != domain.ErrProdutoInvalido {
			t.Errorf("esperava ErrProdutoInvalido, mas recebeu: %v", err)
		}
	})

	// Subteste 2: Nome vazio
	t.Run("Nome vazio deve retornar ErrProdutoInvalido", func(t *testing.T) {
		_, err := domain.NovoProduto("P001", "", 100.0, 5)
		if err != domain.ErrProdutoInvalido {
			t.Errorf("esperava ErrProdutoInvalido, mas recebeu: %v", err)
		}
	})

	// Subteste 3: Preço menor ou igual a zero
	t.Run("Preco menor ou igual a zero deve retornar ErrProdutoInvalido", func(t *testing.T) {
		_, err := domain.NovoProduto("P001", "Teclado", 0.0, 5)
		if err != domain.ErrProdutoInvalido {
			t.Errorf("esperava ErrProdutoInvalido para preco 0, mas recebeu: %v", err)
		}
	})

	// Subteste 4: Estoque negativo na criação
	t.Run("Estoque negativo deve retornar ErrProdutoInvalido", func(t *testing.T) {
		_, err := domain.NovoProduto("P001", "Teclado", 100.0, -1)
		if err != domain.ErrProdutoInvalido {
			t.Errorf("esperava ErrProdutoInvalido para estoque negativo, mas recebeu: %v", err)
		}
	})
}

// TestProduto_GerenciamentoEstoque testa os métodos de movimentação de estoque
func TestProduto_GerenciamentoEstoque(t *testing.T) {
	// Cria um produto base com 10 unidades para os testes
	p, err := domain.NovoProduto("P001", "Teclado", 100.0, 10)
	if err != nil {
		t.Fatalf("erro ao criar produto para teste de estoque: %v", err)
	}

	// Teste de Redução de Estoque com sucesso
	t.Run("ReduzirEstoque com saldo suficiente deve subtrair corretamente", func(t *testing.T) {
		err := p.ReduzirEstoque(3) // Reduz 3 de 10
		if err != nil {
			t.Fatalf("nao esperava erro ao reduzir estoque, recebeu: %v", err)
		}

		if p.Estoque != 7 {
			t.Errorf("esperava estoque igual a 7, mas obteve %d", p.Estoque)
		}
	})

	// Teste de Redução de Estoque maior que o saldo (Invariante)
	t.Run("ReduzirEstoque sem saldo deve retornar ErrEstoqueInsuficiente", func(t *testing.T) {
		// O estoque atual é 7. Tentamos tirar 10.
		err := p.ReduzirEstoque(10)
		if err != domain.ErrEstoqueInsuficiente {
			t.Errorf("esperava ErrEstoqueInsuficiente, mas recebeu: %v", err)
		}

		// Garante que o estoque não ficou negativo após a tentativa
		if p.Estoque != 7 {
			t.Errorf("estoque nao deveria ter alterado, esperava 7, obteve %d", p.Estoque)
		}
	})

	t.Run("ReduzirEstoque com quantidade não positiva deve retornar ErrQuantidadeInvalida", func(t *testing.T) {
		err := p.ReduzirEstoque(0)
		if err != domain.ErrQuantidadeInvalida {
			t.Errorf("esperava ErrQuantidadeInvalida, mas recebeu: %v", err)
		}
	})

	// Teste de Devolução de Estoque
	t.Run("DevolverEstoque deve somar ao saldo atual", func(t *testing.T) {
		p.DevolverEstoque(5) // Soma 5 ao saldo atual de 7
		if p.Estoque != 12 {
			t.Errorf("esperava estoque igual a 12, mas obteve %d", p.Estoque)
		}
	})
}
