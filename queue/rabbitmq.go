package queue

import (
	"context"
	"fmt"
	"log"

	"github.com/wagslane/go-rabbitmq"
)

const RMQDefaultExchangeName = "hookie.exchange.default"
const RMQDefaultRoutingKey = "hookie.webhook.default"
const RMQDefaultQueueName = "hookie.webhook.default"

type RabbitMQ struct {
	conn         *rabbitmq.Conn
	publisher    *rabbitmq.Publisher
	ExchangeName string
	RoutingKey   string
	QueueName    string
}

type RabbitMQOpts struct {
	Username     string
	Password     string
	Host         string
	Port         int
	ExchangeName string
	RoutingKey   string
	QueueName    string
}

func NewRabbitMQ(opts *RabbitMQOpts) (*RabbitMQ, error) {
	if opts.ExchangeName == "" {
		opts.ExchangeName = RMQDefaultExchangeName
	}
	if opts.RoutingKey == "" {
		opts.RoutingKey = RMQDefaultRoutingKey
	}
	if opts.QueueName == "" {
		opts.QueueName = RMQDefaultQueueName
	}
	conn, err := rabbitmq.NewConn(
		fmt.Sprintf("amqp://%s:%s@%s:%d/", opts.Username, opts.Password, opts.Host, opts.Port),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to RabbitMQ: %w", err)
	}

	publisher, err := rabbitmq.NewPublisher(
		conn,
		rabbitmq.WithPublisherOptionsLogging,
		rabbitmq.WithPublisherOptionsExchangeName(opts.ExchangeName),
		rabbitmq.WithPublisherOptionsExchangeDeclare,
		rabbitmq.WithPublisherOptionsExchangeDurable,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to create RabbitMQ publisher: %w", err)
	}

	return &RabbitMQ{
		conn:         conn,
		publisher:    publisher,
		ExchangeName: opts.ExchangeName,
		RoutingKey:   opts.RoutingKey,
		QueueName:    opts.QueueName,
	}, nil
}

func (r *RabbitMQ) Publish(ctx context.Context, body []byte) error {
	err := r.publisher.Publish(
		body,
		[]string{r.RoutingKey},
		rabbitmq.WithPublishOptionsContentType("application/octet-stream"),
		rabbitmq.WithPublishOptionsExchange(r.ExchangeName),
	)
	fmt.Println("errror", err)
	return err
}

func (r *RabbitMQ) StartConsumer(ctx context.Context, processor func(payload interface{}) error) error {
	consumer, err := rabbitmq.NewConsumer(
		r.conn,
		func(d rabbitmq.Delivery) rabbitmq.Action {
			if err := processor(d.Body); err != nil {
				log.Printf("failed to process payload: %v", err)
				return rabbitmq.NackRequeue
			}

			return rabbitmq.Ack
		},
		r.QueueName,
		rabbitmq.WithConsumerOptionsRoutingKey(r.RoutingKey),
		rabbitmq.WithConsumerOptionsExchangeName(r.ExchangeName),
	)

	if err != nil {
		log.Fatalf("failed to create RabbitMQ consumer: %v", err)
		return err
	}

	defer consumer.Close()

	// wait till context is done
	<-ctx.Done()

	return nil
}
