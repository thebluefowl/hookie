package server

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestServer_handleInstantDelivery(t *testing.T) {
	server := createMockServer(t)
	defer server.Close()

	p := &Server{}

	u, _ := url.Parse(server.URL)
	u.Path = "/test-path"
	req, err := http.NewRequest(http.MethodGet, u.String(), strings.NewReader("test-body"))
	require.NoError(t, err)

	resp, err := p.handleInstantDelivery(context.TODO(), req)
	require.NoError(t, err)
	defer resp.Body.Close()

	assertExpectedResponse(t, resp, "Hello, world!")
}

func TestServer_handleQueuedDelivery(t *testing.T) {
	// Common setup
	req := httptest.NewRequest("GET", "http://example.com", nil)
	rawReqDump, _ := httputil.DumpRequest(req, true)
	ctx := context.TODO()
	// Scenario: Successful Enqueue
	t.Run("successful enqueue", func(t *testing.T) {
		publisher := new(MockPublisher)
		publisher.On("Publish", ctx, rawReqDump).Return(nil)

		p := &Server{publisher: publisher}

		resp, err := p.handleQueuedDelivery(ctx, req)

		assert.NoError(t, err)
		assert.Equal(t, http.StatusAccepted, resp.StatusCode)
		publisher.AssertExpectations(t)
	})

	// Scenario: Failed Enqueue
	t.Run("failed enqueue", func(t *testing.T) {
		ctx := context.TODO()
		publisher := new(MockPublisher)
		publisher.On("Publish", ctx, rawReqDump).Return(errors.New("enqueue failed"))

		p := &Server{publisher: publisher}

		resp, err := p.handleQueuedDelivery(ctx, req)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})

	// Scenario: Request Dump Error
	t.Run("request dump error", func(t *testing.T) {
		ctx := context.TODO()
		errorReq := httptest.NewRequest("GET", "http://example.com", &errorReader{})
		p := &Server{}
		resp, err := p.handleQueuedDelivery(ctx, errorReq)

		assert.Error(t, err)
		assert.Nil(t, resp)
	})
}

func createMockServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assertHTTPRequest(t, r)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("Hello, world!")) // nolint: errcheck
	}))
}

func assertHTTPRequest(t *testing.T, r *http.Request) {
	assert.Equal(t, http.MethodGet, r.Method)
	assert.Equal(t, "/test-path", r.URL.Path)
	body, err := io.ReadAll(r.Body)
	require.NoError(t, err)
	assert.Equal(t, "test-body", string(body))
}

func assertExpectedResponse(t *testing.T, resp *http.Response, expected string) {
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, expected, string(body))
}

// MockEnqueuer is a mock for the Enqueuer interface
type MockPublisher struct {
	mock.Mock
}

func (p *MockPublisher) Publish(ctx context.Context, payload []byte) error {
	args := p.Called(ctx, payload)
	return args.Error(0)
}

// This is used to test the httputil.DumpRequest error path
type errorReader struct{}

func (er *errorReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("simulated read error")
}
