package model

import (
	"context"
	"net/http"
	"net/url"
)

type QueuedRequest struct {
	Headers http.Header
	Body    []byte
	Method  string
	URL     url.URL
}

type Publisher interface {
	Publish(ctx context.Context, payload []byte) error
}

type Consumer interface {
	StartConsumer(context.Context, func(body interface{}) error) error
}

type PubSub interface {
	Publisher
	Consumer
}
