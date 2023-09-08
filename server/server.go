package server

import (
	"context"
	"errors"
	"io"

	"net/http"

	"github.com/google/uuid"
	"github.com/thebluefowl/hookie/forwarder"
	"github.com/thebluefowl/hookie/model"
	"golang.org/x/exp/slog"
)

// Server represents the main HTTP server struct, holding necessary ruleset actions and the publisher.
type Server struct {
	rulesetActions []model.Rule
	forwarders     map[string]forwarder.Forwarder
}

// New creates a new instance of the Server.
func New(rulesetActions []model.Rule, instantForwarder *forwarder.InstantForwarder, queuedForwarder *forwarder.QueuedForwarder) *Server {
	fallbackForwarder := forwarder.NewFallbackForwarder(instantForwarder, queuedForwarder)
	return &Server{
		rulesetActions: rulesetActions,
		forwarders: map[string]forwarder.Forwarder{
			model.DeliveryModeInstant:  instantForwarder,
			model.DeliveryModeQueued:   queuedForwarder,
			model.DeliveryModeFallback: fallbackForwarder,
		},
	}

}

// ServeHTTP is the HTTP request handler for the server.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {

	requestID := uuid.New().String()
	ctx := context.WithValue(req.Context(), model.ContextKey("request-id"), requestID)

	slog.Info("INCOMING-REQUEST", slog.Any("request-id", requestID), slog.Any("method", req.Method), slog.Any("url", req.URL.String()))

	r, err := s.matchRule(req)
	if err != nil {
		slog.Error("failed to select ruleset-action", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	slog.Info("MATCHING-RULE", slog.String("request-id", requestID), slog.Any("rule", r.Name))

	res, err := s.process(ctx, req, r)
	if err != nil {
		slog.Error("failed to process request", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	w.WriteHeader(res.StatusCode)
	if res.Body != nil {
		_, err := io.Copy(w, res.Body)
		if err != nil {
			slog.Error("failed to copy response body", slog.Any("err", err))
		}
	}
}

// ListenAndServe starts the server on the given address.
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s)
}

// process handles the incoming request by matching it to a ruleset and then processing it based on the delivery mode.
func (s *Server) process(ctx context.Context, req *http.Request, ra *model.Rule) (*http.Response, error) {
	if ra != nil {
		fw, ok := s.forwarders[ra.Action.DeliveryMode]
		if !ok {
			return nil, model.ErrUnknownDeliveryMode
		}
		return fw.Forward(ctx, req, ra.Action.URL())
	}
	return nil, nil
}

// matchingRulesetAction finds the first matching ruleset action for a given request.
func (s *Server) matchRule(req *http.Request) (*model.Rule, error) {
	for _, ra := range s.rulesetActions {
		res, err := ra.TriggerSet.Match(req)
		if err != nil {
			return nil, err
		}
		if res {
			return &ra, nil
		}
	}
	return nil, errors.New("no matching ruleset action found")
}
