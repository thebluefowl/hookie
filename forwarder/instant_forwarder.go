package forwarder

import (
	"context"
	"net/http"
	"net/url"
	"time"

	"github.com/thebluefowl/hookie/model"
	"github.com/thebluefowl/hookie/proxyutils"
	"golang.org/x/exp/slog"
)

var now = func() int64 {
	return time.Now().UnixMilli()
}

type InstantForwarder struct {
	roundTripper http.RoundTripper
}

func NewInstantForwarder(roundTripper http.RoundTripper) *InstantForwarder {
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}

	return &InstantForwarder{
		roundTripper: roundTripper,
	}
}

func (fw *InstantForwarder) Forward(ctx context.Context, req *http.Request, target *url.URL) (*http.Response, error) {
	requestID := ctx.Value(model.ContextKey("request-id")).(string)
	targetRequest, err := proxyutils.NewTargetRequest(requestID, req, target)
	if err != nil {
		return nil, err
	}

	slog.Info("REQUEST-SENDING", slog.String("request-id", requestID))
	t0 := now()
	res, err := fw.roundTripper.RoundTrip(targetRequest.Request)
	t1 := now()
	slog.Info("RESPONSE-RECEIVED", slog.String("request-id", requestID), slog.Int("status-code", res.StatusCode), slog.Int64("duration-ms", t1-t0))
	if err != nil {
		return nil, err
	}

	targetResponse := proxyutils.NewTargetResponse(res)
	return targetResponse.Response, nil
}
