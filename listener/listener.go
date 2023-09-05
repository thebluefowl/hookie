package listener

import (
	"context"

	"github.com/thebluefowl/hookie/model"
)

type Listener struct {
	consumer model.Consumer
}

func New(consumer model.Consumer) *Listener {
	return &Listener{
		consumer: consumer,
	}
}

func (l *Listener) Listen(ctx context.Context) error {
	return l.consumer.StartConsumer(ctx, func(body interface{}) error {

		return nil
	})
}
