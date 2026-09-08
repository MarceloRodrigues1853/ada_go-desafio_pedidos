package controllers

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"pedidos/internal/events"
	"pedidos/internal/payments"
	"pedidos/internal/service"

	"github.com/google/uuid"
)

type paymentSelectionService struct {
	selection service.PaymentSelection
	called    bool
}

func (s *paymentSelectionService) Create(context.Context, uuid.UUID, []service.OrderItemInput) (*service.OrderOutput, error) {
	return nil, nil
}

func (s *paymentSelectionService) CreateWithPayment(_ context.Context, _ uuid.UUID, _ []service.OrderItemInput, selection service.PaymentSelection) (*service.OrderOutput, error) {
	s.called = true
	s.selection = selection
	return &service.OrderOutput{ID: uuid.New(), Status: "PENDING"}, nil
}

func (s *paymentSelectionService) GetByID(context.Context, uuid.UUID) (*service.OrderOutput, error) {
	return nil, nil
}

func (s *paymentSelectionService) ListPaginado(context.Context, int32, int32) ([]service.OrderOutput, error) {
	return nil, nil
}

func (s *paymentSelectionService) Pay(context.Context, uuid.UUID) error { return nil }

func (s *paymentSelectionService) Cancel(context.Context, uuid.UUID) error { return nil }

func (s *paymentSelectionService) ProcessPaymentResult(context.Context, events.PaymentProcessedEvent) error {
	return nil
}

func (s *paymentSelectionService) ProcessPaymentFailure(context.Context, events.PaymentFailedEvent) error {
	return nil
}

func TestOrderControllerForwardsPaymentSimulation(t *testing.T) {
	serviceStub := &paymentSelectionService{}
	controller := NewOrderController(serviceStub)
	body := []byte(`{
		"cliente_id":"` + uuid.NewString() + `",
		"payment_method":"PIX",
		"simulation_outcome":"DECLINED",
		"itens":[{"produto_id":"SKU-1","quantidade":1}]
	}`)
	request := httptest.NewRequest(http.MethodPost, "/pedidos", bytes.NewReader(body))
	response := httptest.NewRecorder()

	controller.Create(response, request)

	if response.Code != http.StatusCreated {
		t.Fatalf("status = %d; esperado %d; body: %s", response.Code, http.StatusCreated, response.Body.String())
	}
	if !serviceStub.called {
		t.Fatal("CreateWithPayment() não foi chamado")
	}
	if serviceStub.selection.Method != payments.PaymentMethodPix {
		t.Errorf("método = %q; esperado %q", serviceStub.selection.Method, payments.PaymentMethodPix)
	}
	if serviceStub.selection.SimulationOutcome != payments.PaymentOutcomeDeclined {
		t.Errorf("resultado = %q; esperado %q", serviceStub.selection.SimulationOutcome, payments.PaymentOutcomeDeclined)
	}
}
