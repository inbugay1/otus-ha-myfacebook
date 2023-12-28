package rmq

import (
	"context"
	"fmt"
	"net"

	amqp "github.com/rabbitmq/amqp091-go"
)

type Config struct {
	Host     string
	Port     string
	Username string
	Password string
}

type RMQ struct {
	config  *Config
	conn    *amqp.Connection
	channel *amqp.Channel
}

func New(config *Config) *RMQ {
	return &RMQ{
		config: config,
	}
}

func (rmq *RMQ) Connect() error {
	conn, err := amqp.Dial(fmt.Sprintf("amqp://%s:%s@%s/", rmq.config.Username, rmq.config.Password, net.JoinHostPort(rmq.config.Host, rmq.config.Port)))
	if err != nil {
		return fmt.Errorf("failed to connect to rmq on %s:%s: %w", rmq.config.Host, rmq.config.Port, err)
	}

	channel, err := conn.Channel()
	if err != nil {
		return fmt.Errorf("failed to open rmq channel: %w", err)
	}

	rmq.conn = conn
	rmq.channel = channel

	return nil
}

func (rmq *RMQ) Publish(ctx context.Context, exchangeName string, message []byte) error {
	err := rmq.channel.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil)
	if err != nil {
		return fmt.Errorf("failed do declare rmq exchange %q: %w", exchangeName, err)
	}

	err = rmq.channel.PublishWithContext(ctx, exchangeName, "", false, false, amqp.Publishing{
		ContentType: "text/plain",
		Body:        message,
	})
	if err != nil {
		return fmt.Errorf("failed to publish rmq message to exchange %q: %w", exchangeName, err)
	}

	return nil
}

func (rmq *RMQ) Consume(_ context.Context, exchangeName, queueName string) (<-chan amqp.Delivery, error) {
	err := rmq.channel.ExchangeDeclare(exchangeName, "fanout", true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed do declare rmq exchange %q: %w", exchangeName, err)
	}

	queue, err := rmq.channel.QueueDeclare(queueName, true, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to declare rmq queue: %w", err)
	}

	err = rmq.channel.QueueBind(queue.Name, "", exchangeName, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to bind queue to exchange: %w", err)
	}

	msgs, err := rmq.channel.Consume(queue.Name, "", false, false, false, false, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to consume messages from queue %q: %w", queue.Name, err)
	}

	return msgs, nil
}

func (rmq *RMQ) Disconnect() error {
	if err := rmq.channel.Close(); err != nil {
		return fmt.Errorf("failed to close rmq channel: %w", err)
	}

	if err := rmq.conn.Close(); err != nil {
		return fmt.Errorf("failed to close rmq connection: %w", err)
	}

	return nil
}
