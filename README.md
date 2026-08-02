# 📦 Sistema de Processamento de Pedidos (Web API - Go)

Este repositório contém a resolução do desafio prático de arquitetura e domínio em Go, desenvolvido durante a formação **Ser+Tech (Ada Tech & Núclea Associação)**.

O projeto evoluiu de uma simulação no terminal para uma **API Web 100% funcional**, operando com banco de dados relacional (**PostgreSQL**), roteamento HTTP, geração automática de queries com **sqlc** e gerenciamento de pool de conexões com **pgx**.

## 🎯 Objetivo do Projeto
Construir o núcleo de um serviço de pedidos isolando as regras de negócio puras de frameworks externos. A aplicação simula um ambiente de **e-commerce**, protegendo a integridade do banco de dados e garantindo regras como: não vender sem estoque, validar a existência de clientes e impedir mudanças de status inválidas (ex: **cancelar um pedido já pago**).

## 🏗️ Arquitetura e Estrutura

O projeto segue os princípios da **Clean Architecture** e **Domain-Driven Design (DDD)**, garantindo desacoplamento entre a regra de negócio pura e a infraestrutura externa (banco de dados, servidor HTTP e frameworks). 

### 📐 Divisão das Camadas:

* **`cmd/app/`**: Ponto de entrada da aplicação (`main.go`). Lê variáveis de ambiente (`.env`), estabelece o pool de conexões com o PostgreSQL (`pgxpool`) e registra as rotas no roteador **go-chi**.
* **`internal/domain/`**: **Núcleo Puro (Core Domain)**. Contém as entidades, objetos de valor e invariantes da aplicação. Isento de qualquer dependência de banco de dados ou pacote externo.
* **`internal/service/`**: **Casos de Uso (Use Cases)**. Orquestra a execução das regras de negócio, gerencia transações atômicas no PostgreSQL via `pgxpool` (garantindo consistência no estoque) e expõe interfaces para a camada web.
* **`internal/controllers/`**: **Adaptadores de Entrada (Handlers HTTP)**. Converte payloads JSON em structs tipadas do Go, valida UUIDs de entrada e traduz os retornos em respostas HTTP REST (`200 OK`, `201 Created`, `400 Bad Request`, `404 Not Found`, `409 Conflict`).
* **`internal/repository/db/`**: **Adaptador de Persistência**. Código Go fortemente tipado gerado automaticamente pelo **sqlc** a partir das consultas SQL puras.
* **`migrations/` & `sqlc/`**: Infraestrutura de versionamento do schema do banco de dados e arquivos de queries SQL.

---

### 📂 Árvore de Arquivos do Projeto:

```text
pedidos/
├── cmd/
│   └── app/
│       └── main.go                         # Ponto de entrada, injeção de dependências e servidor HTTP
├── internal/
│   ├── controllers/                        # Handlers HTTP REST e suíte de testes unitários da API
│   │   ├── client_controller.go
│   │   ├── client_controller_test.go
│   │   ├── order_controller.go
│   │   ├── order_controller_test.go
│   │   ├── product_controller.go
│   │   └── product_controller_test.go
│   ├── domain/                             # Entidades puras, agregados e validações de invariantes
│   │   ├── client.go
│   │   ├── client_test.go
│   │   ├── errors.go
│   │   ├── product.go
│   │   ├── product_test.go
│   │   └── order/
│   │       ├── order.go
│   │       ├── order_item.go
│   │       ├── order_test.go
│   │       └── status.go
│   ├── repository/
│   │   └── db/                             # Persistência tipada gerada automaticamente via sqlc + pgx
│   │       ├── clientes.sql.go
│   │       ├── db.go
│   │       ├── models.go
│   │       ├── pedidos.sql.go
│   │       └── produtos.sql.go
│   └── service/                            
│       ├── order_service.go                    # Criação atômica, pagamento e cancelamento com estorno
│       └── order_service_integration_test.go   # Orquestração dos casos de uso e testes de integração
├── migrations/                             # Versionamento do schema PostgreSQL (.sql)
├── sqlc/
│   └── queries/                            # Consultas SQL mapeadas para geração do sqlc
├── .env                                    # Variáveis de ambiente
├── docker-compose.yml                      # Containerização do banco PostgreSQL
├── Makefile                                # Automatisador de rotinas (make run, make migrate-up)
├── sqlc.yaml                               # Configuração e overrides de tipos do sqlc (UUID, Decimal)
└── README.md
```

