package forwarder

import (
	"context"
	"net/http"
	"net/url"
)

type Forwarder interface {
	Forward(ctx context.Context, req *http.Request, target *url.URL) (*http.Response, error)
}
