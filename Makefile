# Carrega as variáveis do .env
include .env

# O comando 'run' liga o servidor
run:
	go run cmd/app/main.go

# Testes rápidos, sem banco, RabbitMQ ou Docker
test-unit:
	go test ./...

# Testes que usam PostgreSQL e RabbitMQ configurados no .env
test-integration:
	go test -tags=integration ./internal/controllers ./internal/service ./internal/infra/broker

# Teste ponta a ponta que sobe dependências com Testcontainers
test-e2e:
	go test -tags=e2e ./cmd/app -run TestSagaE2E -v

# Testes do RAG simples e da orquestração LangGraph
test-python:
	python -m unittest -v test_rag_tutor.py
	python -m unittest -v test_langgraph_rag.py

# O comando 'sqlc' gera o código do banco automaticamente
sqlc:
	sqlc generate

# O comando 'migrate-up' cria as tabelas no banco de dados
migrate-up:
	migrate -path migrations -database "$(DB_URL)" -verbose up

# O comando 'migrate-down' destroi as tabelas (rollback)
migrate-down:
	migrate -path migrations -database "$(DB_URL)" -verbose down
