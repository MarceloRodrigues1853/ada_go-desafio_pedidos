//go:build integration

package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"pedidos/internal/controllers"
	"pedidos/internal/events"
	"pedidos/internal/infra/broker"
	"pedidos/internal/payments"
	"pedidos/internal/repository"
	"pedidos/internal/repository/db"
	"pedidos/internal/service"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/modules/rabbitmq"
)

func TestSagaE2E_FluxoFeliz(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	postgresContainer, err := postgres.Run(ctx, "postgres:18-alpine",
		postgres.WithDatabase("pedidos_db"),
		postgres.WithUsername("root"),
		postgres.WithPassword("password"),
		postgres.BasicWaitStrategies(),
	)
	if err != nil {
		t.Fatalf("subir PostgreSQL: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(postgresContainer); err != nil {
			t.Errorf("encerrar PostgreSQL: %v", err)
		}
	})

	rabbitContainer, err := rabbitmq.Run(ctx, "rabbitmq:4-alpine")
	if err != nil {
		t.Fatalf("subir RabbitMQ: %v", err)
	}
	t.Cleanup(func() {
		if err := testcontainers.TerminateContainer(rabbitContainer); err != nil {
			t.Errorf("encerrar RabbitMQ: %v", err)
		}
	})

	dbURL, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		t.Fatalf("obter URL do PostgreSQL: %v", err)
	}
	pool, err := pgxpool.New(ctx, dbURL)
	if err != nil {
		t.Fatalf("criar pool PostgreSQL: %v", err)
	}
	t.Cleanup(pool.Close)
	applyMigrations(t, ctx, pool)

	rabbitURL, err := rabbitContainer.AmqpURL(ctx)
	if err != nil {
		t.Fatalf("obter URL do RabbitMQ: %v", err)
	}

	queries := db.New(pool)
	clientRepository := repository.NewClientPostgresRepository(queries)
	productRepository := repository.NewProductPostgresRepository(queries)
	orderRepository := repository.NewOrderPostgresRepository(queries, pool)

	orderPublisher, err := broker.NewRabbitMQPublisher(rabbitURL)
	if err != nil {
		t.Fatalf("criar publisher de pedidos: %v", err)
	}
	t.Cleanup(func() { _ = orderPublisher.Close() })

	orderService := service.NewOrderService(clientRepository, productRepository, orderRepository, orderPublisher)
	router := newRouter(
		controllers.NewClientController(service.NewClientService(clientRepository)),
		controllers.NewProductController(service.NewProductService(productRepository)),
		controllers.NewOrderController(orderService),
	)
	server := httptest.NewServer(router)
	t.Cleanup(server.Close)

	paymentPublisher, err := broker.NewRabbitMQPublisher(rabbitURL)
	if err != nil {
		t.Fatalf("criar publisher de pagamentos: %v", err)
	}
	t.Cleanup(func() { _ = paymentPublisher.Close() })

	paymentConsumer, err := broker.NewRabbitMQConsumer(rabbitURL)
	if err != nil {
		t.Fatalf("criar consumer de pagamentos: %v", err)
	}
	t.Cleanup(func() { _ = paymentConsumer.Close() })
	resultConsumer, err := broker.NewRabbitMQConsumer(rabbitURL)
	if err != nil {
		t.Fatalf("criar consumer de resultados: %v", err)
	}
	t.Cleanup(func() { _ = resultConsumer.Close() })

	consumerCtx, stopConsumers := context.WithCancel(context.Background())
	t.Cleanup(stopConsumers)
	paymentHandler := payments.NewPaymentHandler(
		payments.NewPaymentService(paymentPublisher),
		repository.NewPaymentPostgresRepository(queries, pool),
	)
	consumerErrors := make(chan error, 2)
	go consume(consumerCtx, paymentConsumer, events.TopicOrderCreated, paymentHandler.HandleMessage, consumerErrors)
	go consume(consumerCtx, resultConsumer, events.TopicPaymentProcessed, func(body []byte) error {
		var event events.PaymentProcessedEvent
		if err := json.Unmarshal(body, &event); err != nil {
			return err
		}
		return orderService.ProcessPaymentResult(consumerCtx, event)
	}, consumerErrors)

	clientID := createClientE2E(t, server.URL)
	createProductE2E(t, server.URL, "PROD-E2E", 50, 10)
	orderID := createOrderE2E(t, server.URL, clientID, "PROD-E2E", 2)

	deadline := time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		select {
		case err := <-consumerErrors:
			t.Fatalf("consumer encerrou durante a SAGA: %v", err)
		default:
		}

		order, orderErr := queries.GetPedidoByID(ctx, orderID)
		product, productErr := queries.GetProduto(ctx, "PROD-E2E")
		_, paymentErr := queries.GetProcessedEvent(ctx, orderID)
		if orderErr == nil && productErr == nil && paymentErr == nil && order.Status == "PAID" && product.Estoque == 8 {
			return
		}
		time.Sleep(100 * time.Millisecond)
	}

	order, orderErr := queries.GetPedidoByID(ctx, orderID)
	product, productErr := queries.GetProduto(ctx, "PROD-E2E")
	_, paymentErr := queries.GetProcessedEvent(ctx, orderID)
	t.Fatalf("SAGA não concluiu: pedido=%+v (erro=%v), produto=%+v (erro=%v), pagamento processado erro=%v",
		order, orderErr, product, productErr, paymentErr)
}

