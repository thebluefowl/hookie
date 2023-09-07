package listener

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/thebluefowl/hookie/model"
	"github.com/thebluefowl/hookie/proxyutils"
)

type Listener struct {
	consumer  model.Consumer
	transport http.RoundTripper
}

func New(consumer model.Consumer, transport http.RoundTripper) *Listener {
	if transport == nil {
		transport = http.DefaultTransport
	}
	return &Listener{
		consumer:  consumer,
		transport: transport,
	}
}

func (l *Listener) Listen(ctx context.Context) error {
	return l.consumer.StartConsumer(ctx, func(body interface{}) error {
		b, ok := body.([]byte)
		if !ok {
			return errors.New("payload should be []byte")
		}
		tr := &proxyutils.TargetRequest{}
		if err := json.Unmarshal(b, tr); err != nil {
			return fmt.Errorf("failed to unmarshal target request: %w", err)
		}
		resp, err := l.transport.RoundTrip(tr.Request)
		if err != nil {
			return fmt.Errorf("failed to forward request: %w", err)
		}
		defer resp.Body.Close()
		fmt.Println(resp.StatusCode)
		return nil
	})
}
