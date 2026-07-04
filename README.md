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
* **`internal/service/`**: Orquestrador dos casos de uso (Criar Pedido, Pagar, Cancelar), coordenando a comunicação entre o Domínio e a camada de Dados.
  
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
git clone https://github.com/MarceloRodrigues1853/ada_go-desafio_pedidos.git
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

## 🧪 Exemplos de Uso e Testes (Cenários)
A aplicação inclui um script de simulação no `main.go` que atua como um **teste de integração** das regras de negócio. Abaixo estão as evidências de execução cobrindo os cenários de **sucesso** e **bloqueio**:

### Bloqueio de Produto Inválido
O sistema impede a inicialização se tentar registrar um produto com dados ausentes (como nome vazio) ou valores negativos.

![produto_invalido](./assets/produto_invalido.png)

### Validação de Cliente e Pedido
O sistema barra a criação de carrinhos de compra caso a identificação do pedido ou do cliente estejam em branco.

![cliente_invalido](./assets/cliente_invalido.png)

## Proteção contra Estoque Insuficiente
Se a quantidade solicitada for maior que o saldo do repositório, a entidade do domínio entra em ação e estorna a operação.

![estoque_invalido](./assets/estoque_invalido.png)

### Caminho Feliz (Compra e Pagamento)
O fluxo ideal, onde o sistema aprova a compra, reduz o estoque no repositório em memória e atualiza o status via Service.

![sucesso](./assets/sucesso.png)

## 🛠️ Tecnologias
- **Linguagem**: Go (Golang)

- **Padrões**: Clean Architecture, Dependency Injection, Interface Segregation.

