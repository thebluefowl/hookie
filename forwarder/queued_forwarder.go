package forwarder

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/thebluefowl/hookie/proxyutils"
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

func (fw *QueuedForwarder) Forward(req *http.Request, target *url.URL) (*http.Response, error) {
	ctx := req.Context()
	out, err := proxyutils.NewTargetRequest(req, target)
	if err != nil {
		return nil, fmt.Errorf("failed to create target request: %w", err)
	}

	payload, err := out.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal target request: %w", err)
	}

	if err := fw.publisher.Publish(ctx, payload); err != nil {
		return nil, fmt.Errorf("failed to publish target request: %w", err)
	}

	return &http.Response{
		StatusCode: http.StatusAccepted,
	}, nil
}
