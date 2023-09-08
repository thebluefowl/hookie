package forwarder

import (
	"context"
	"net/http"
	"net/url"

	"github.com/thebluefowl/hookie/model"
	"golang.org/x/exp/slog"
)

type FallbackForwarder struct {
	instantForwarder *InstantForwarder
	queuedForwarder  *QueuedForwarder
}

func NewFallbackForwarder(instantForwarder *InstantForwarder, queuedForwarder *QueuedForwarder) Forwarder {
	return &FallbackForwarder{
		instantForwarder: instantForwarder,
		queuedForwarder:  queuedForwarder,
	}
}

func (fw *FallbackForwarder) Forward(ctx context.Context, req *http.Request, target *url.URL) (*http.Response, error) {
	requestID := ctx.Value(model.ContextKey("request-id")).(string)
	res, err := fw.instantForwarder.Forward(ctx, req, target)
	if err != nil || res.StatusCode < http.StatusInternalServerError {
		slog.Info("FALLBACK-TO-QUEUED", slog.String("request-id", requestID))
		return fw.queuedForwarder.Forward(ctx, req, target)
	} else {
		return res, nil
	}
}
