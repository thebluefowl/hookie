package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"os"

	"github.com/thebluefowl/hookie/forwarder"
	"github.com/thebluefowl/hookie/listener"
	"github.com/thebluefowl/hookie/model"
	"github.com/thebluefowl/hookie/queue"
	"github.com/thebluefowl/hookie/server"
	"golang.org/x/exp/slog"
	"gopkg.in/yaml.v2"
)

func main() {

	ctx := context.Background()

	var rulesPath = flag.String("rules", "rules.yaml", "path to rules file")
	var configPath = flag.String("config", "config.yaml", "path to config file")
	flag.Parse()

	r, err := os.Open(*rulesPath)
	defer r.Close()
	handleErrorWithMessage(err, "failed to open rules file")

	c, err := os.Open(*configPath)
	defer c.Close()
	handleErrorWithMessage(err, "failed to open config file")

	rule, err := parseRules(r)
	handleErrorWithMessage(err, "failed to parse rules")

	config, err := parseConfig(c)
	handleErrorWithMessage(err, "failed to parse config")

	queue, err := getQueue(config)
	handleErrorWithMessage(err, "failed to initialize queue")

	instantForwarder := forwarder.NewInstantForwarder(http.DefaultTransport)
	queuedForwarder := forwarder.NewQueuedForwarder(queue)

	listener := listener.New(queue, http.DefaultTransport)
	go func() {
		err = listener.Listen(ctx)
	}()

	server := server.New(rule, instantForwarder, queuedForwarder)
	err = server.ListenAndServe(fmt.Sprintf("%s:%d", "", config.Port))
	handleErrorWithMessage(err, "failed to start server")

}

func handleErrorWithMessage(err error, message string) {
	if err != nil {
		slog.Error(message, slog.Any("err", err))
		os.Exit(1)
	}
}

func parseRules(r io.Reader) ([]model.Rule, error) {
	var rules []model.Rule
	err := yaml.NewDecoder(r).Decode(&rules)
	return rules, err
}

func parseConfig(r io.Reader) (*Config, error) {
	config := &Config{}
	err := yaml.NewDecoder(r).Decode(config)
	return config, err
}

func getQueue(cfg *Config) (model.PubSub, error) {
	if cfg.RabbitMQ != nil {
		opts := queue.RabbitMQOpts{
			Username:     cfg.RabbitMQ.Username,
			Password:     cfg.RabbitMQ.Password,
			Host:         cfg.RabbitMQ.Host,
			Port:         cfg.RabbitMQ.Port,
			ExchangeName: cfg.RabbitMQ.Exchange,
			RoutingKey:   cfg.RabbitMQ.RoutingKey,
			QueueName:    cfg.RabbitMQ.Queue,
		}
		return queue.NewRabbitMQ(&opts)
	}
	return nil, errors.New("queue not configured")
}
