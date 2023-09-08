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

	rulesPath, configPath := parseFlags()

	config := loadConfig(configPath)
	rules := loadRules(rulesPath)

	queue := initializeQueue(config)

	initializeListener(ctx, queue)
	initializeAndRunServer(rules, config, queue)
}

func parseFlags() (string, string) {
	var rulesPath = flag.String("rules", "rules.yaml", "path to rules file")
	var configPath = flag.String("config", "config.yaml", "path to config file")
	flag.Parse()
	return *rulesPath, *configPath
}

func loadRules(rulesPath string) []model.Rule {
	r, err := os.Open(rulesPath)
	handleErrorWithMessage(err, "failed to open rules file")
	defer closeFile(r, "failed to close rules file")

	rules, err := parseRules(r)
	handleErrorWithMessage(err, "failed to parse rules")
	return rules
}

func loadConfig(configPath string) *Config {
	c, err := os.Open(configPath)
	handleErrorWithMessage(err, "failed to open config file")
	defer closeFile(c, "failed to close config file")

	config, err := parseConfig(c)
	handleErrorWithMessage(err, "failed to parse config")
	return config
}

func closeFile(f *os.File, message string) {
	if err := f.Close(); err != nil {
		slog.Error(message, slog.Any("err", err))
		os.Exit(1)
	}
}

func initializeQueue(config *Config) model.PubSub {
	queue, err := getQueue(config)
	handleErrorWithMessage(err, "failed to initialize queue")
	return queue
}

func initializeListener(ctx context.Context, queue model.PubSub) {
	listener := listener.New(queue, http.DefaultTransport)
	go func() {
		if err := listener.Listen(ctx); err != nil {
			slog.Error("listener error", slog.Any("err", err))
			os.Exit(1)
		}
	}()
}

func initializeAndRunServer(rules []model.Rule, config *Config, queue model.PubSub) {
	instantForwarder := forwarder.NewInstantForwarder(http.DefaultTransport)
	queuedForwarder := forwarder.NewQueuedForwarder(queue)

	server := server.New(rules, instantForwarder, queuedForwarder)
	if err := server.ListenAndServe(fmt.Sprintf(":%d", config.Port)); err != nil {
		handleErrorWithMessage(err, "failed to start server")
	}
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
