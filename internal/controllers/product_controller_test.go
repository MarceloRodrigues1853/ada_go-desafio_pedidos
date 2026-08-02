package controllers_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"pedidos/internal/controllers"
	"pedidos/internal/repository/db"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestProductController_Full(t *testing.T) {
	queries, pool := setupDB(t)
	defer pool.Close()

	ctrl := controllers.NewProductController(queries)
	r := chi.NewRouter()
	r.Post("/produtos", ctrl.Create)
	r.Get("/produtos", ctrl.List)
	r.Get("/produtos/{id}", ctrl.GetByID)

	prodID := "PROD_CTRL_" + uuid.New().String()[:5]

	// 1. Criar Produto Sucesso (201)
	t.Run("POST /produtos - Sucesso", func(t *testing.T) {
		payload := db.CreateProdutoParams{
			ID:      prodID,
			Nome:    "Mouse Gamer Teste",
			Preco:   150.0,
			Estoque: 20,
		}
		body, _ := json.Marshal(payload)

		req := httptest.NewRequest(http.MethodPost, "/produtos", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusCreated {
			t.Fatalf("esperava 201 Created, obteve %d", rr.Code)
		}
	})

	// 2. Listar Produtos (200 OK)
	t.Run("GET /produtos - Sucesso", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/produtos", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("esperava 200 OK, obteve %d", rr.Code)
		}
	})

	// 3. Buscar Produto por ID Sucesso (200 OK)
	t.Run("GET /produtos/{id} - Sucesso", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/produtos/"+prodID, nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("esperava 200 OK, obteve %d", rr.Code)
		}
	})
}
