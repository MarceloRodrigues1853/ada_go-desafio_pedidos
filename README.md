# 📦 Sistema de Processamento de Pedidos (Go)

Este repositório contém a resolução do desafio prático de arquitetura e domínio em Go, desenvolvido durante a formação **Ser+ Tech (Ada Tech & Núclea)**.

## 🎯 Objetivo do Projeto
Construir o núcleo de um serviço de pedidos focando puramente em **regras de negócio, Design de Domínio (DDD) e Clean Architecture**, utilizando **exclusivamente a biblioteca padrão do Go** (Standard Library). 

O projeto simula um ambiente de e-commerce e processamento de pedidos no terminal, operando 100% em memória, sem o uso de frameworks externos, APIs HTTP ou banco de dados.

## 🏗️ Arquitetura e Estrutura
O projeto foi estruturado seguindo os padrões do ecossistema Go, separando claramente as responsabilidades:

* **`cmd/app/`**: Ponto de entrada da aplicação (onde a mágica acontece no terminal). Sem regras de negócio.
* **`internal/domain/`**: O coração do sistema. Contém as entidades (`Product`, `Order`, `Status`), os erros de domínio (Sentinel Errors) e as regras inegociáveis do negócio.
* **`internal/repository/`**: Contratos (Interfaces) e a implementação do armazenamento em memória utilizando `Maps`.
* **Linguagem**: Go (Golang)

```text
ada_go-desafio_pedidos/
├── cmd/
│   └── app/
│       └── main.go                  # Ponto de entrada (orquestração principal)
├── internal/
│   ├── domain/
│   │   ├── errors.go                # Erros globais do sistema (Sentinel Errors)
│   │   ├── order.go                 # Entidades e regras de Pedido
│   │   └── product.go               # Entidades e regras de Produto
│   ├── repository/
│   │   ├── memory_order_repo.go     # Implementação em memória (map) para pedidos
│   │   ├── memory_product_repo.go   # Implementação em memória (map) para produtos
│   │   ├── order_repository.go      # Interface (contrato) do repositório de pedidos
│   │   └── product_repository.go    # Interface (contrato) do repositório de produtos
│   └── service/
│       └── order_service.go         # Casos de uso e orquestração do negócio
└── go.mod                           # Gerenciador de dependências do Go
```

Padrões: Clean Architecture, Dependency Injection, Interface Segregation.
Orquestrador dos casos de uso (Criar Pedido, Pagar, Cancelar), coordenando a comunicação entre o Domínio e a camada de Dados.

## 🚀 Como Executar

Clone este repositório em sua máquina:
```bash
git clone [https://github.com/MarceloRodrigues1853/ada_go-desafio_pedidos.git](https://github.com/MarceloRodrigues1853/ada_go-desafio_pedidos.git)
```
---
### Navegue até a pasta do projeto:
```bash
cd ada_go-desafio_pedidos
```
### Execute a aplicação:
```bash
go run cmd/app/main.go
```
## 🛠️ Tecnologias
- **Linguagem**: Go (Golang)

- **Padrões**: Clean Architecture, Dependency Injection, Interface Segregation.

