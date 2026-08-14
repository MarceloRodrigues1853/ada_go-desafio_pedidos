package controllers_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"pedidos/internal/controllers"
	"pedidos/internal/repository"
	"pedidos/internal/repository/db"
	"pedidos/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func setupDB(t *testing.T) (*db.Queries, *pgxpool.Pool) {
	_ = godotenv.Load("../../.env")
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		t.Skip("DB_URL não configurada no .env")
	}

	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		t.Fatalf("Erro ao conectar no banco: %v", err)
	}

	return db.New(pool), pool
}

func TestClientController_Full(t *testing.T) {
	queries, pool := setupDB(t)
	defer pool.Close()

	ctrl := controllers.NewClientController(service.NewClientService(repository.NewClientPostgresRepository(queries)))
	r := chi.NewRouter()
	r.Post("/clientes", ctrl.Create)
	r.Get("/clientes", ctrl.List)
	r.Get("/clientes/{id}", ctrl.GetByID)

	// 1. Bad Request
	t.Run("POST /clientes - Body Invalido", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/clientes", bytes.NewBufferString("bad-json"))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)
		if rr.Code != http.StatusBadRequest {
			t.Errorf("esperava 400, obteve %d", rr.Code)
		}
	})

	// 2. Criar Cliente Sucesso (201)
	emailUnico := "teste_ctrl_" + uuid.New().String()[:8] + "@email.com"
	var clienteCriado db.Cliente

	t.Run("POST /clientes - Sucesso", func(t *testing.T) {
		payload := map[string]string{
			"name":     "Marcelo Teste",
			"email":    emailUnico,
			"password": "senha123_segura",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/clientes", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("esperava 201 Created, obteve %d", rr.Code)
		}

		_ = json.NewDecoder(rr.Body).Decode(&clienteCriado)
	})

	// 3. Email Duplicado (409 Conflict)
	t.Run("POST /clientes - Email Duplicado (409)", func(t *testing.T) {
		payload := map[string]string{
			"name":     "Marcelo Duplicado",
			"email":    emailUnico,
			"password": "senha123_segura",
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/clientes", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusConflict {
			t.Errorf("esperava 409 Conflict, obteve %d", rr.Code)
		}
	})

	// 4. Listar Clientes (200 OK)
	t.Run("GET /clientes - Sucesso", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/clientes", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("esperava 200 OK, obteve %d", rr.Code)
		}
	})

	// 5. Buscar Cliente por ID Sucesso (200 OK)
	t.Run("GET /clientes/{id} - Sucesso", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/clientes/"+clienteCriado.ID.String(), nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("esperava 200 OK, obteve %d", rr.Code)
		}
	})

	// 6. UUID Invalido (400 Bad Request)
	t.Run("GET /clientes/{id} - UUID Invalido", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/clientes/id-falso", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("esperava 400, obteve %d", rr.Code)
		}
	})
}
