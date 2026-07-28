package service_test

import (
	"pedidos/internal/domain/order"
	"testing"

	"github.com/google/uuid"
)

// TestOrderService_DomainRulesInService valida se as invariantes do DDD são respeitadas na orquestração
func TestOrderService_DomainRulesInService(t *testing.T) {
	// 1. Instancia um pedido de teste usando a fábrica de Domínio (Restaurando estado)
	pedidoID := uuid.New()
	clienteID := uuid.New()

	domainOrder := order.Restore(
		pedidoID,
		clienteID,
		order.StatusPending, // Estado ininial: PENDING
		[]order.OrderItem{},
	)

	//2. Valida a transação de estado pa PAID (caminho Feliz)
	t.Run("deve permitir pagar pedido pendente", func(t *testing.T) {
		err := domainOrder.Pay()
		if err != nil {
			t.Fatalf("esperava sucesso ao pagar pedido pendente, obteve erro: %v", err)
		}

		if domainOrder.Status() != order.StatusPaid {
			t.Errorf("esperava status PAID, mas obteve %s", domainOrder.Status())
		}
	})

	// 3. Valida a regra de proteção contra cancelamento de pedido pago (Caminho Triste)
	t.Run("nao deve permitir cancelar pedido que ja foi pago", func(t *testing.T) {
		// O pedido agora está PAID. A tentativa de cancelamento deve falhar no Domínio.
		err := domainOrder.Cancel()
		if err == nil {
			t.Error("esperava erro ao tentar cancelar pedido pago, mas obteve nil")
		}
	})
}

// TestOrderService_EmptyItemsValidation valida se a regra de carrinho vazio é bloqueada
func TestOrderService_EmptyItemsValidation(t *testing.T) {
	// Tenta criar um pedido sem itens
	_, err := order.NewOrder(uuid.New(), []order.OrderItem{})

	if err == nil {
		t.Errorf("esperava erro ao criar pedido sem itens na camada de serviço/domínio, mas obteve nil")
	}
}
