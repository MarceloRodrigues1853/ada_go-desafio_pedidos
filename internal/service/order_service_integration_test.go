package service_test

import (
	"context"
	"os"
	"testing"

	"pedidos/internal/repository/db"
	"pedidos/internal/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

// setupTestDB conecta ao banco de dados e retorna o serviço pronto
func setupTestDB(t *testing.T) (*service.OrderService, *db.Queries, *pgxpool.Pool) {
	_ = godotenv.Load("../../.env")
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		t.Skip("DB_URL não configurada no arquivo .env. Pulando teste de integração do Service.")
	}

	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		t.Fatalf("Erro ao conectar no banco de testes: %v", err)
	}

	queries := db.New(pool)
	srv := service.NewOrderService(queries, pool)

	return srv, queries, pool
}

func TestOrderService_Integration_CreateAndCancel(t *testing.T) {
	srv, queries, pool := setupTestDB(t)
	defer pool.Close()

	ctx := context.Background()

	// 1. Prepara massa de dados de teste (Cliente e Produto)
	cliente, err := queries.CreateCliente(ctx, db.CreateClienteParams{
		Name:         "Cliente Teste Service",
		Email:        "teste_service_" + uuid.New().String()[:8] + "@email.com",
		PasswordHash: "hash123",
	})
	if err != nil {
		t.Fatalf("Falha ao criar cliente para o teste: %v", err)
	}

	produtoID := "PROD_TEST_" + uuid.New().String()[:5]
	produto, err := queries.CreateProduto(ctx, db.CreateProdutoParams{
		ID:      produtoID,
		Nome:    "Teclado Teste",
		Preco:   100.0,
		Estoque: 10,
	})
	if err != nil {
		t.Fatalf("Falha ao criar produto para o teste: %v", err)
	}

	// 2. Testa a CRIAÇÃO DO PEDIDO (Diminui estoque na transação)
	itens := []db.CreateItemPedidoParams{
		{
			ProdutoID:     produto.ID,
			Quantidade:    2,
			PrecoUnitario: produto.Preco,
		},
	}

	pedidoSalvo, err := srv.Create(ctx, cliente.ID, itens)
	if err != nil {
		t.Fatalf("Esperava sucesso na criação do pedido, obteve erro: %v", err)
	}

	if pedidoSalvo.Status != "PENDING" {
		t.Errorf("Esperava status PENDING, obteve: %s", pedidoSalvo.Status)
	}

	// Verifica se o estoque foi reduzido para 8
	prodAtualizado, _ := queries.GetProduto(ctx, produto.ID)
	if prodAtualizado.Estoque != 8 {
		t.Errorf("Esperava estoque igual a 8 após compra, obteve: %d", prodAtualizado.Estoque)
	}

	// 3. Testa o PAGAMENTO DO PEDIDO
	err = srv.Pay(ctx, pedidoSalvo.ID)
	if err != nil {
		t.Fatalf("Esperava sucesso ao pagar pedido, obteve erro: %v", err)
	}

	// 4. Tenta CANCELAR pedido PAGO (Deve falhar pela Regra de Negócio)
	err = srv.Cancel(ctx, pedidoSalvo.ID)
	if err == nil {
		t.Error("Esperava erro ao tentar cancelar pedido já pago, mas obteve nil")
	}
}
