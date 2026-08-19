package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pedidos/internal/events"
	"pedidos/internal/infra/broker"
	internalLogger "pedidos/internal/infra/logger"
	"pedidos/internal/infra/metrics"
	"pedidos/internal/payments"
	"pedidos/internal/repository"
	"pedidos/internal/repository/db"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// 1. Configuração de Logs estruturados (JSON)
	baseHandler := slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	})
	logger := slog.New(internalLogger.NewContextHandler(baseHandler))
	slog.SetDefault(logger)

	_ = godotenv.Load()

	// Registra os coletores de métricas Prometheus
	metrics.InitMetrics()

	dbUrl := os.Getenv("DB_URL")
	if dbUrl == "" {
		dbUrl = "postgres://root:password@localhost:5432/pedidos_db?sslmode=disable"
	}

	pool, err := pgxpool.New(context.Background(), dbUrl)
	if err != nil {
		logger.Error("Erro ao conectar ao banco de dados", "erro", err.Error())
		os.Exit(1)
	}
	defer pool.Close()

	queries := db.New(pool)
	paymentRepo := repository.NewPaymentPostgresRepository(queries, pool)

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	logger.Info("Iniciando microsserviço de pagamentos...")

	// 2. Conexão Publisher (para enviar PaymentProcessedEvent / PaymentFailedEvent)
	publisher, err := broker.NewRabbitMQPublisher(rabbitURL)
	if err != nil {
		logger.Error("Erro ao conectar publisher ao RabbitMQ", "erro", err.Error())
		os.Exit(1)
	}
	defer publisher.Close()

	// 3. Conexão Consumer (para consumir OrderCreatedEvent)
	consumer, err := broker.NewRabbitMQConsumer(rabbitURL)
	if err != nil {
		logger.Error("Erro ao conectar consumer ao RabbitMQ", "erro", err.Error())
		os.Exit(1)
	}
	defer consumer.Close()

	paymentService := payments.NewPaymentService(publisher)
	paymentHandler := payments.NewPaymentHandler(paymentService, paymentRepo)

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// 4. Tratamento de Graceful Shutdown (SIGINT / SIGTERM)
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		logger.Info("Sinal de encerramento recebido", "sinal", sig.String())
		cancel()
	}()

	logger.Info("Microsserviço de pagamentos escutando eventos", "topic", events.TopicOrderCreated)

	// 5. Expõe o endpoint /metrics (Prometheus) em porta dedicada
	metricsPort := os.Getenv("METRICS_PORT")
	if metricsPort == "" {
		metricsPort = "9091"
	}
	go func() {
		mux := http.NewServeMux()
		mux.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":"+metricsPort, mux); err != nil {
			logger.Error("Erro ao expor métricas Prometheus", "erro", err.Error())
		}
	}()
	logger.Info("Endpoint de métricas Prometheus ativo", "porta", metricsPort)

	// 6. Consome fila em background/loop
	errChan := make(chan error, 1)
	go func() {
		err := consumer.Consume(ctx, events.TopicOrderCreated, paymentHandler.HandleMessage)
		if err != nil && ctx.Err() == nil {
			errChan <- err
		}
	}()

	select {
	case <-ctx.Done():
		logger.Info("Encerrando microsserviço de pagamentos graciosamente...")
		time.Sleep(1 * time.Second)
	case err := <-errChan:
		logger.Error("Erro fatal no consumidor RabbitMQ", "erro", err.Error())
		os.Exit(1)
	}

	logger.Info("Microsserviço de pagamentos finalizado com sucesso.")
}
