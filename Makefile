# Carrega as variáveis do .env
include .env

# O comando 'run' liga o servidor
run:
	go run cmd/app/main.go

# O comando 'sqlc' gera o código do banco automaticamente
sqlc:
	sqlc generate

# O comando 'migrate-up' cria as tabelas no banco de dados
migrate-up:
	migrate -path migrations -database "$(DB_URL)" -verbose up

# O comando 'migrate-down' destroi as tabelas (rollback)
migrate-down:
	migrate -path migrations -database "$(DB_URL)" -verbose down