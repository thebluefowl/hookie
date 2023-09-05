package model

import (
	"fmt"
	"net/http"
	"net/url"
)

type DeliveryMode string

const (
	DeliveryModeInstant  = "instant"
	DeliveryModeFallback = "fallback"
	DeliveryModeQueued   = "queued"
)

type Action struct {
	UpstreamHost string       `yaml:"upstream"`
	upstreamHost url.URL      `yaml:"-"`
	DeliveryMode DeliveryMode `yaml:"delivery_mode"`
	TimeOut      int          `yaml:"timeout"`
	Delay        int          `yaml:"delay"`
	Retries      int          `yaml:"retries"`
}

func (a *Action) UnmarshalYAML(unmarshal func(interface{}) error) error {
	type plain Action
	if err := unmarshal((*plain)(a)); err != nil {
		return err
	}
	u, err := url.Parse(a.UpstreamHost)
	if err != nil {
		return fmt.Errorf("failed to parse upstream URL: %w", err)
	}
	a.upstreamHost = *u
	return nil
}

func (a *Action) OutboundRequest(in *http.Request) (out *http.Request) {
	base, err := url.Parse(a.upstreamHost.String())
	if err != nil {
		panic(err)
	}

	// check if in url is absolute
	if !in.URL.IsAbs() {
		in.URL = base.ResolveReference(in.URL)
	}

	return &http.Request{
		Method:     in.Method,
		URL:        in.URL,
		Proto:      in.Proto,
		ProtoMajor: in.ProtoMajor,
		ProtoMinor: in.ProtoMinor,
		Header:     cloneHeader(in.Header),
		Body:       in.Body,
		Close:      in.Close,
		Host:       a.upstreamHost.Host,
	}
}

func cloneHeader(h http.Header) http.Header {
	h2 := make(http.Header, len(h))
	for k, vv := range h {
		vv2 := make([]string, len(vv))
		copy(vv2, vv)
		h2[k] = vv2
	}
	return h2
}