Padrões Utilizados: Clean Architecture, Domain-Driven Design (DDD), Dependency Injection (DI), Interface Segregation Principle (ISP), Transaction Script / Unit of Work.

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
### 1. Suba a infraestrutura (Banco de Dados):
Certifique-se de ter o Docker instalado e o arquivo .env configurado.
```bash
docker-compose up -d
```

### 2. Crie as tabelas (Migrations) e baixe as dependências:
```bash
go mod tidy
make migrate-up
```

### 3. Execute a API:
```bash
make run
```

### A API ficará rodando no endereço `http://localhost:8080`
---

## 🧪 Rotas Disponíveis (Postman)
### 1. Clientes
**POST** /clientes -> Cadastra um novo cliente no banco (Retorna UUID).

**GET** /clientes -> Lista os clientes cadastrados.

### 2. Produtos (Estoque)
**POST** /produtos -> Cadastra um novo produto.

**GET** /produtos -> Lista catálogo e saldo em estoque.

### 3. Pedidos (Vendas)
**POST** /pedidos -> Cria carrinho, valida cliente via UUID, associa produtos via Foreign Key e desconta estoque.

**PUT** /pedidos/{id}/pagar -> Aprova pagamento e altera status para PAID.

**PUT** /pedidos/{id}/cancelar -> Cancela venda (apenas se estiver pendente) e altera status para CANCELED.

---
## Cobertura de Testes Automatizados
O projeto possui suíte de testes unitários e de integração validando desde as regras de domínio até os fluxos de borda da API Web:

```bash
go test -count=1 -cover ./...
```
- **internal/controllers**: ~54.1% de cobertura de declarações.

- **internal/service**: ~49.2% de cobertura (**Testes de Integração em banco com pgxpool**).

- **internal/domain**: ~95.5% de cobertura nas entidades puras.

![testes_unitarios](./assets/testes_unitarios.png)

---

## 🧪 Exemplos de Uso e Testes no Terminal(Cenários)
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

---

## 🧪 Rotas e Testes (Postman)

Com o servidor rodando, você pode utilizar o Postman, Thunder Client ou `curl` para interagir com a API.

### 1. Produtos (Estoque)
* **POST** `/produtos` -> Cadastra produto no sistema.
* **GET** `/produtos` -> Lista catálogo e estoque.

### 2. Pedidos (Vendas)
* **POST** `/pedidos` -> Cria carrinho e desconta estoque automaticamente.
* **PUT** `/pedidos/{id}/pagar` -> Aprova pagamento.
* **PUT** `/pedidos/{id}/cancelar` -> Cancela venda e estorna estoque.

---

## Casos de Teste (Postman)

Abaixo estão os resultados das requisições reais feitas à API, comprovando o funcionamento do roteamento e a blindagem das regras de negócio pela Clean Architecture.

**1. Rota de Criação com Sucesso (201 Created)** O sistema aceita o payload e cadastra o produto ou pedido perfeitamente em memória.
![Criação Produto com Sucesso](./assets/criar_produto_postman.png)

![Criação Pedido com Sucesso](./assets/criar_pedido_postman.png)

**2. Bloqueio de Domínio - Validação de Dados (400 Bad Request)** Tentativa de cadastrar um produto com estoque negativo (-5). A entidade de Domínio intercepta e bloqueia a ação antes de chegar ao repositório.
![Bloqueio Produto Inválido](assets/criar_produto_invalido.png)

**3. Orquestração de Pagamento (200 OK)** O Service localiza o pedido, o Domínio aprova a mudança de status e o Repositório salva o novo estado.
![Pagamento Aprovado](assets/pagar_pedido_postman.png)

**4. Proteção contra Status Inválido (400 Bad Request)** Tentativa de cancelar um pedido que **já foi pago**. A regra de negócio proíbe a ação e devolve o erro customizado (`mudança de status inválida`).
![Bloqueio Cancelamento](assets/cancelar_pedido_postman.png)

---

## 🛠️ Tecnologias Utilizadas
- **Linguagem**: Go (Golang) 1.22+

- **Banco de Dados**: PostgreSQL rodando em Docker

- **Drivers e Conexão**: `pgxpool` (jackc/pgx/v5)

- **Roteamento HTTP**: `go-chi/chi`

- **Geração de DB Code**: `sqlc`

- **Migrations**: `golang-migrate`

- **Tipagem**: UUIDs oficiais do Google (`google/uuid`)

---
*Desenvolvido como parte do módulo de backend em Go da Ada Tech (Formação GO / Ser+Tech).*
