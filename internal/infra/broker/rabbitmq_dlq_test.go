package broker_test

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"

	"pedidos/internal/events"
	"pedidos/internal/infra/broker"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
	amqp "github.com/rabbitmq/amqp091-go"
)

func TestRabbitMQ_DLQ(t *testing.T) {
	_ = godotenv.Load("../../.env")
	rabbitURL := os.Getenv("RABBITMQ_URL")
	if rabbitURL == "" {
		rabbitURL = "amqp://guest:guest@localhost:5672/"
	}

	pub, err := broker.NewRabbitMQPublisher(rabbitURL)
	if err != nil {
		t.Skip("RabbitMQ não disponível. Pulando teste de DLQ.")
	}
	defer pub.Close()

	con, err := broker.NewRabbitMQConsumer(rabbitURL)
	if err != nil {
		t.Skip("RabbitMQ não disponível. Pulando teste de DLQ.")
	}
	defer con.Close()

	testQueue := "test_dlq_queue_" + uuid.New().String()[:8]

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Publica uma mensagem que acionará erro no handler
	event := events.OrderCreatedEvent{
		SagaID:      uuid.New(),
		OrderID:     uuid.New(),
		ClientID:    uuid.New(),
		TotalAmount: 100.0,
		Status:      "PENDING",
		CreatedAt:   time.Now(),
	}

	err = pub.Publish(ctx, testQueue, event)
	if err != nil {
		t.Fatalf("Falha ao publicar para teste de DLQ: %v", err)
	}

	// Consome a mensagem com handler retornando erro para forçar Nack(false, false)
	errChan := make(chan error, 1)
	go func() {
		errChan <- con.Consume(ctx, testQueue, func(body []byte) error {
			return errors.New("erro simulado para envio à DLQ")
		})
	}()

	// Aguarda um instante para a mensagem ser consumida e Nackada
	time.Sleep(500 * time.Millisecond)
	cancel()

	// Verifica se a mensagem foi entregue na payments_dlq
	conn, err := amqp.Dial(rabbitURL)
	if err != nil {
		t.Skip("RabbitMQ não disponível para verificar DLQ.")
		return
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("Erro ao abrir canal para verificar DLQ: %v", err)
	}
	defer ch.Close()

	msg, ok, err := ch.Get("payments_dlq", true)
	if err != nil {
		t.Fatalf("Erro ao ler da payments_dlq: %v", err)
	}

	if !ok {
		t.Errorf("Esperava encontrar mensagem na payments_dlq, mas a fila está vazia")
	} else {
		if len(msg.Body) == 0 {
			t.Errorf("Mensagem na DLQ está vazia")
		}
	}
}
