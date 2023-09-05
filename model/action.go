package model

import "net/url"

const (
	DeliveryModeInstant  = "instant"
	DeliveryModeFallback = "fallback"
	DeliveryModeQueued   = "queued"
)

type Action struct {
	UpstreamHost string `yaml:"upstream"`
	DeliveryMode string `yaml:"delivery_mode"`
	TimeOut      int    `yaml:"timeout"`
	Delay        int    `yaml:"delay"`
	Retries      int    `yaml:"retries"`
}

func (a *Action) URL() *url.URL {
	u, _ := url.Parse(a.UpstreamHost)
	return u
}
