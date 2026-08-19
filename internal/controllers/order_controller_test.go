package controllers_test

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"pedidos/internal/controllers"
	"pedidos/internal/repository"
	"pedidos/internal/repository/db"
	"pedidos/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestOrderController_Full(t *testing.T) {
	queries, pool := setupDB(t)
	defer pool.Close()

	srv := service.NewOrderService(repository.NewClientPostgresRepository(queries), repository.NewProductPostgresRepository(queries), repository.NewOrderPostgresRepository(queries, pool), nil)
	ctrl := controllers.NewOrderController(srv)

	r := chi.NewRouter()
	r.Post("/pedidos", ctrl.Create)
	r.Get("/pedidos", ctrl.ListPaginado)
	r.Get("/pedidos/{id}", ctrl.GetByID)
	r.Post("/pedidos/{id}/pagar", ctrl.Pay)
	r.Post("/pedidos/{id}/cancelar", ctrl.Cancel)

	// 1. Listar Paginado Sucesso (200 OK)
	t.Run("GET /pedidos - Sucesso Paginado", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pedidos?limit=5&offset=0", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("esperava 200 OK, obteve %d", rr.Code)
		}
	})

	// 2. Pedido Não Encontrado (404 Not Found)
	t.Run("GET /pedidos/{id} - Nao Encontrado (404)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pedidos/"+uuid.New().String(), nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("esperava 404 Not Found, obteve %d", rr.Code)
		}
	})

	// 3. Body invalido na criacao
	t.Run("POST /pedidos - Body Invalido (400)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/pedidos", bytes.NewBufferString("json-quebrado"))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("esperava 400 Bad Request, obteve %d", rr.Code)
		}
	})

	// 4. ID invalido ao pagar
	t.Run("POST /pedidos/{id}/pagar - ID Invalido (400)", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/pedidos/nao-e-uuid/pagar", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest { //retorna 400 Bad Request se o ID for invalido
			t.Errorf("esperava 400 Bad Request, obteve %d", rr.Code)
		}
	})

	// 5. Pagar pedido que não existe
	t.Run("POST /pedidos/{id}/pagar - Pedido inexistente", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/pedidos/"+uuid.New().String()+"/pagar", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusNotFound {
			t.Errorf("esperava 404, obteve %d", rr.Code)
		}
	})

	// 6. Pagar pedido com sucesso (200)
	t.Run("POST /pedidos/{id}/pagar - Sucesso (200)", func(t *testing.T) {
		ctx := context.Background()

		// --- PREPARAR: igual ao order_service_integration_test ---
		cliente, err := queries.CreateCliente(ctx, db.CreateClienteParams{
			Name:         "Cliente Ctrl Pagar",
			Email:        "ctrl_pay_" + uuid.New().String()[:8] + "@email.com",
			PasswordHash: "hash123",
		})
		if err != nil {
			t.Fatalf("falha ao criar cliente: %v", err)
		}

		produto, err := queries.CreateProduto(ctx, db.CreateProdutoParams{
			ID:      "PROD_PAY_" + uuid.New().String()[:5],
			Nome:    "Mouse Teste",
			Preco:   50.0,
			Estoque: 10,
		})
		if err != nil {
			t.Fatalf("falha ao criar produto: %v", err)
		}

		itens := []service.OrderItemInput{
			{
				ProductID: produto.ID,
				Quantity:  1,
			},
		}

		pedido, err := srv.Create(ctx, cliente.ID, itens)
		if err != nil {
			t.Fatalf("falha ao criar pedido: %v", err)
		}

		// --- AGIR: agora sim via HTTP (o que o controller faz) ---
		req := httptest.NewRequest(http.MethodPost, "/pedidos/"+pedido.ID.String()+"/pagar", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		// --- CONFERIR ---
		if rr.Code != http.StatusOK {
			t.Fatalf("esperava 200 OK, obteve %d, body: %s", rr.Code, rr.Body.String())
		}
	})

	// 7. Cancelar pedido com sucesso (200)
	t.Run("POST /pedidos/{id}/cancelar - Sucesso (200)", func(t *testing.T) {
		ctx := context.Background()

		// --- PREPARAR: igual ao order_service_integration_test ---
		cliente, err := queries.CreateCliente(ctx, db.CreateClienteParams{
			Name:         "Cliente Ctrl Cancelar",
			Email:        "ctrl_cancel_" + uuid.New().String()[:8] + "@email.com",
			PasswordHash: "hash123",
		})
		if err != nil {
			t.Fatalf("falha ao criar cliente: %v", err)
		}

		produto, err := queries.CreateProduto(ctx, db.CreateProdutoParams{
			ID:      "PROD_CANCEL_" + uuid.New().String()[:5],
			Nome:    "Mouse Teste",
			Preco:   50.0,
			Estoque: 10,
		})
		if err != nil {
			t.Fatalf("falha ao criar produto: %v", err)
		}

		itens := []service.OrderItemInput{
			{
				ProductID: produto.ID,
				Quantity:  1,
			},
		}

		pedido, err := srv.Create(ctx, cliente.ID, itens)
		if err != nil {
			t.Fatalf("falha ao criar pedido: %v", err)
		}

		// --- NÃO chama Pay aqui — pedido continua PENDING ---
		req := httptest.NewRequest(http.MethodPost, "/pedidos/"+pedido.ID.String()+"/cancelar", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		// --- CONFERIR ---
		if rr.Code != http.StatusOK {
			t.Fatalf("esperava 200 OK, obteve %d, body: %s", rr.Code, rr.Body.String())
		}
	})
}
