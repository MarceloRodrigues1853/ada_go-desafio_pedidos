package broker

import (
	"context"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
)

type RabbitMQConsumer struct {
	conn *amqp.Connection
	ch   *amqp.Channel
}

// NewRabbitMQConsumer conecta ao broker e prepara o canal para consumo
func NewRabbitMQConsumer(url string) (*RabbitMQConsumer, error) {
	conn, err := amqp.Dial(url)
	if err != nil {
		return nil, fmt.Errorf("falha ao conectar ao RabbitMQ (consumer): %w", err)
	}

	ch, err := conn.Channel()
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("falha ao abrir canal AMQP (consumer): %w", err)
	}

	return &RabbitMQConsumer{conn: conn, ch: ch}, nil
}

// Consume consome mensagens da fila especificada (topic) executando o handler fornecido
func (c *RabbitMQConsumer) Consume(ctx context.Context, topic string, handler func([]byte) error) error {
	_, err := c.ch.QueueDeclare(
		topic, // nome da fila
		true,  // durable
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("falha ao declarar fila %s: %w", topic, err)
	}

	msgs, err := c.ch.Consume(
		topic, // queue
		"",    // consumer tag
		false, // auto-ack (false para processamento manual com Ack/Nack)
		false, // exclusive
		false, // no-local
		false, // no-wait
		nil,   // args
	)
	if err != nil {
		return fmt.Errorf("falha ao registrar consumidor para a fila %s: %w", topic, err)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case msg, ok := <-msgs:
			if !ok {
				return fmt.Errorf("canal de mensagens fechado para a fila %s", topic)
			}

			if err := handler(msg.Body); err != nil {
				slog.Error("Erro ao processar mensagem", "topic", topic, "erro", err.Error())
				_ = msg.Nack(false, false) // Rejeita sem requeue ou ajuste conforme política
			} else {
				_ = msg.Ack(false)
			}
		}
	}
}

func (c *RabbitMQConsumer) Close() error {
	if c.ch != nil {
		_ = c.ch.Close()
	}
	if c.conn != nil {
		return c.conn.Close()
	}
	return nil
}
