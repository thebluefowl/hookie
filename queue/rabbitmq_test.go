package queue

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewRabbitMQ(t *testing.T) {
	opts := &RabbitMQOpts{
		Host:     "localhost",
		Port:     5672,
		Username: "hookie",
		Password: "hookie",
	}

	rmq, err := NewRabbitMQ(opts)
	assert.NoError(t, err)
	assert.NotNil(t, rmq)
}

func TestRabbitMQ_Publish(t *testing.T) {
	opts := &RabbitMQOpts{
		Host:     "localhost",
		Port:     5672,
		Username: "hookie",
		Password: "hookie",
	}

	rmq, err := NewRabbitMQ(opts)
	require.NoError(t, err)

	err = rmq.Publish(context.TODO(), []byte("payload"))
	assert.NoError(t, err)
}

func TestRabbitMQ_StartConsumer(t *testing.T) {
	opts := &RabbitMQOpts{
		Host:     "localhost",
		Port:     5672,
		Username: "hookie",
		Password: "hookie",
	}

	rmq, err := NewRabbitMQ(opts)
	require.NoError(t, err)

	messages := []string{
		"1",
		"2",
		"3",
		"4",
	}

	for _, msg := range messages {
		err = rmq.Publish(context.TODO(), []byte(msg))
		require.NoError(t, err)
	}

	ctx, cancel := context.WithCancel(context.TODO())
	defer cancel()

	go func() {
		time.Sleep(1 * time.Second)
		cancel()
	}()

	received := make(chan interface{}, len(messages))
	err = rmq.StartConsumer(ctx, func(body interface{}) error {
		fmt.Println(body)
		received <- body
		return nil
	})

	require.NoError(t, err)

	for i := 0; i < len(messages); i++ {
		select {
		case <-received:
		case <-time.After(2 * time.Second):
			t.Fatalf("did not receive all messages")
		}
	}
}
