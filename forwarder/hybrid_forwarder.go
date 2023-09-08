package forwarder

import (
	"context"
	"net/http"
	"net/url"
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
	res, err := fw.instantForwarder.Forward(ctx, req, target)
	if err != nil || res.StatusCode > http.StatusInternalServerError {
		return fw.queuedForwarder.Forward(ctx, req, target)
	} else {
		return res, nil
	}
}
