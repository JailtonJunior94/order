package httpclient

import (
	"fmt"
	"math/rand/v2"
	"net/http"
	"sync"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jailtonjunior94/order/pkg/observability"
)

type RetryConfig struct {
	MaxRetries           int
	InitialDelay         time.Duration
	MaxDelay             time.Duration
	RetryableStatusCodes map[int]bool
	BackoffMultiplier    float64
}

type retryClient struct {
	client              HTTPClient
	config              RetryConfig
	o11y                observability.Observability
	retryableStatusCode sync.Map // thread-safe map para evitar race conditions
}

// NewRetryClient wraps an HTTPClient with retry logic and observability.
func NewRetryClient(client HTTPClient, config RetryConfig, o11y observability.Observability) HTTPClient {
	if config.BackoffMultiplier == 0 {
		config.BackoffMultiplier = 2.0
	}

	rc := &retryClient{
		client: client,
		config: config,
		o11y:    o11y,
	}

	// Populate sync.Map with retryable status codes (thread-safe, no race conditions)
	if config.RetryableStatusCodes != nil {
		for code, retryable := range config.RetryableStatusCodes {
			if retryable {
				rc.retryableStatusCode.Store(code, true)
			}
		}
	}

	return rc
}

func (r *retryClient) Do(req *http.Request) (*http.Response, error) {
	ctx := req.Context()
	ctx, span := r.o11y.Tracer().Start(ctx, "http_request_with_retry",
		observability.WithAttributes(
			observability.String("http.method", req.Method),
			observability.String("http.url", req.URL.String()),
			observability.Int("http.max_retries", r.config.MaxRetries),
		),
	)
	defer span.End()

	// Create exponential backoff
	expBackoff := backoff.NewExponentialBackOff()
	expBackoff.InitialInterval = r.config.InitialDelay
	expBackoff.MaxInterval = r.config.MaxDelay
	expBackoff.Multiplier = r.config.BackoffMultiplier
	expBackoff.MaxElapsedTime = 0 // No max elapsed time, controlled by MaxRetries

	var lastResponse *http.Response
	var lastErr error
	attempt := 0

	for attempt <= r.config.MaxRetries {
		// Clone request for retry (body needs to be preserved)
		reqClone := req.Clone(ctx)

		// Execute request
		resp, err := r.client.Do(reqClone)
		lastResponse = resp
		lastErr = err

		// Success - no retry needed
		if err == nil && !r.shouldRetry(resp.StatusCode) {
			span.SetAttributes(
				observability.Any("http.status_code", resp.StatusCode),
				observability.Any("http.attempts", attempt + 1),
			)
			return resp, nil
		}

		// Check if we should retry
		shouldRetry := false
		if err != nil {
			// Network error - always retry
			shouldRetry = true
			span.AddEvent("http_request_error",
				observability.Any("attempt", attempt + 1),
				observability.Any("error", err.Error()),
			)
		} else if r.shouldRetry(resp.StatusCode) {
			// HTTP error with retryable status code
			shouldRetry = true
			span.AddEvent("http_retryable_status",
				observability.Any("attempt", attempt + 1),
				observability.Any("status_code", resp.StatusCode),
			)
			// Close response body to avoid resource leak
			if resp.Body != nil {
				if err := resp.Body.Close(); err != nil {
					span.AddEvent("error closing response body", observability.Any("error", err.Error()))
				}
			}
		}

		// No more retries available
		if !shouldRetry || attempt >= r.config.MaxRetries {
			if err != nil {
				span.SetAttributes(observability.Any("error", err.Error()))
			} else if resp != nil {
				span.SetAttributes(observability.Any("http.status_code", resp.StatusCode))
			}
			span.SetAttributes(observability.Any("http.attempts", attempt + 1))
			return lastResponse, lastErr
		}

		// Calculate delay with jitter
		delay := expBackoff.NextBackOff()
		if delay == backoff.Stop {
			delay = r.config.MaxDelay
		}

		// Add jitter (±25% randomization) using crypto-safe random for better distribution
		jitter := time.Duration(float64(delay) * 0.25 * (rand.Float64()*2 - 1))
		delay += jitter

		statusCodeValue := 0
		if lastResponse != nil {
			statusCodeValue = lastResponse.StatusCode
		}

		span.AddEvent("http_retry_attempt",
			observability.Any("attempt", attempt + 1),
			observability.Any("delay", delay.String()),
			observability.Any("status_code", statusCodeValue),
		)

		// Wait before retry (respect context cancellation)
		select {
		case <-ctx.Done():
			span.SetAttributes(observability.Any("error", "context cancelled during retry"))
			return nil, ctx.Err()
		case <-time.After(delay):
			attempt++
		}
	}

	return lastResponse, lastErr
}

func (r *retryClient) shouldRetry(statusCode int) bool {
	_, ok := r.retryableStatusCode.Load(statusCode)
	return ok
}

// DefaultRetryConfig returns a default retry configuration.
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:        3,
		InitialDelay:      100 * time.Millisecond,
		MaxDelay:          2 * time.Second,
		BackoffMultiplier: 2.0,
		RetryableStatusCodes: map[int]bool{
			408: true, // Request Timeout
			429: true, // Too Many Requests
			500: true, // Internal Server Error
			502: true, // Bad Gateway
			503: true, // Service Unavailable
			504: true, // Gateway Timeout
		},
	}
}

// ParseRetryableStatusCodes parses a comma-separated string of status codes into a map.
func ParseRetryableStatusCodes(codes string) map[int]bool {
	result := make(map[int]bool)
	if codes == "" {
		return result
	}

	var code int
	for i := 0; i < len(codes); {
		n, err := fmt.Sscanf(codes[i:], "%d", &code)
		if err != nil || n != 1 {
			break
		}
		result[code] = true

		// Find next comma or end
		for i < len(codes) && codes[i] != ',' {
			i++
		}
		i++ // Skip comma
	}

	return result
}
