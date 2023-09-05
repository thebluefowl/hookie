package server

import (
	"context"
	"io"

	"net/http"
	"net/http/httputil"

	"github.com/thebluefowl/hookie/model"
	"golang.org/x/exp/slog"
)

// Server represents the main HTTP server struct, holding necessary ruleset actions and the publisher.
type Server struct {
	rulesetActions []model.RulesetAction
	publisher      model.Publisher
	ctx            context.Context
}

// New creates a new instance of the Server.
func New(ctx context.Context, rulesetActions []model.RulesetAction, publisher model.Publisher) *Server {
	return &Server{
		rulesetActions: rulesetActions,
		publisher:      publisher,
		ctx:            ctx,
	}
}

// ServeHTTP is the HTTP request handler for the server.
func (s *Server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := s.ctx
	res, err := s.process(ctx, r)
	if err != nil {
		slog.Error("failed to process request", slog.Any("err", err))
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}
	if res == nil {
		http.NotFound(w, r)
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

// process handles the incoming request by matching it to a ruleset and then processing it based on the delivery mode.
func (s *Server) process(ctx context.Context, req *http.Request) (*http.Response, error) {
	ra, err := s.matchingRulesetAction(req)
	if err != nil {
		return nil, err
	}

	if ra != nil {
		out := ra.Action.OutboundRequest(req)
		switch ra.Action.DeliveryMode {
		case model.DeliveryModeInstant:
			response, err := s.handleInstantDelivery(ctx, out)
			if ra.Action.DeliveryMode == model.DeliveryModeFallback && (err != nil || response.StatusCode >= 500) {
				return s.handleQueuedDelivery(ctx, out)
			}
			return response, err
		case model.DeliveryModeQueued:
			return s.handleQueuedDelivery(ctx, out)
		default:
			return nil, nil
		}
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
	return nil, nil
}

// handleInstantDelivery processes the request for instant delivery mode.
func (s *Server) handleInstantDelivery(ctx context.Context, req *http.Request) (*http.Response, error) {
	response, err := http.DefaultTransport.RoundTrip(req)
	if err != nil {
		slog.Error("failed to make outbound request", slog.Any("err", err))
		return nil, err
	}
	return response, nil
}

// handleQueuedDelivery processes the request for queued delivery mode, storing the request for later processing.
func (s *Server) handleQueuedDelivery(ctx context.Context, req *http.Request) (*http.Response, error) {
	raw, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}

	err = s.publisher.Publish(ctx, raw)
	if err != nil {
		return nil, err
	}

	return &http.Response{
		StatusCode: http.StatusAccepted,
	}, nil
}

// ListenAndServe starts the server on the given address.
func (s *Server) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s)
}
