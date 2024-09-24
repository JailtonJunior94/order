package httpclient

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
)

func MakeRequest[TSuccess any, TError any](ctx context.Context, client HTTPClient, method, url string, headers map[string]string, payload io.Reader) (int, *TSuccess, *TError, error) {
	request, err := http.NewRequestWithContext(ctx, method, url, payload)
	if err != nil {
		return http.StatusInternalServerError, nil, nil, err
	}

	for key, value := range headers {
		request.Header.Add(key, value)
	}

	response, err := client.Do(request)
	if err != nil {
		return response.StatusCode, nil, nil, err
	}

	if response != nil {
		defer response.Body.Close()
	}

	if response.StatusCode < 200 || response.StatusCode > 299 {
		var errorResponse *TError
		if err := json.NewDecoder(response.Body).Decode(&errorResponse); err != nil {
			return http.StatusInternalServerError, nil, nil, err
		}
		return response.StatusCode, nil, errorResponse, nil
	}

	var successResponse *TSuccess
	if err := json.NewDecoder(response.Body).Decode(&successResponse); err != nil {
		return http.StatusInternalServerError, nil, nil, err
	}
	return response.StatusCode, successResponse, nil, nil
}
