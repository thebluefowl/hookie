package server

import (
	"errors"
	"io"

	"net/http"
	"net/url"

	"github.com/thebluefowl/hookie/model"
	"golang.org/x/exp/slog"
)

type Forwarder interface {
	Forward(req *http.Request, target *url.URL) (*http.Response, error)
}

// Server represents the main HTTP server struct, holding necessary ruleset actions and the publisher.
type Server struct {
	rulesetActions []model.RulesetAction
	forwarders     map[string]Forwarder
}

// New creates a new instance of the Server.
func New(rulesetActions []model.RulesetAction, instantForwarder, queuedForwarder Forwarder) *Server {
	return &Server{
		rulesetActions: rulesetActions,
		forwarders: map[string]Forwarder{
			model.DeliveryModeInstant: instantForwarder,
			model.DeliveryModeQueued:  queuedForwarder,
		},
	}

}

// ServeHTTP is the HTTP request handler for the server.
func (s *Server) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	ra, err := s.matchingRulesetAction(req)
	if err != nil {
		slog.Error("failed to select ruleset-action", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	res, err := s.process(req, ra)
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
func (s *Server) process(req *http.Request, ra *model.RulesetAction) (*http.Response, error) {
	if ra != nil {
		fw, ok := s.forwarders[ra.Action.DeliveryMode]
		if !ok {
			return nil, model.ErrUnknownDeliveryMode
		}
		return fw.Forward(req, ra.Action.URL())
	}
	return nil, nil
}

// matchingRulesetAction finds the first matching ruleset action for a given request.
func (s *Server) matchingRulesetAction(req *http.Request) (*model.RulesetAction, error) {
	for _, ra := range s.rulesetActions {
		res, err := ra.Ruleset.Match(req)
		if err != nil {
			return nil, err
		}
		if res {
			return &ra, nil
		}
	}
	return nil, errors.New("no matching ruleset action found")
}
