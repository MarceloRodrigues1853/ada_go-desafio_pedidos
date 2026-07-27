package domain_test

import (
	"testing"

	// Importa o pacote de domínio local onde o Cliente está definido
	"pedidos/internal/domain"
)

// TestNovoCliente_Sucesso valida a criação de um cliente válido e a geração do hash de senha
func TestNovoCliente_Sucesso(t *testing.T) {
	// Dados de entrada para o teste
	name := "Marcelo Rodrigues"
	email := "marcelo@example.com"
	password := "senha123"

	// Executa a função do domínio
	client, err := domain.NovoCliente(name, email, password)

	// Garante que a criação não retornou erro inesperado
	if err != nil {
		t.Fatalf("esperava sucesso ao criar cliente, mas recebeu erro: %v", err)
	}

	// Valida se o nome foi atribuído corretamente
	if client.Name != name {
		t.Errorf("esperava nome %q, recebeu %q", name, client.Name)
	}

	// Valida se o e-mail foi atribuído corretamente
	if client.Email != email {
		t.Errorf("esperava email %q, recebeu %q", email, client.Email)
	}

	// Valida se a senha NÃO foi salva em texto puro e se o hash foi gerado
	if client.PasswordHash == "" || client.PasswordHash == password {
		t.Error("esperava que a senha estivesse criptografada com hash do bcrypt")
	}
}

// TestNovoCliente_InvalidData testa as validações de dados obrigatórios do cliente
func TestNovoCliente_InvalidData(t *testing.T) {
	// Subteste 1: Nome vazio deve ser recusado
	t.Run("nome vazio deve retornar erro", func(t *testing.T) {
		_, err := domain.NovoCliente("", "email@example.com", "senha123")
		if err == nil {
			t.Error("esperava erro para nome vazio, mas obteve nil")
		}
	})

	// Subteste 2: E-mail vazio deve ser recusado
	t.Run("email vazio deve retornar erro", func(t *testing.T) {
		_, err := domain.NovoCliente("Marcelo", "", "senha123")
		if err == nil {
			t.Error("esperava erro para email vazio, mas obteve nil")
		}
	})

	// Subteste 3: Senha vazia deve ser recusada
	t.Run("senha vazia deve retornar erro", func(t *testing.T) {
		_, err := domain.NovoCliente("Marcelo", "marcelo@example.com", "")
		if err == nil {
			t.Error("esperava erro para senha vazia, mas obteve nil")
		}
	})
}
