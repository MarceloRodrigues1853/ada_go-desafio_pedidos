package controllers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"pedidos/internal/controllers"

	"github.com/go-chi/chi/v5"
)

// TestOrderController_Create_InvalidUUID valida recusa de cliente_id malformado
func TestOrderController_Create_InvalidUUID(t *testing.T) {
	controller := controllers.NewOrderController(nil)

	payload := []byte(`{
		"cliente_id": "id-invalido-abc",
		"itens": []
	}`)

	req := httptest.NewRequest(http.MethodPost, "/pedidos", bytes.NewBuffer(payload))
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	controller.Create(rr, req)

	// Rota sem ID deve ser interceptada pelo roteador Chi ou retornar 400
	if rr.Code != http.StatusBadRequest {
		t.Errorf("esperava status 400 Bad Request para UUID invalido, obteve %d", rr.Code)
	}
}

// TestOrderController_Pay_InvalidURLParam valida recusa de acionar pagamento com ID quebrado
func TestOrderController_Pay_InvalidURLParam(t *testing.T) {
	controller := controllers.NewOrderController(nil)

	// Prepara a rota com parâmetro de URL falso via chi.Mux
	r := chi.NewRouter()
	r.Post("/pedidos/{id}/pagar", controller.Pay)

	req := httptest.NewRequest(http.MethodPost, "/pedidos/uuid-falso/pagar", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	// Rota sem ID deve ser interceptada pelo roteador Chi ou retornar 400
	if rr.Code != http.StatusBadRequest {
		t.Errorf("esperava status 400 Bad Request para ID de pedido invalido, obteve %d", rr.Code)
	}
}
