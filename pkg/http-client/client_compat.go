package httpclient

import (
	"net/http"
	"strconv"
	"strings"
	"time"
)

// RetryConfig holds retry configuration for NewRetryClient.
type RetryConfig struct {
	MaxRetries           int
	InitialDelay         time.Duration
	MaxDelay             time.Duration
	RetryableStatusCodes map[int]bool
	BackoffMultiplier    float64
}

// NewHTTPClientWithOptions creates a new HTTP client with options.
// Uses the existing Option type from retryable.go.
func NewHTTPClientWithOptions(opts ...Option) HTTPClient {
	return NewHTTPClientRetryable(opts...)
}

// NewRetryClient creates an HTTP client with retry capabilities.
// The o11y parameter is accepted for API compatibility but not currently used.
func NewRetryClient(base HTTPClient, config RetryConfig, o11y interface{}) HTTPClient {
	return &retryableHTTPClient{
		base:   base,
		config: config,
	}
}

// ParseRetryableStatusCodes parses a comma-separated string of status codes.
// Example: "408,429,500,502,503,504"
func ParseRetryableStatusCodes(codes string) map[int]bool {
	if codes == "" {
		return nil
	}

	result := make(map[int]bool)
	parts := strings.Split(codes, ",")

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if code, err := strconv.Atoi(part); err == nil {
			result[code] = true
		}
	}

	return result
}

// retryableHTTPClient wraps an HTTPClient with retry logic.
type retryableHTTPClient struct {
	base   HTTPClient
	config RetryConfig
}

// Do implements HTTPClient interface with retry logic.
func (c *retryableHTTPClient) Do(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	delay := c.config.InitialDelay

	for attempt := 0; attempt <= c.config.MaxRetries; attempt++ {
		resp, err = c.base.Do(req)

		if err == nil && !c.shouldRetry(resp.StatusCode) {
			return resp, nil
		}

		if attempt < c.config.MaxRetries {
			if resp != nil && resp.Body != nil {
				_ = resp.Body.Close()
			}

			time.Sleep(delay)
			delay = time.Duration(float64(delay) * c.config.BackoffMultiplier)
			if delay > c.config.MaxDelay {
				delay = c.config.MaxDelay
			}
		}
	}

	return resp, err
}

func (c *retryableHTTPClient) shouldRetry(statusCode int) bool {
	if c.config.RetryableStatusCodes == nil {
		return statusCode >= 500
	}
	return c.config.RetryableStatusCodes[statusCode]
}