func consume(ctx context.Context, consumer *broker.RabbitMQConsumer, topic string, handler func([]byte) error, errors chan<- error) {
	if err := consumer.Consume(ctx, topic, handler); err != nil && ctx.Err() == nil {
		errors <- err
	}
}

func applyMigrations(t *testing.T, ctx context.Context, pool *pgxpool.Pool) {
	t.Helper()

	files, err := filepath.Glob(filepath.Join("..", "..", "migrations", "*.up.sql"))
	if err != nil {
		t.Fatalf("listar migrations: %v", err)
	}
	if len(files) == 0 {
		t.Fatal("nenhuma migration encontrada")
	}
	for _, file := range files {
		contents, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("ler migration %s: %v", file, err)
		}
		if _, err := pool.Exec(ctx, string(contents)); err != nil {
			t.Fatalf("aplicar migration %s: %v", file, err)
		}
	}
}

func createClientE2E(t *testing.T, baseURL string) uuid.UUID {
	t.Helper()
	var response struct {
		ID uuid.UUID `json:"id"`
	}
	postJSONE2E(t, baseURL+"/clientes", map[string]any{
		"name": "Cliente E2E", "email": "cliente-e2e@example.com", "password": "senha-segura",
	}, &response)
	return response.ID
}

func createProductE2E(t *testing.T, baseURL, id string, price float64, stock int) {
	t.Helper()
	postJSONE2E(t, baseURL+"/produtos", map[string]any{
		"id": id, "nome": "Produto E2E", "preco": price, "estoque": stock,
	}, nil)
}

func createOrderE2E(t *testing.T, baseURL string, clientID uuid.UUID, productID string, quantity int) uuid.UUID {
	t.Helper()
	var response struct {
		ID uuid.UUID `json:"id"`
	}
	postJSONE2E(t, baseURL+"/pedidos", map[string]any{
		"cliente_id": clientID,
		"itens":      []map[string]any{{"produto_id": productID, "quantidade": quantity}},
	}, &response)
	return response.ID
}

func postJSONE2E(t *testing.T, url string, payload any, output any) {
	t.Helper()
	body, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("serializar payload: %v", err)
	}
	response, err := http.Post(url, "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("POST %s: %v", url, err)
	}
	defer response.Body.Close()
	if response.StatusCode != http.StatusCreated {
		t.Fatalf("POST %s: esperado HTTP 201, obtido %s", url, response.Status)
	}
	if output != nil {
		if err := json.NewDecoder(response.Body).Decode(output); err != nil {
			t.Fatalf("decodificar resposta de %s: %v", url, err)
		}
	}
}
