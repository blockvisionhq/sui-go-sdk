package sui

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"
)

const (
	testRequestInterval = time.Second
	testMaxRetries      = 3
)

type throttledTransport struct {
	base       http.RoundTripper
	interval   time.Duration
	retryBase  time.Duration
	maxRetries int

	mu            sync.Mutex
	nextRequestAt time.Time
}

func newThrottledTestHTTPClient() *http.Client {
	return &http.Client{
		Timeout: 30 * time.Second,
		Transport: &throttledTransport{
			base:       http.DefaultTransport,
			interval:   testRequestInterval,
			retryBase:  time.Second,
			maxRetries: testMaxRetries,
		},
	}
}

func (t *throttledTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	for attempt := 0; ; attempt++ {
		if err := t.wait(req.Context()); err != nil {
			return nil, err
		}

		retry, err := cloneRequest(req)
		if err != nil {
			return nil, err
		}
		response, err := t.roundTrip(retry)
		if err != nil {
			return nil, err
		}

		body, err := io.ReadAll(response.Body)
		closeErr := response.Body.Close()
		if err != nil {
			return nil, err
		}
		if closeErr != nil {
			return nil, closeErr
		}
		response.Body = io.NopCloser(bytes.NewReader(body))

		if !isRateLimited(response, body) || attempt >= t.maxRetries {
			return response, nil
		}
		if err := waitFor(req.Context(), t.retryBase<<attempt); err != nil {
			return nil, err
		}
	}
}

func (t *throttledTransport) roundTrip(req *http.Request) (*http.Response, error) {
	if t.base != nil {
		return t.base.RoundTrip(req)
	}
	return http.DefaultTransport.RoundTrip(req)
}

func (t *throttledTransport) wait(ctx context.Context) error {
	t.mu.Lock()
	now := time.Now()
	requestAt := t.nextRequestAt
	if requestAt.Before(now) {
		requestAt = now
	}
	t.nextRequestAt = requestAt.Add(t.interval)
	t.mu.Unlock()

	return waitFor(ctx, time.Until(requestAt))
}

func cloneRequest(req *http.Request) (*http.Request, error) {
	clone := req.Clone(req.Context())
	if req.GetBody == nil {
		return clone, nil
	}
	body, err := req.GetBody()
	if err != nil {
		return nil, err
	}
	clone.Body = body
	return clone, nil
}

func isRateLimited(response *http.Response, body []byte) bool {
	return response.StatusCode == http.StatusTooManyRequests || strings.Contains(strings.ToLower(string(body)), "too frequent")
}

func waitFor(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}
	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func TestThrottledTransportRetriesRateLimitedResponse(t *testing.T) {
	var requests int
	server := httptest.NewServer(http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		requests++
		if requests == 1 {
			_, _ = writer.Write([]byte(`{"error_msg":"Your request is too frequent. Try again later."}`))
			return
		}
		_, _ = writer.Write([]byte(`{"jsonrpc":"2.0","result":"ok"}`))
	}))
	defer server.Close()

	transport := &throttledTransport{
		base:       http.DefaultTransport,
		interval:   time.Millisecond,
		retryBase:  time.Millisecond,
		maxRetries: 1,
	}
	request, err := http.NewRequestWithContext(context.Background(), http.MethodPost, server.URL, bytes.NewBufferString(`{"jsonrpc":"2.0"}`))
	if err != nil {
		t.Fatalf("create request: %v", err)
	}

	response, err := transport.RoundTrip(request)
	if err != nil {
		t.Fatalf("round trip: %v", err)
	}
	defer func() { _ = response.Body.Close() }()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read response: %v", err)
	}
	if got, want := requests, 2; got != want {
		t.Fatalf("request count = %d, want %d", got, want)
	}
	if strings.Contains(string(body), "too frequent") {
		t.Fatalf("returned rate-limited response: %s", body)
	}
}
