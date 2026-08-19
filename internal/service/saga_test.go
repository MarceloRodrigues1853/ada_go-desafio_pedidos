package service_test

import (
	"context"
	"os"
	"testing"
	"time"

	"pedidos/internal/events"
	"pedidos/internal/repository"
	"pedidos/internal/repository/db"
	"pedidos/internal/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func setupSagaTestDB(t *testing.T) (*service.OrderService, *db.Queries, *pgxpool.Pool) {
	_ = godotenv.Load("../../.env")
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		t.Skip("DB_URL não configurada no arquivo .env. Pulando teste de integração da Saga.")
	}

	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		t.Fatalf("Erro ao conectar no banco de testes: %v", err)
	}

	queries := db.New(pool)
	srv := service.NewOrderService(
		repository.NewClientPostgresRepository(queries),
		repository.NewProductPostgresRepository(queries),
		repository.NewOrderPostgresRepository(queries, pool),
		nil,
	)

	return srv, queries, pool
}

func TestSaga_FluxoFeliz_Aprovacao(t *testing.T) {
	srv, queries, pool := setupSagaTestDB(t)
	defer pool.Close()

	ctx := context.Background()

	// 1. Cria cliente e produto
	cliente, err := queries.CreateCliente(ctx, db.CreateClienteParams{
		Name:         "Cliente Saga Feliz",
		Email:        "saga_feliz_" + uuid.New().String()[:8] + "@email.com",
		PasswordHash: "hash123",
	})
	if err != nil {
		t.Fatalf("Falha ao criar cliente: %v", err)
	}

	prodID := "PROD_FELIZ_" + uuid.New().String()[:5]
	_, err = queries.CreateProduto(ctx, db.CreateProdutoParams{
		ID:      prodID,
		Nome:    "Produto Saga Feliz",
		Preco:   50.0,
		Estoque: 10,
	})
	if err != nil {
		t.Fatalf("Falha ao criar produto: %v", err)
	}

	// 2. Cria pedido (status PENDING, estoque reduzido para 8)
	pedido, err := srv.Create(ctx, cliente.ID, []service.OrderItemInput{
		{ProductID: prodID, Quantity: 2},
	})
	if err != nil {
		t.Fatalf("Falha ao criar pedido: %v", err)
	}

	if pedido.Status != "PENDING" {
		t.Errorf("Esperava status PENDING, obteve %s", pedido.Status)
	}

	// 3. Simula evento de pagamento aprovado (PaymentProcessedEvent com PAID)
	paymentEvent := events.PaymentProcessedEvent{
		SagaID:      uuid.New(),
		OrderID:     pedido.ID,
		ClientID:    cliente.ID,
		TotalAmount: 100.0,
		Status:      "PAID",
		ProcessedAt: time.Now(),
	}

	err = srv.ProcessPaymentResult(ctx, paymentEvent)
	if err != nil {
		t.Fatalf("ProcessPaymentResult falhou: %v", err)
	}

	// 4. Valida se o status final do pedido é PAID
	pedidoAtualizado, err := srv.GetByID(ctx, pedido.ID)
	if err != nil {
		t.Fatalf("GetByID falhou: %v", err)
	}

	if pedidoAtualizado.Status != "PAID" {
		t.Errorf("Esperava status final PAID na SAGA, obteve %s", pedidoAtualizado.Status)
	}
}

func TestSaga_FluxoCompensacao_Falha(t *testing.T) {
	srv, queries, pool := setupSagaTestDB(t)
	defer pool.Close()

	ctx := context.Background()

	// 1. Cria cliente e produto com estoque inicial 5
	cliente, err := queries.CreateCliente(ctx, db.CreateClienteParams{
		Name:         "Cliente Saga Falha",
		Email:        "saga_falha_" + uuid.New().String()[:8] + "@email.com",
		PasswordHash: "hash123",
	})
	if err != nil {
		t.Fatalf("Falha ao criar cliente: %v", err)
	}

	prodID := "PROD_FALHA_" + uuid.New().String()[:5]
	_, err = queries.CreateProduto(ctx, db.CreateProdutoParams{
		ID:      prodID,
		Nome:    "Produto Saga Falha",
		Preco:   200.0,
		Estoque: 5,
	})
	if err != nil {
		t.Fatalf("Falha ao criar produto: %v", err)
	}

	// 2. Cria pedido (estoque reduzido de 5 para 2, status PENDING)
	pedido, err := srv.Create(ctx, cliente.ID, []service.OrderItemInput{
		{ProductID: prodID, Quantity: 3},
	})
	if err != nil {
		t.Fatalf("Falha ao criar pedido: %v", err)
	}

	prodAntesCompensacao, _ := queries.GetProduto(ctx, prodID)
	if prodAntesCompensacao.Estoque != 2 {
		t.Errorf("Esperava estoque 2 antes da compensação, obteve %d", prodAntesCompensacao.Estoque)
	}

	// 3. Simula evento de pagamento rejeitado (PaymentProcessedEvent com FAILED)
	paymentEvent := events.PaymentProcessedEvent{
		SagaID:      uuid.New(),
		OrderID:     pedido.ID,
		ClientID:    cliente.ID,
		TotalAmount: 600.0,
		Status:      "FAILED",
		ProcessedAt: time.Now(),
	}

	err = srv.ProcessPaymentResult(ctx, paymentEvent)
	if err != nil {
		t.Fatalf("ProcessPaymentResult (compensação) falhou: %v", err)
	}

	// 4. Valida se o pedido foi alterado para CANCELED e o estoque foi estornado para 5
	pedidoAtualizado, err := srv.GetByID(ctx, pedido.ID)
	if err != nil {
		t.Fatalf("GetByID falhou: %v", err)
	}

	if pedidoAtualizado.Status != "CANCELED" {
		t.Errorf("Esperava status CANCELED após compensação, obteve %s", pedidoAtualizado.Status)
	}

	prodAposCompensacao, err := queries.GetProduto(ctx, prodID)
	if err != nil {
		t.Fatalf("GetProduto falhou: %v", err)
	}

	if prodAposCompensacao.Estoque != 5 {
		t.Errorf("Esperava estoque estornado para 5 após compensação, obteve %d", prodAposCompensacao.Estoque)
	}
}
