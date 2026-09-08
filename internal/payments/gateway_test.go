package payments

import (
	"context"
	"errors"
	"testing"
)

func TestFakePaymentGatewayApprovesSupportedMethods(t *testing.T) {
	tests := []struct {
		name   string
		method PaymentMethod
	}{
		{name: "cartão", method: PaymentMethodCard},
		{name: "pix", method: PaymentMethodPix},
		{name: "boleto", method: PaymentMethodBoleto},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gateway := NewFakePaymentGateway(PaymentOutcomeApproved, "")

			result, err := gateway.Process(context.Background(), PaymentInput{
				Method: tt.method,
				Amount: 150.90,
			})
			if err != nil {
				t.Fatalf("Process() retornou erro: %v", err)
			}
			if result.Method != tt.method {
				t.Errorf("método = %q; esperado %q", result.Method, tt.method)
			}
			if result.Outcome != PaymentOutcomeApproved {
				t.Errorf("resultado = %q; esperado %q", result.Outcome, PaymentOutcomeApproved)
			}
			if result.Amount != 150.90 {
				t.Errorf("valor = %.2f; esperado 150.90", result.Amount)
			}
		})
	}
}

func TestFakePaymentGatewayDeclinesPayment(t *testing.T) {
	gateway := NewFakePaymentGateway(PaymentOutcomeDeclined, "pagamento recusado na simulação")

	result, err := gateway.Process(context.Background(), PaymentInput{
		Method: PaymentMethodCard,
		Amount: 300,
	})
	if err != nil {
		t.Fatalf("Process() retornou erro: %v", err)
	}
	if result.Outcome != PaymentOutcomeDeclined {
		t.Errorf("resultado = %q; esperado %q", result.Outcome, PaymentOutcomeDeclined)
	}
	if result.Reason != "pagamento recusado na simulação" {
		t.Errorf("motivo = %q; esperado motivo configurado", result.Reason)
	}
}

func TestFakePaymentGatewayUsesOutcomeRequestedBySimulation(t *testing.T) {
	gateway := NewFakePaymentGateway(PaymentOutcomeApproved, "")

	result, err := gateway.Process(context.Background(), PaymentInput{
		Method:            PaymentMethodPix,
		Amount:            80,
		SimulationOutcome: PaymentOutcomeDeclined,
	})
	if err != nil {
		t.Fatalf("Process() retornou erro: %v", err)
	}
	if result.Outcome != PaymentOutcomeDeclined {
		t.Errorf("resultado = %q; esperado %q", result.Outcome, PaymentOutcomeDeclined)
	}
}

func TestFakePaymentGatewayRejectsInvalidMethod(t *testing.T) {
	gateway := NewFakePaymentGateway(PaymentOutcomeApproved, "")

	_, err := gateway.Process(context.Background(), PaymentInput{
		Method: PaymentMethod("CASH"),
		Amount: 100,
	})
	if !errors.Is(err, ErrInvalidPaymentMethod) {
		t.Fatalf("erro = %v; esperado %v", err, ErrInvalidPaymentMethod)
	}
}

func TestFakePaymentGatewayRejectsInvalidAmount(t *testing.T) {
	gateway := NewFakePaymentGateway(PaymentOutcomeApproved, "")

	_, err := gateway.Process(context.Background(), PaymentInput{
		Method: PaymentMethodPix,
		Amount: 0,
	})
	if !errors.Is(err, ErrInvalidPaymentAmount) {
		t.Fatalf("erro = %v; esperado %v", err, ErrInvalidPaymentAmount)
	}
}
