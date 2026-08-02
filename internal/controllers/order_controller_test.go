package controllers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"pedidos/internal/controllers"
	"pedidos/internal/service"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func TestOrderController_Full(t *testing.T) {
	queries, pool := setupDB(t)
	defer pool.Close()

	srv := service.NewOrderService(queries, pool)
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
}
