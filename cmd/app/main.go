package main

import (
	// Pacotes nativos do Go
	"context"  // Gerencia o ciclo de vida e timeouts de conexões
	"log"      // Exibe logs formatados no terminal
	"net/http" // Servidor e rotas HTTP
	"os"       // Interage com o sistema operacional e lê variáveis do .env

	// Camadas internas da aplicação
	"pedidos/internal/controllers"   // Handlers Web
	"pedidos/internal/repository/db" // Acesso ao banco via sqlc
	"pedidos/internal/service"       // Regras de negócio da aplicação

	// Bibliotecas de terceiros
	"github.com/go-chi/chi/v5"            // Roteador HTTP
	"github.com/go-chi/chi/v5/middleware" // Middlewares de log e recuperação de pânicos
	"github.com/jackc/pgx/v5/pgxpool"     // Gerenciador de pool de conexões do Postgres
	"github.com/joho/godotenv"            // Carregador do arquivo .env
)

func main() {
	// =========================================================================
	// 1. CONFIGURAÇÃO DO AMBIENTE (.env)
	// =========================================================================
	if err := godotenv.Load(); err != nil {
		log.Println("Aviso: Arquivo .env não encontrado. O sistema tentará usar variáveis nativas.")
	}

	// =========================================================================
	// 2. CONEXÃO COM O BANCO DE DADOS (POSTGRESQL VIA PGXPOOL)
	// =========================================================================
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		log.Fatal("Erro Crítico: A variável de ambiente DB_URL não foi definida no arquivo .env")
	}

	// Cria o pool de conexões para reutilização eficiente
	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		log.Fatalf("Erro Crítico ao tentar conectar ao banco de dados: %v", err)
	}
	defer pool.Close() // Fecha o pool com segurança ao encerrar o programa

	log.Println("📦 Conexão com o PostgreSQL estabelecida com sucesso!")

	// =========================================================================
	// 3. INJEÇÃO DE DEPENDÊNCIAS
	// =========================================================================
	queries := db.New(pool)

	// Instancia a camada de serviço enviando o banco de dados
	orderService := service.NewOrderService(queries, pool)

	// Instancia os controladores HTTP
	clientController := controllers.NewClientController(queries)
	productController := controllers.NewProductController(queries)
	orderController := controllers.NewOrderController(orderService)

	// =========================================================================
	// 4. CONFIGURAÇÃO DE ROTAS E MIDDLEWARES (CHI ROUTER)
	// =========================================================================
	r := chi.NewRouter()

	r.Use(middleware.Logger)    // Loga cada requisição HTTP no console
	r.Use(middleware.Recoverer) // Evita crash da aplicação em caso de erro grave

	// --- ROTAS DA ENTIDADE: CLIENTES ---
	r.Post("/clientes", clientController.Create)      // POST /clientes
	r.Get("/clientes", clientController.List)         // GET /clientes
	r.Get("/clientes/{id}", clientController.GetByID) // GET /clientes/{id} (EXIGIDO)

	// --- ROTAS DA ENTIDADE: PRODUTOS ---
	r.Post("/produtos", productController.Create)      // POST /produtos
	r.Get("/produtos", productController.List)         // GET /produtos
	r.Get("/produtos/{id}", productController.GetByID) // GET /produtos/{id} (EXIGIDO)

	// --- ROTAS DA ENTIDADE: PEDIDOS ---
	r.Post("/pedidos", orderController.Create)               // POST /pedidos
	r.Get("/pedidos", orderController.ListPaginado)          // GET /pedidos?limit=10&offset=0 (EXIGIDO)
	r.Get("/pedidos/{id}", orderController.GetByID)          // GET /pedidos/{id} (EXIGIDO)
	r.Post("/pedidos/{id}/pagar", orderController.Pay)       // POST /pedidos/{id}/pagar (Ajustado verbo POST)
	r.Post("/pedidos/{id}/cancelar", orderController.Cancel) // POST /pedidos/{id}/cancelar (Ajustado verbo POST)

	// =========================================================================
	// 5. INICIALIZAÇÃO DO SERVIDOR HTTP
	// =========================================================================
	porta := os.Getenv("PORT")
	if porta == "" {
		porta = "8080"
	}

	log.Printf("🚀 Servidor HTTP inicializado com sucesso na porta %s\n", porta)

	if err := http.ListenAndServe(":"+porta, r); err != nil {
		log.Fatalf("Erro Fatal: Não foi possível iniciar o servidor HTTP: %v", err)
	}
}
