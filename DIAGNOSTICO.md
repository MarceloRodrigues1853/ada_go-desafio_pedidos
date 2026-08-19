# Relatório de Auditoria de Arquitetura Go

## 1. Diagnóstico Geral
O projeto apresenta uma estrutura sólida de microsserviços com separação clara de responsabilidades (`cmd/app` vs `cmd/payments`). A infraestrutura de mensageria (RabbitMQ) está integrada e o serviço de pagamentos já possui um esqueleto funcional de consumo e publicação.

### Checklist de Fases

| Fase | Status | Observações |
| :--- | :--- | :--- |
| **FASE 2: Microsserviço Payments** | **IMPLEMENTADO** | Serviço isolado em `cmd/payments`, com lógica de handler e service. |
| **FASE 3: Mensageria** | **PARCIAL** | Broker configurado, mas a lógica de **idempotência** no consumidor de pagamentos não está explícita. |
| **FASE 4: Saga** | **NÃO IDENTIFICADO** | Fluxos de compensação (ex: falha no pagamento -> cancelar pedido) não estão implementados. |
| **FASE 5: Logs Estruturados** | **PARCIAL** | `slog` configurado no `main.go`, mas falta propagação consistente de `correlation_id` ou `saga_id` entre serviços. |
| **FASE 6: Documentação** | **PARCIAL** | Existem assets de testes, mas falta documentação técnica da Saga. |

---

## 2. O que realmente falta
1.  **Idempotência:** O `PaymentHandler` precisa verificar se o `order_id` já foi processado para evitar pagamentos duplicados.
2.  **Orquestração da Saga:** O serviço de pedidos (`app`) precisa escutar eventos de `PaymentProcessed` ou `PaymentFailed` para atualizar o status do pedido.
3.  **Contexto de Saga:** Implementar um middleware ou decorator que injete um `SagaID` nos headers das mensagens RabbitMQ para rastreabilidade.
4.  **Compensação:** Implementar o fluxo de cancelamento de pedido caso o pagamento falhe.

---

## 3. Próximo Passo de Implementação: Idempotência e Saga-Feedback

O foco imediato é garantir que o microsserviço de pagamentos não processe o mesmo pedido duas vezes e que o serviço de pedidos saiba o resultado do pagamento.

### Tarefas Detalhadas:

1.  **Implementar Idempotência no `internal/payments/handler.go`**:
    *   Adicionar uma tabela `processed_payments` no banco de dados (ou verificar status no serviço de pedidos via API/DB).
    *   Modificar `HandleMessage` para verificar existência antes de processar.

2.  **Implementar Consumidor de Resposta no `cmd/app/main.go`**:
    *   O serviço de pedidos deve escutar o tópico de eventos de pagamento.
    *   Criar `internal/infra/broker/consumer_app.go` para processar `PaymentProcessedEvent` e `PaymentFailedEvent`.

3.  **Atualizar `internal/service/order_service.go`**:
    *   Adicionar método `UpdateOrderStatus(ctx, orderID, status)`.

### Arquivos a criar/modificar:
*   **Modificar:** `internal/payments/handler.go` (Adicionar lógica de verificação).
*   **Criar:** `internal/events/payment_events.go` (Definir structs de eventos de resposta).
*   **Modificar:** `cmd/app/main.go` (Adicionar o consumidor de eventos de pagamento).

### Teste de Validação:
*   Criar `internal/payments/handler_idempotency_test.go`: Simular o recebimento da mesma mensagem duas vezes e garantir que o serviço de pagamento só execute a lógica de negócio uma vez.

---
**Auditor:** Tech Lead Go
**Status:** Auditoria concluída. Aguardando execução do próximo passo.