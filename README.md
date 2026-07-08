# 📦 Sistema de Processamento de Pedidos (Web API - Go)

Este repositório contém a resolução do desafio prático de arquitetura e domínio em Go, desenvolvido durante a formação **Ser+Tech (Ada Tech & Núclea Associação)**.

O projeto evoluiu de uma simulação no terminal para uma **API Web 100% funcional**, focada em regras de negócio, Design de Domínio (DDD) e Clean Architecture, operando com dados em memória e roteamento HTTP.

## 🎯 Objetivo do Projeto
Construir o núcleo de um serviço de pedidos isolando as regras de negócio puras de frameworks externos. 
A aplicação simula um ambiente de **e-commerce**, protegendo o estado do sistema contra **compras sem estoque** ou **mudanças de status inválidas** (ex: **cancelar um pedido já pago**)

## 🏗️ Arquitetura e Estrutura
O projeto foi estruturado seguindo os padrões do ecossistema Go, separando claramente as responsabilidades:

* `cmd/app/` : Ponto de entrada da aplicação e orquestrador do servidor HTTP (**go-chi**).
* `internal/controllers/` : Camada Web (**Handlers**). Traduz as requisições JSON da internet para a linguagem do sistema.
* `internal/domain/` : O coração do sistema. Contém as entidades (`Produto`, `Pedido`, `Status`), os erros de domínio (**Sentinel Errors**) e as regras inegociáveis.
* `internal/repository/` : Contratos (**Interfaces**) e a implementação do armazenamento em memória utilizando Maps.
* `internal/service/` : Orquestrador dos casos de uso (**Criar Pedido, Pagar, Cancelar**), coordenando a comunicação entre o **Domínio** e a **camada de Dados**.
  
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
### Instale as dependências e execute a aplicação:
```bash
go mod tidy
go run cmd/app/main.go
```

### A API ficará rodando no endereço `http://localhost:8080`
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

## 📸 Evidências de Teste (Postman)

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

## 🛠️ Tecnologias
- **Linguagem**: Go (Golang)

- **Padrões**: Clean Architecture, Dependency Injection, Interface Segregation.

---
*Desenvolvido como parte do módulo de backend em Go da Ada Tech.*