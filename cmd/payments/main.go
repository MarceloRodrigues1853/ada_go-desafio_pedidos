package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"pedidos/internal/events"
	"pedidos/internal/infra/broker"
	"pedidos/internal/payments"
)

func main() {
	// 1. Configuração de Logs estruturados (JSON)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
	slog.SetDefault(logger)

	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	logger.Info("Iniciando microsserviço de pagamentos...")

	// 2. Conexão Publisher (para enviar PaymentProcessedEvent)
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
	paymentHandler := payments.NewPaymentHandler(paymentService)

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

	// 5. Consome fila em background/loop
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
