# CLAUDE.md - Diretrizes do Projeto Go (DDD & TDD)

## Arquitetura do Projeto
Este projeto segue Clean Architecture e Domain-Driven Design (DDD) em Go.

### Estrutura de Camadas
- `internal/domain/order/`: Entidades puras, Value Objects e Erros de Domínio. **(NÃO deve importar pgx, sqlc ou pacotes de banco de dados)**.
- `internal/service/`: Casos de uso e orquestração.
- `internal/repository/db/`: Código de persistência gerado pelo sqlc e conexões Postgres.
- `internal/controllers/`: Handlers HTTP.

## Regras de Desenvolvimento (TDD Primeiro)
1. **Sempre escreva ou atualize os testes primeiro (TDD)** na camada `internal/domain/...` antes de finalizar a implementação.
2. A camada de domínio deve ser Go puro, usando apenas a biblioteca padrão e `github.com/google/uuid`.
3. Erros de domínio devem ser explícitos e exportados (ex: `var ErrCannotCancelPaidOrder = errors.New(...)`).
4. Entidades do domínio não expõem alteração direta de estado via campos públicos. A alteração deve ocorrer via métodos com intenções de negócio (ex: `order.Pay()`, `order.Cancel()`).

## Comandos Úteis
- Rodar testes de domínio: `go test -v ./internal/domain/...`
- Rodar todos os testes: `go test -v ./...`