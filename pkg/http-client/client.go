package httpclient

import (
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

type ClientOption func(*clientConfig)

type clientConfig struct {
	timeout   time.Duration
	transport http.RoundTripper
}

// NewHTTPClient creates a new HTTP client with default settings (backward compatible).
func NewHTTPClient() HTTPClient {
	return NewHTTPClientWithOptions()
}

// NewHTTPClientWithOptions creates a new HTTP client with custom options.
func NewHTTPClientWithOptions(opts ...ClientOption) HTTPClient {
	cfg := &clientConfig{
		timeout:   30 * time.Second, // default timeout
		transport: otelhttp.NewTransport(http.DefaultTransport),
	}

	for _, opt := range opts {
		opt(cfg)
	}

	return &http.Client{
		Timeout:   cfg.timeout,
		Transport: cfg.transport,
	}
}

// WithTimeout sets the timeout for the HTTP client.
func WithTimeout(timeout time.Duration) ClientOption {
	return func(c *clientConfig) {
		c.timeout = timeout
	}
}

// WithTransport sets the transport for the HTTP client.
func WithTransport(transport http.RoundTripper) ClientOption {
	return func(c *clientConfig) {
		c.transport = transport
	}
}
