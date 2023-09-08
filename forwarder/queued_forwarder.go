package forwarder

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/thebluefowl/hookie/model"
	"github.com/thebluefowl/hookie/proxyutils"
	"golang.org/x/exp/slog"
)

type Publisher interface {
	Publish(ctx context.Context, payload []byte) error
}

type QueuedForwarder struct {
	publisher Publisher
}

func NewQueuedForwarder(publisher Publisher) *QueuedForwarder {
	return &QueuedForwarder{
		publisher: publisher,
	}
}

func (fw *QueuedForwarder) Forward(ctx context.Context, req *http.Request, target *url.URL) (*http.Response, error) {
	requestID := ctx.Value(model.ContextKey("request-id")).(string)
	out, err := proxyutils.NewTargetRequest(requestID, req, target)
	if err != nil {
		return nil, fmt.Errorf("failed to create target request: %w", err)
	}

	payload, err := out.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal target request: %w", err)
	}

	slog.Info("PUBLISH-ATTEMPT", slog.String("request-id", requestID))
	if err := fw.publisher.Publish(ctx, payload); err != nil {
		slog.Error("PUBLISH-FAIL", slog.String("request-id", requestID), slog.Any("err", err))
		return nil, fmt.Errorf("failed to publish target request: %w", err)
	}
	slog.Info("PUBLISH-SUCCESS", slog.String("request-id", requestID))

	return &http.Response{
		StatusCode: http.StatusAccepted,
	}, nil
}
