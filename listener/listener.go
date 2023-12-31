package listener

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/thebluefowl/hookie/model"
	"github.com/thebluefowl/hookie/proxyutils"
	"github.com/thebluefowl/hookie/queue"
	"golang.org/x/exp/slog"
)

var now = func() int64 {
	return time.Now().UnixMilli()
}

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
			return queue.NewError(errors.New("payload should be []byte"), true)
		}
		tr := &proxyutils.TargetRequest{}
		if err := json.Unmarshal(b, tr); err != nil {
			return queue.NewError(fmt.Errorf("failed to unmarshal payload: %w", err), true)
		}

		slog.Info("LISTENER-REQUEST-SENDING", slog.String("request-id", tr.ID))
		t0 := now()
		resp, err := l.transport.RoundTrip(tr.Request)
		t1 := now()
		slog.Info("LISTENER-RESPONSE-RECEIVED", slog.String("request-id", tr.ID), slog.Int("status-code", resp.StatusCode), slog.Int64("duration-ms", t1-t0))
		if err != nil {
			return queue.NewError(fmt.Errorf("failed to forward request: %w", err), false)
		}
		defer resp.Body.Close()

		if resp.StatusCode > http.StatusInternalServerError {
			return queue.NewError(fmt.Errorf("failed to forward request: %s", resp.Status), false)

		}
		return nil
	})
}
