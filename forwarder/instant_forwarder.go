package forwarder

import (
	"context"
	"net/http"
	"net/url"

	"github.com/thebluefowl/hookie/proxyutils"
)

type InstantForwarder struct {
	roundTripper http.RoundTripper
}

func NewInstantForwarder(roundTripper http.RoundTripper) Forwarder {
	if roundTripper == nil {
		roundTripper = http.DefaultTransport
	}

	return &InstantForwarder{
		roundTripper: roundTripper,
	}
}

func (fw *InstantForwarder) Forward(ctx context.Context, req *http.Request, target *url.URL) (*http.Response, error) {
	targetRequest, err := proxyutils.NewTargetRequest(req, target)
	if err != nil {
		return nil, err
	}
	res, err := fw.roundTripper.RoundTrip(targetRequest.Request)
	if err != nil {
		return nil, err
	}

	targetResponse := proxyutils.NewTargetResponse(res)
	return targetResponse.Response, nil
}
