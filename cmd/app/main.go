package main

import (
	// Pacotes nativos do Go
	"context"  // Gerencia o ciclo de vida e timeouts de conexões e requisições
	"log"      // Utilizado para formatar e exibir mensagens no terminal (logs)
	"net/http" // Provê as ferramentas para criar o servidor HTTP e lidar com requisições Web
	"os"       // Permite interagir com o Sistema Operacional (ex: ler variáveis do .env)

	// Camadas internas da aplicação (Módulo: pedidos)
	"pedidos/internal/controllers"   // Controladores que recebem o JSON do Postman e tratam HTTP
	"pedidos/internal/repository/db" // Código gerado pelo sqlc que executa comandos no Postgres
	"pedidos/internal/service"       // Camada de Serviço que isola as Regras de Negócio da aplicação

	// Bibliotecas externas de terceiros (baixadas via go get)
	"github.com/go-chi/chi/v5"            // Roteador HTTP moderno, leve e extremamente rápido
	"github.com/go-chi/chi/v5/middleware" // Componentes prontos para logar requisições e tratar pânicos
	"github.com/jackc/pgx/v5/pgxpool"     // Driver oficial e gerenciador de "piscina" de conexões Postgres
	"github.com/joho/godotenv"            // Biblioteca que lê o arquivo .env e injeta no sistema
)

func main() {
	// =========================================================================
	// 1. CONFIGURAÇÃO DO AMBIENTE (.env)
	// =========================================================================
	// O godotenv procura o arquivo ".env" na raiz para isolar dados sensíveis (senhas) do código
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: Arquivo .env não encontrado. O sistema tentará usar variáveis nativas.")
	}

	// =========================================================================
	// 2. CONEXÃO COM O BANCO DE DADOS (POSTGRESQL VIA PGXPOOL)
	// =========================================================================
	// Captura a URL de conexão configurada no arquivo .env
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("Erro Crítico: A variável de ambiente DB_URL não foi definida no arquivo .env")
	}

	// O pgxpool.New cria uma "piscina" (pool) de conexões reutilizáveis com o banco.
	// Isso evita abrir e fechar uma conexão a cada clique do usuário, otimizando o servidor.
	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatalf("Erro Crítico ao tentar conectar ao banco de dados: %v", err)
	}
	// O "defer" garante que a piscina de conexões será fechada com segurança quando a API desligar
	defer pool.Close()

	log.Println("📦 Conexão com o PostgreSQL estabelecida com sucesso!")

	// =========================================================================
	// 3. INJEÇÃO DE DEPENDÊNCIAS - CAMADA DE REPOSITÓRIO (BANCO DE DADOS)
	// =========================================================================
	// Instancia o repositório automático de Clientes gerado pelo sqlc, passando o pool do Postgres
	queries := db.New(pool)

	// =========================================================================
	// 4. INJEÇÃO DE DEPENDÊNCIAS - CAMADA DE SERVIÇO (REGRAS DE NEGÓCIO)
	// =========================================================================
	// O Service centraliza as regras cruciais (ex: "pedido precisa ter itens", "estoque deve diminuir").
	// Ele recebe os repositórios necessários para consultar e persistir dados.
	orderService := service.NewOrderService(queries)

	// =========================================================================
	// 5. INJEÇÃO DE DEPENDÊNCIAS - CAMADA DE CONTROLE (HTTP / API)
	// =========================================================================
	// Os Controladores servem como a ponte entre o Postman e o núcleo da aplicação.
	// Injetamos a variável "queries" (o banco real) em vez da memória RAM provisória
	clientController := controllers.NewClientController(queries)
	productController := controllers.NewProductController(queries)
	orderController := controllers.NewOrderController(orderService)

	// =========================================================================
	// 6. CONFIGURAÇÃO DE ROTAS E MIDDLEWARES (CHI ROUTER)
	// =========================================================================
	r := chi.NewRouter()

	// Middlewares interceptam a requisição para executar tarefas repetitivas automaticamente:
	r.Use(middleware.Logger)    // Imprime no terminal detalhes de cada requisição (ex: GET /clientes 200)
	r.Use(middleware.Recoverer) // Se o código sofrer um crash inesperado (panic), ele segura e evita a API de cair

	// --- ROTAS DA ENTIDADE: CLIENTES (INTEGRADO AO POSTGRESQL REAL) ---
	r.Post("/clientes", clientController.Create) // Recebe dados e cadastra um novo cliente no banco
	r.Get("/clientes", clientController.List)    // Busca e lista todos os clientes cadastrados do banco

	// --- ROTAS DA ENTIDADE: PRODUTOS (INTEGRADO AO POSTGRESQL REAL) ---
	r.Post("/produtos", productController.Create)
	r.Get("/produtos", productController.List)

	// --- ROTAS DA ENTIDADE: PEDIDOS (INTEGRADO AO POSTGRESQL REAL) ---
	r.Post("/pedidos", orderController.Create)              // Cria uma intenção de compra
	r.Put("/pedidos/{id}/pagar", orderController.Pay)       // Altera o status do pedido para PAID (Pago)
	r.Put("/pedidos/{id}/cancelar", orderController.Cancel) // Cancela e devolve os produtos ao estoque

	// =========================================================================
	// 7. INICIALIZAÇÃO DO SERVIDOR HTTP
	// =========================================================================
	// Captura a porta definida no arquivo .env (padrão: 8080)
	porta := os.Getenv("PORT")
	if porta == "" {
		porta = "8080"
	}

	log.Printf("🚀 Servidor HTTP inicializado com sucesso na porta %s\n", porta)

	// Liga o servidor de fato e fica escutando as requisições que chegam da rede
	if err := http.ListenAndServe(":"+porta, r); err != nil {
		log.Fatalf("Erro Fatal: Não foi possível iniciar o servidor HTTP: %v", err)
	}
}
