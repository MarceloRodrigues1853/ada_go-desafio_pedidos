# =============================================================================
# Build multi-estágio: compila os dois binários (app e payments) em um único stage
# e depois copia cada um para uma imagem mínima alpine.
# =============================================================================

# Estágio 1: compilação dos binários Go
FROM golang:1.26-alpine AS builder
WORKDIR /build

# Baixa as dependências primeiro para aproveitar o cache de camadas do Docker
COPY go.mod go.sum ./
RUN go mod download

# Compila o código-fonte em binários estáticos (CGO desabilitado para rodar em alpine)
COPY . .
RUN CGO_ENABLED=0 go build -o /out/app ./cmd/app \
 && CGO_ENABLED=0 go build -o /out/payments ./cmd/payments

# Estágio 2: imagem final do serviço de pedidos (API HTTP)
FROM alpine:3.21 AS app
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/app /usr/local/bin/app
EXPOSE 8080
ENTRYPOINT ["app"]

# Estágio 3: imagem final do microsserviço de pagamentos (consumer RabbitMQ)
FROM alpine:3.21 AS payments
RUN apk add --no-cache ca-certificates
COPY --from=builder /out/payments /usr/local/bin/payments
EXPOSE 9091
ENTRYPOINT ["payments"]