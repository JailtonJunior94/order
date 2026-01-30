package httpclient

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

const (
	// DefaultMaxResponseSize is the default maximum response body size (10MB).
	DefaultMaxResponseSize int64 = 10 * 1024 * 1024
)

var (
	// ErrResponseTooLarge is returned when the response body exceeds the maximum size.
	ErrResponseTooLarge = errors.New("response body exceeds maximum allowed size")
)

// MakeRequest performs an HTTP request and decodes the response.
// The response body is limited to DefaultMaxResponseSize (10MB) to prevent memory exhaustion.
// Returns: statusCode, successResponse, errorResponse, error
func MakeRequest[TSuccess any, TError any](ctx context.Context, client HTTPClient, method, url string, headers map[string]string, payload io.Reader) (int, *TSuccess, *TError, error) {
	return MakeRequestWithLimit[TSuccess, TError](ctx, client, method, url, headers, payload, DefaultMaxResponseSize)
}

// MakeRequestWithLimit performs an HTTP request with a custom response body size limit.
// Set maxBodySize to 0 or negative for no limit (not recommended).
// Returns: statusCode, successResponse, errorResponse, error
func MakeRequestWithLimit[TSuccess any, TError any](ctx context.Context, client HTTPClient, method, url string, headers map[string]string, payload io.Reader, maxBodySize int64) (int, *TSuccess, *TError, error) {
	request, err := http.NewRequestWithContext(ctx, method, url, payload)
	if err != nil {
		return 0, nil, nil, err
	}

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	response, err := client.Do(request)
	if err != nil {
		return 0, nil, nil, err
	}

	if response != nil {
		defer func() {
			_ = response.Body.Close()
		}()
	}

	statusCode := response.StatusCode

	// Limit the response body size to prevent memory exhaustion attacks
	var bodyReader io.Reader = response.Body
	if maxBodySize > 0 {
		bodyReader = io.LimitReader(response.Body, maxBodySize+1)
	}

	if statusCode < 200 || statusCode > 299 {
		var errorResponse *TError
		if err := json.NewDecoder(bodyReader).Decode(&errorResponse); err != nil {
			return statusCode, nil, nil, err
		}
		return statusCode, nil, errorResponse, nil
	}

	var successResponse *TSuccess
	if err := json.NewDecoder(bodyReader).Decode(&successResponse); err != nil {
		return statusCode, nil, nil, err
	}
	return statusCode, successResponse, nil, nil
}
