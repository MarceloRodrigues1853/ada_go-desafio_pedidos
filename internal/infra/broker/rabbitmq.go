package broker

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"

	amqp "github.com/rabbitmq/amqp091-go"
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

// Publish serializa o evento em JSON e publica na fila informada
func (r *RabbitMQPublisher) Publish(ctx context.Context, topic string, event any) error {
	// Garante que a fila existe
	_, err := r.ch.QueueDeclare(
		topic, // nome da fila
		true,  // durable (persiste se o RabbitMQ reiniciar)
		false, // auto-delete
		false, // exclusive
		false, // no-wait
		nil,   // arguments
	)
	if err != nil {
		return fmt.Errorf("falha ao declarar fila %s: %w", topic, err)
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
