package httpclient

import (
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
)

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

func NewHTTPClient() HTTPClient {
	return &http.Client{
		Transport: otelhttp.NewTransport(http.DefaultTransport),
	}
}
