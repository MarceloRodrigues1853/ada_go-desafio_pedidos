package main

import (
	// Pacotes nativos do Go
	"context" // Gerencia o ciclo de vida e timeouts de conexões
	"encoding/json"
	// Exibe logs formatados no terminal
	"log/slog" // Logs estruturados
	"net/http" // Servidor e rotas HTTP
	"os"       // Interage com o sistema operacional e lê variáveis do .env

	// Camadas internas da aplicação
	"pedidos/internal/controllers" // Handlers Web
	"pedidos/internal/events"
	"pedidos/internal/infra/broker"
	"pedidos/internal/repository"
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
	// 1. CONFIGURAÇÃO DE LOGS (SLOG JSON) E AMBIENTE (.env)
	// =========================================================================
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	if err := godotenv.Load(); err != nil {
		logger.Warn("Arquivo .env não encontrado. O sistema tentará usar variáveis nativas.")
	}

	// =========================================================================
	// 2. CONEXÃO COM O BANCO DE DADOS (POSTGRESQL VIA PGXPOOL)
	// =========================================================================
	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		logger.Error("A variável de ambiente DB_URL não foi definida no arquivo .env")
		os.Exit(1)
	}

	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		logger.Error("Erro ao tentar conectar ao banco de dados", "erro", err.Error())
		os.Exit(1)
	}
	defer pool.Close()

	logger.Info("Conexão com o PostgreSQL estabelecida com sucesso")

	// =========================================================================
	// 3. INJEÇÃO DE DEPENDÊNCIAS (SERVICE + DECORATOR + CONTROLLERS)
	// =========================================================================
	queries := db.New(pool)

	// 3.1 Instancia o serviço base
	clientRepository := repository.NewClientPostgresRepository(queries)
	productRepository := repository.NewProductPostgresRepository(queries)
	orderRepository := repository.NewOrderPostgresRepository(queries, pool)
	clientService := service.NewClientService(clientRepository)
	productService := service.NewProductService(productRepository)

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}
	eventPublisher, err := broker.NewRabbitMQPublisher(rabbitURL)
	if err != nil {
		logger.Warn("Aviso: Falha ao conectar ao RabbitMQ. Mensageria desativada.", "erro", err.Error())
	}

	baseOrderService := service.NewOrderService(clientRepository, productRepository, orderRepository, eventPublisher)

	// 3.2 Decora o serviço base com o LoggingOrderService
	orderService := service.NewLoggingOrderService(baseOrderService, logger)

	// 3.4 Inicializa consumidor RabbitMQ para fechar a SAGA (payment.processed)
	paymentConsumer, err := broker.NewRabbitMQConsumer(rabbitURL)
	if err != nil {
		logger.Warn("Aviso: Falha ao conectar consumidor RabbitMQ para payment.processed.", "erro", err.Error())
	} else {
		go func() {
			ctxBg := context.Background()
			err := paymentConsumer.Consume(ctxBg, events.TopicPaymentProcessed, func(body []byte) error {
				var paymentEvent events.PaymentProcessedEvent
				if err := json.Unmarshal(body, &paymentEvent); err != nil {
					logger.Error("Erro ao desserializar PaymentProcessedEvent", "erro", err.Error())
					return err
				}
				return orderService.ProcessPaymentResult(ctxBg, paymentEvent)
			})
			if err != nil {
				logger.Error("Erro no consumidor RabbitMQ payment.processed", "erro", err.Error())
			}
		}()
	}

	// 3.3 Instancia os controladores HTTP passando o serviço decorado
	clientController := controllers.NewClientController(clientService)
	productController := controllers.NewProductController(productService)
	orderController := controllers.NewOrderController(orderService)

	// =========================================================================
	// 4. CONFIGURAÇÃO DE ROTAS E MIDDLEWARES (CHI ROUTER)
	// =========================================================================
	r := chi.NewRouter()

	r.Use(middleware.Logger)    // Loga cada requisição HTTP no console
	r.Use(middleware.Recoverer) // Evita crash da aplicação em caso de erro grave

	// --- ROTAS DA ENTIDADE: CLIENTES ---
	r.Post("/clientes", clientController.Create)
	r.Get("/clientes", clientController.List)
	r.Get("/clientes/{id}", clientController.GetByID)

	// --- ROTAS DA ENTIDADE: PRODUTOS ---
	r.Post("/produtos", productController.Create)
	r.Get("/produtos", productController.List)
	r.Get("/produtos/{id}", productController.GetByID)

	// --- ROTAS DA ENTIDADE: PEDIDOS ---
	r.Post("/pedidos", orderController.Create)
	r.Get("/pedidos", orderController.ListPaginado)
	r.Get("/pedidos/{id}", orderController.GetByID)
	r.Post("/pedidos/{id}/pagar", orderController.Pay)
	r.Post("/pedidos/{id}/cancelar", orderController.Cancel)

	// =========================================================================
	// 5. INICIALIZAÇÃO DO SERVIDOR HTTP
	// =========================================================================
	porta := os.Getenv("PORT")
	if porta == "" {
		porta = "8080"
	}

	logger.Info("Servidor HTTP inicializado com sucesso", "porta", porta)

	if err := http.ListenAndServe(":"+porta, r); err != nil {
		logger.Error("Não foi possível iniciar o servidor HTTP", "erro", err.Error())
		os.Exit(1)
	}
}
