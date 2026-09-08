package payments

import (
	"context"
	"errors"
)

// PaymentMethod identifica o meio escolhido para o pagamento simulado.
type PaymentMethod string

const (
	PaymentMethodCard   PaymentMethod = "CARD"
	PaymentMethodPix    PaymentMethod = "PIX"
	PaymentMethodBoleto PaymentMethod = "BOLETO"
)

// PaymentOutcome representa a decisão produzida por um gateway de pagamento.
type PaymentOutcome string

const (
	PaymentOutcomeApproved PaymentOutcome = "APPROVED"
	PaymentOutcomeDeclined PaymentOutcome = "DECLINED"
)

var (
	ErrInvalidPaymentMethod  = errors.New("método de pagamento inválido")
	ErrInvalidPaymentAmount  = errors.New("valor do pagamento deve ser maior que zero")
	ErrInvalidPaymentOutcome = errors.New("resultado de pagamento inválido")
)

// PaymentInput contém somente os dados necessários para solicitar um pagamento.
// Dados sensíveis de cartão não fazem parte do protótipo.
type PaymentInput struct {
	Method PaymentMethod
	Amount float64
}

// PaymentResult contém a decisão retornada pelo gateway.
type PaymentResult struct {
	Method  PaymentMethod
	Amount  float64
	Outcome PaymentOutcome
	Reason  string
}

// PaymentGateway permite trocar o simulador por uma integração real no futuro.
type PaymentGateway interface {
	Process(ctx context.Context, input PaymentInput) (PaymentResult, error)
}

// FakePaymentGateway retorna um resultado previamente configurado. Isso mantém
// demonstrações e testes determinísticos, sem chamadas para serviços externos.
type FakePaymentGateway struct {
	outcome PaymentOutcome
	reason  string
}

// NewFakePaymentGateway cria um gateway com decisão controlada pelo chamador.
func NewFakePaymentGateway(outcome PaymentOutcome, reason string) *FakePaymentGateway {
	return &FakePaymentGateway{outcome: outcome, reason: reason}
}

// Process valida a solicitação e devolve a decisão configurada.
func (g *FakePaymentGateway) Process(ctx context.Context, input PaymentInput) (PaymentResult, error) {
	if err := ctx.Err(); err != nil {
		return PaymentResult{}, err
	}
	if !input.Method.IsValid() {
		return PaymentResult{}, ErrInvalidPaymentMethod
	}
	if input.Amount <= 0 {
		return PaymentResult{}, ErrInvalidPaymentAmount
	}
	if !g.outcome.IsValid() {
		return PaymentResult{}, ErrInvalidPaymentOutcome
	}

	return PaymentResult{
		Method:  input.Method,
		Amount:  input.Amount,
		Outcome: g.outcome,
		Reason:  g.reason,
	}, nil
}

// IsValid informa se o método pertence ao conjunto aceito pelo protótipo.
func (m PaymentMethod) IsValid() bool {
	switch m {
	case PaymentMethodCard, PaymentMethodPix, PaymentMethodBoleto:
		return true
	default:
		return false
	}
}

// IsValid informa se o resultado pode ser produzido pelo gateway simulado.
func (o PaymentOutcome) IsValid() bool {
	return o == PaymentOutcomeApproved || o == PaymentOutcomeDeclined
}
