# Relatório de Auditoria de Arquitetura Go

**Data:** 22/05/2024  
**Auditor:** Tech Lead & Arquiteto Go  
**Status:** Auditoria de Fases 2-6

---

## 1. Diagnóstico Geral

O projeto apresenta uma estrutura sólida em Go, utilizando `sqlc` para persistência e `chi` para roteamento. A infraestrutura de mensageria (RabbitMQ) está presente e funcional para publicação. No entanto, a implementação da Saga e a separação física dos serviços ainda carecem de maturidade operacional.

### Checklist de Fases

| Fase | Status | Observações |
| :--- | :--- | :--- |
| **FASE 2: Microsserviço Payments** | **PARCIAL** | Existe `cmd/payments`, mas a separação de dependências e o isolamento do schema de banco não estão totalmente segregados. |
| **FASE 3: Mensageria** | **IMPLEMENTADO** | Publisher funcional em `internal/infra/broker`. Consumo ainda precisa ser validado em produção. |
| **FASE 4: Saga** | **PARCIAL** | Existe um `saga_test.go`, mas a orquestração real entre `orders` e `payments` não está integrada no fluxo principal. |
| **FASE 5: Logs Estruturados** | **PARCIAL** | Uso de `slog` iniciado, mas falta padronização de `correlation_id` em todo o fluxo. |
| **FASE 6: Documentação** | **PARCIAL** | Documentação visual (assets) presente, mas falta documentação técnica de arquitetura (ADR). |

---

## 2. O que realmente falta

1.  **Isolamento de Schema:** O serviço de `payments` deve apontar para um schema ou banco de dados distinto, não apenas compartilhar o mesmo banco de `pedidos`.
2.  **Orquestrador de Saga:** Falta um componente que gerencie o estado da transação distribuída (ex: `Pending` -> `Paid` ou `Failed` -> `Compensated`).
3.  **Idempotência:** Não há verificação de `idempotency_key` no consumidor de eventos de pagamento.
4.  **Correlation ID:** O `slog` não está sendo injetado via Middleware para rastrear requisições entre serviços.

---

## 3. Próximo Passo de Implementação

O foco imediato é a **implementação da Saga de Pagamento com Idempotência**.

### Tarefas Detalhadas:

1.  **Middleware de Correlation ID:**
    *   Criar `internal/infra/middleware/correlation.go`.
    *   Capturar `X-Correlation-ID` do header e injetar no `context.Context`.
    *   Configurar `slog` para extrair esse valor do contexto em cada log.

2.  **Refatoração do Handler de Pagamento:**
    *   No arquivo `internal/payments/handler.go`, implementar a verificação de idempotência antes de processar o pagamento.
    *   Adicionar campo `status` na tabela de pagamentos para suportar a transição de estados da Saga.

3.  **Teste de Integração da Saga:**
    *   Criar um teste em `internal/service/saga_integration_test.go` que simule:
        1. Criação de pedido (Status: PENDING).
        2. Publicação de evento de pagamento.
        3. Consumo do evento e atualização do pedido (Status: PAID).
        4. Simulação de falha e execução da compensação (Status: CANCELLED).

### Arquivos a criar/modificar:
*   `internal/infra/middleware/correlation.go` (Novo)
*   `internal/payments/handler.go` (Modificar para incluir lógica de idempotência)
*   `internal/service/saga_integration_test.go` (Novo)

**Ação imediata:** Iniciar pelo Middleware de Correlation ID para garantir rastreabilidade antes de avançar na lógica complexa da Saga.