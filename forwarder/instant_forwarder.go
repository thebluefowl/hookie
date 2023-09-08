package forwarder

import (
	"context"
	"net/http"
	"net/http/httputil"
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

	x, _ := httputil.DumpRequest(targetRequest.Request, true)
	println(string(x))

	res, err := fw.roundTripper.RoundTrip(targetRequest.Request)
	if err != nil {
		return nil, err
	}

	targetResponse := proxyutils.NewTargetResponse(res)
	return targetResponse.Response, nil
}
