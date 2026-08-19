package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	DLXExchange = "dlx_exchange"
	PaymentsDLQ = "payments_dlq"
	OrdersDLQ   = "orders_dlq"
)

type RabbitMQPublisher struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewRabbitMQPublisher conecta ao broker e prepara o canal
func NewRabbitMQPublisher(url string) (*RabbitMQPublisher, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao RabbitMQ: %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("falha ao abrir canal AMQP: %w", err)
	}

	return &RabbitMQPublisher{conn: conn, ch: ch}, nil
}

func declareQueueWithDLQ(ch *amqp.Channel, queueName string) error {
	// 1. Declarar a Dead Letter Exchange (DLX)
	err := ch.ExchangeDeclare(
		DLXExchange, // name
		"direct",    // type
		true,        // durable
		false,       // auto-deleted
		false,       // internal
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("falha ao declarar DLX: %w", err)
	}

	// 2. Determinar a DLQ correspondente
	dlqName := PaymentsDLQ
	if queueName == "payment.processed" || queueName == "payment.failed" {
		dlqName = OrdersDLQ
	}

	// 3. Declarar a fila DLQ
	_, err = ch.QueueDeclare(
		dlqName, // name
		true,    // durable
		false,   // auto-delete
		false,   // exclusive
		false,   // no-wait
		nil,     // arguments
	)
	if err != nil {
		return fmt.Errorf("falha ao declarar DLQ %s: %w", dlqName, err)
	}

	// 4. Fazer o QueueBind da DLQ na DLX
	err = ch.QueueBind(
		dlqName,     // queue name
		dlqName,     // routing key
		DLXExchange, // exchange
		false,       // no-wait
		nil,         // arguments
	)
	if err != nil {
		return fmt.Errorf("falha ao fazer bind da DLQ: %w", err)
	}

	// 5. Declarar a fila principal com argumentos de DLX
	args := amqp.Table{
		"x-dead-letter-exchange":    DLXExchange,
		"x-dead-letter-routing-key": dlqName,
	}

	_, err = ch.QueueDeclare(
		queueName, // name
		true,      // durable
		false,     // auto-delete
		false,     // exclusive
		false,     // no-wait
		args,      // arguments
	)
	if err != nil {
		return fmt.Errorf("falha ao declarar fila principal %s com DLQ: %w", queueName, err)
	}

	return nil
}

// Publish serializa o evento em JSON e publica na fila informada
func (r *RabbitMQPublisher) Publish(ctx context.Context, topic string, event any) error {
	// Garante que a fila e DLQ existem
	if err := declareQueueWithDLQ(r.ch, topic); err != nil {
		return err
	}

	body, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("falha ao serializar evento: %w", err)
	}

	err = r.ch.PublishWithContext(
		ctx,
		"",    // exchange padrão (direct)
		topic, // routing key = nome da fila
		false, // mandatory
		false, // immediate
		amqp.Publishing{
			ContentType:  "application/json",
			DeliveryMode: amqp.Persistent, // Mensagem persistente em disco
			Body:         body,
		},
	)
	if err != nil {
		return fmt.Errorf("falha ao publicar mensagem na fila %s: %w", topic, err)
	}

	slog.Info("Evento publicado com sucesso no RabbitMQ", "topic", topic)
	return nil
}

func (r *RabbitMQPublisher) Close() error {
	if r.ch != nil {
		_ = r.ch.Close()
	}
	if r.conn != nil {
		return r.conn.Close()
	}
	return nil
}
