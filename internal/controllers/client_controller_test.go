package controllers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"pedidos/internal/controllers"

	"github.com/go-chi/chi/v5"
)

// TestClientController_Create_InvalidJSON valida a resposta 400 ao enviar um JSON malformado
func TestClientController_Create_InvalidJSON(t *testing.T) {
	// 1. Instancia o controller passando nil para as queries (pois a validação ocorre antes de chamar o banco)
	controller := controllers.NewClientController(nil)

	// 2. Prepara uma requisição HTTP com JSON inválido no corpo
	jsonInvalido := []byte(`{"name": "Marcelo", "email":}`)
	req := httptest.NewRequest(http.MethodPost, "/clientes", bytes.NewBuffer(jsonInvalido))
	req.Header.Set("Content-Type", "application/json")

	// 3. Cria o ResponseRecorder para capturar a resposta
	rr := httptest.NewRecorder()

	// 4. Executa a função do controller diretamente
	controller.Create(rr, req)

	// 5. Valida se o status retornado foi 400 Bad Request
	if rr.Code != http.StatusBadRequest {
		t.Errorf("esperava status 400 Bad Request, mas recebeu %d", rr.Code)
	}
}

// TestClientController_GetByID_InvalidUUID valida a resposta 400 ao passar um ID em formato inválido
func TestClientController_GetByID_InvalidUUID(t *testing.T) {
	controller := controllers.NewClientController(nil)

	// Prepara a rota com parâmetro de URL falso via chi.Mux
	r := chi.NewRouter()
	r.Get("/clientes/{id}", controller.GetByID)

	req := httptest.NewRequest(http.MethodGet, "/clientes/id-invalido-123", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	// Rota sem ID deve ser interceptada pelo roteador Chi ou retornar 400
	if rr.Code != http.StatusBadRequest {
		t.Errorf("esperava status 400 para UUID invalido, mas recebeu %d", rr.Code)
	}
}
