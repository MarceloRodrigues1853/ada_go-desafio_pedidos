package controllers_test

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"pedidos/internal/controllers"
	"pedidos/internal/repository/db"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

// MockOrderService simula a camada de serviço em memória
type MockOrderService struct {
	ShouldFail bool
}

func (m *MockOrderService) Create(ctx context.Context, clienteID uuid.UUID, itens []db.CreateItemPedidoParams) (*db.Pedido, error) {
	if m.ShouldFail {
		return nil, errors.New("estoque insuficiente")
	}
	return &db.Pedido{ID: uuid.New(), ClienteID: clienteID, Status: "PENDING"}, nil
}

func (m *MockOrderService) GetByID(ctx context.Context, pedidoID uuid.UUID) (*db.Pedido, error) {
	if m.ShouldFail {
		return nil, errors.New("pedido não encontrado")
	}
	return &db.Pedido{ID: pedidoID, Status: "PENDING"}, nil
}

func (m *MockOrderService) ListPaginado(ctx context.Context, limit, offset int32) ([]db.Pedido, error) {
	return []db.Pedido{}, nil
}

func (m *MockOrderService) Pay(ctx context.Context, pedidoID uuid.UUID) error {
	if m.ShouldFail {
		return errors.New("pedido já pago")
	}
	return nil
}

func (m *MockOrderService) Cancel(ctx context.Context, pedidoID uuid.UUID) error {
	if m.ShouldFail {
		return errors.New("pedido já cancelado")
	}
	return nil
}

// Testes Unitários das Rotas do OrderController
func TestOrderController_Unit_SuccessAndConflict(t *testing.T) {
	mockService := &MockOrderService{ShouldFail: false}
	ctrl := controllers.NewOrderController(mockService)

	r := chi.NewRouter()
	r.Post("/pedidos", ctrl.Create)
	r.Get("/pedidos/{id}", ctrl.GetByID)
	r.Post("/pedidos/{id}/pagar", ctrl.Pay)

	// 1. Testa Sucesso em GET /pedidos/{id} (200 OK)
	t.Run("GET /pedidos/{id} - Sucesso", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/pedidos/"+uuid.New().String(), nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Esperava 200 OK, obteve %d", rr.Code)
		}
	})

	// 2. Testa Sucesso em POST /pedidos/{id}/pagar (200 OK)
	t.Run("POST /pedidos/{id}/pagar - Sucesso", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/pedidos/"+uuid.New().String()+"/pagar", nil)
		rr := httptest.NewRecorder()
		r.ServeHTTP(rr, req)

		if rr.Code != http.StatusOK {
			t.Errorf("Esperava 200 OK, obteve %d", rr.Code)
		}
	})

	// 3. Testa Erro de Regra em POST /pedidos/{id}/pagar (409 Conflict)
	t.Run("POST /pedidos/{id}/pagar - Regra Inválida (409)", func(t *testing.T) {
		mockServiceFail := &MockOrderService{ShouldFail: true}
		ctrlFail := controllers.NewOrderController(mockServiceFail)

		rFail := chi.NewRouter()
		rFail.Post("/pedidos/{id}/pagar", ctrlFail.Pay)

		req := httptest.NewRequest(http.MethodPost, "/pedidos/"+uuid.New().String()+"/pagar", nil)
		rr := httptest.NewRecorder()
		rFail.ServeHTTP(rr, req)

		if rr.Code != http.StatusConflict {
			t.Errorf("Esperava 409 Conflict, obteve %d", rr.Code)
		}
	})
}
