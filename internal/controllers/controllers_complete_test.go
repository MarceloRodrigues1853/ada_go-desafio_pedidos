package controllers_test

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"pedidos/internal/controllers"

	"github.com/go-chi/chi/v5"
)

// TestControllers_ValidationsCoverage testa exaustivamente as checagens e erros de payload/rotas
func TestControllers_ValidationsCoverage(t *testing.T) {
	// Instancia controladores com nil para validar a camada de borda HTTP
	clientCtrl := controllers.NewClientController(nil)
	productCtrl := controllers.NewProductController(nil)
	orderCtrl := controllers.NewOrderController(nil)

	r := chi.NewRouter()

	// Configuração de rotas no roteador Chi
	r.Post("/clientes", clientCtrl.Create)
	r.Get("/clientes/{id}", clientCtrl.GetByID)

	r.Post("/produtos", productCtrl.Create)
	r.Get("/produtos/{id}", productCtrl.GetByID)

	r.Post("/pedidos", orderCtrl.Create)
	r.Get("/pedidos/{id}", orderCtrl.GetByID)
	r.Post("/pedidos/{id}/pagar", orderCtrl.Pay)
	r.Post("/pedidos/{id}/cancelar", orderCtrl.Cancel)

	// 1. Validação de UUID inválido na busca de cliente por ID
	t.Run("GET /clientes/{id} - UUID malformado deve retornar 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/clientes/123-id-falso", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("esperava 400 Bad Request, obteve %d", rr.Code)
		}
	})

	// 2. Validação de JSON quebrado na criação de cliente
	t.Run("POST /clientes - Body incorreto deve retornar 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/clientes", bytes.NewBufferString("{name: sem_aspas}"))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("esperava 400 Bad Request, obteve %d", rr.Code)
		}
	})

	// 3. Validação de JSON quebrado na criação de produto
	t.Run("POST /produtos - Payload invalido deve retornar 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/produtos", bytes.NewBufferString("not_json"))
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("esperava 400 Bad Request, obteve %d", rr.Code)
		}
	})

	// 4. Validação de UUID de pedido inválido na busca por ID
	t.Run("GET /pedidos/{id} - UUID malformado deve retornar 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pedidos/abc-xyz", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("esperava 400 Bad Request, obteve %d", rr.Code)
		}
	})

	// 5. Validação de UUID de pedido inválido na rota de cancelamento
	t.Run("POST /pedidos/{id}/cancelar - ID invalido deve retornar 400", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/pedidos/id-invalido/cancelar", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusBadRequest {
			t.Errorf("esperava 400 Bad Request, obteve %d", rr.Code)
		}
	})
}
