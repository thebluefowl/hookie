package forwarder

import (
	"net/http"
	"net/url"

	"github.com/thebluefowl/hookie/proxyutils"
)

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

func (fw *InstantForwarder) Forward(req *http.Request, target *url.URL) (*http.Response, error) {
	out, err := proxyutils.NewTargetRequest(req, target)
	if err != nil {
		return nil, err
	}

	res, err := fw.roundTripper.RoundTrip(out.Request)
	if err != nil {
		return nil, err
	}

	tr := proxyutils.NewTargetResponse(res)
	return tr.Response, nil
}
