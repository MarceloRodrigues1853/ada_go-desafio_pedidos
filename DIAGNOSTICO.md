# Relatório de Auditoria de Arquitetura Go

## Diagnóstico Geral
O projeto encontra-se em um estado de transição entre um monólito modular e uma arquitetura orientada a eventos. A infraestrutura básica (RabbitMQ, Postgres) está configurada, mas a lógica de negócio ainda está centralizada no serviço de pedidos.

---

## Checklist de Fases

| Fase | Status | Observações |
| :--- | :--- | :--- |
| **FASE 2: Microsserviço Payments** | **NÃO IDENTIFICADO** | O código reside todo em `internal/`. Não há separação física de serviço ou `cmd/payments`. |
| **FASE 3: Mensageria** | **PARCIAL** | `RabbitMQPublisher` implementado, mas não há consumidores ou lógica de idempotência. |
| **FASE 4: Saga** | **NÃO IDENTIFICADO** | Não há orquestração ou fluxos de compensação implementados. |
| **FASE 5: Logs Estruturados** | **PARCIAL** | Uso de `slog` identificado no broker, mas não padronizado em toda a aplicação. |
| **FASE 6: Documentação** | **PARCIAL** | Documentação visual presente, mas falta documentação técnica de arquitetura (C4/Saga). |

---

## O que realmente falta
1. **Isolamento de Domínio:** O serviço de `Payments` precisa ser extraído para um novo diretório/projeto.
2. **Consumidores:** Implementar o `Consumer` no RabbitMQ para processar eventos de pagamento.
3. **Orquestração de Saga:** Definir o fluxo de estados do pedido (Pendente -> Pago -> Cancelado).
4. **Idempotência:** Mecanismo para evitar processamento duplicado de eventos (tabela `processed_events` ou similar).

---

## Próximo Passo de Implementação: Extração do Serviço de Pagamentos

O objetivo é preparar a base para a Saga, isolando o domínio de pagamentos.

### Tarefas:
1. **Criar estrutura de diretórios:**
   - Criar `cmd/payments/main.go`.
   - Mover lógica de domínio de pagamento (se houver) para `internal/payments/`.
2. **Configurar o novo serviço:**
   - Criar `internal/payments/service.go` e `internal/payments/handler.go`.
   - O serviço deve consumir eventos da fila `order_created`.
3. **Implementar o Consumidor:**
   - Criar `internal/infra/broker/consumer.go` para escutar eventos.

### Arquivos a criar/modificar:
- `cmd/payments/main.go`: Inicialização do serviço de pagamentos.
- `internal/infra/broker/consumer.go`: Lógica de `Consume` para o RabbitMQ.
- `internal/payments/handler.go`: Handler que processa o evento e publica `payment_approved` ou `payment_failed`.

### Teste de Validação:
- Criar `internal/payments/handler_test.go` simulando o recebimento de um evento de pedido e verificando a publicação de um evento de sucesso/falha.

**Nota do Auditor:** A arquitetura atual está acoplada. A prioridade máxima é a separação do `cmd/payments` para permitir que o sistema evolua para uma Saga distribuída. Não avance para a Saga sem antes ter o serviço de pagamentos rodando como um processo independente.