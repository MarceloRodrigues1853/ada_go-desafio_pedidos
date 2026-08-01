package controllers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"pedidos/internal/controllers"

	"github.com/go-chi/chi/v5"
)

// TestProductController_Create_InvalidBody valida o retorno 400 para payload incorreto
func TestProductController_Create_InvalidBody(t *testing.T) {
	controller := controllers.NewProductController(nil)

	body := []byte(`{"nome": "Produto Sem Preco"`) // JSON quebrado
	req := httptest.NewRequest(http.MethodPost, "/produtos", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	controller.Create(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("esperava status 400 Bad Request para JSON quebrado, obteve %d", rr.Code)
	}
}

// TestProductController_GetByID_EmptyID valida requisições sem o parâmetro correto
func TestProductController_GetByID_NotFoundSimulation(t *testing.T) {
	controller := controllers.NewProductController(nil)

	// Prepara a rota com parâmetro de URL falso via chi.Mux
	r := chi.NewRouter()
	r.Get("/produtos/{id}", controller.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/produtos/", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	// Rota sem ID deve ser interceptada pelo roteador Chi ou retornar 404
	if rr.Code != http.StatusNotFound {
		t.Errorf("esperava status 404 Not Found, obteve %d", rr.Code)
	}
}
