package services

import (
	"context"
	"errors"
	"fmt"
	"net/http"

	"github.com/jailtonjunior94/order/configs"
	domainErrors "github.com/jailtonjunior94/order/internal/order/domain/errors"
	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	httpclient "github.com/jailtonjunior94/order/pkg/http-client"
	"github.com/jailtonjunior94/order/pkg/observability"
)

type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

type clientService struct {
	httpClient httpclient.HTTPClient
	config     configs.ClientServiceConfig
	o11y        observability.Observability
}

// NewClientService creates a new client service instance.
func NewClientService(
	httpClient httpclient.HTTPClient,
	o11y observability.Observability,
	config configs.ClientServiceConfig,
) interfaces.ClientService {
	return &clientService{
		httpClient: httpClient,
		config:     config,
		o11y:        o11y,
	}
}

func (s *clientService) GetClient(ctx context.Context, clientID string) (*interfaces.ClientDTO, error) {
	ctx, span := s.o11y.Tracer().Start(ctx, "client_service.get_client")
	defer span.End()

	span.AddEvent("calling client service", observability.Any("client_id", clientID))

	url := fmt.Sprintf("%s/clients/%s", s.config.BaseURL, clientID)

	statusCode, success, errResp, err := httpclient.MakeRequest[interfaces.ClientDTO, ErrorResponse](
		ctx,
		s.httpClient,
		http.MethodGet,
		url,
		map[string]string{"Content-Type": "application/json"},
		nil,
	)

	// Handle network/timeout errors
	if err != nil {
		span.AddEvent("error calling client service", observability.Any("error", err))

		// Check for timeout/context deadline
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, domainErrors.ErrClientServiceTimeout
		}

		return nil, &domainErrors.DomainError{
			Code:    domainErrors.ErrClientServiceUnavailable.Code,
			Message: domainErrors.ErrClientServiceUnavailable.Message,
			Err:     err,
		}
	}

	// Handle HTTP error status codes
	if statusCode != http.StatusOK {
		span.AddEvent("client service returned error",
			observability.Any("status_code", statusCode),
			observability.Any("error_response", errResp),
		)

		switch statusCode {
		case http.StatusNotFound:
			return nil, domainErrors.ErrClientNotFound
		case http.StatusRequestTimeout, http.StatusServiceUnavailable, http.StatusGatewayTimeout:
			return nil, domainErrors.ErrClientServiceUnavailable
		default:
			return nil, &domainErrors.DomainError{
				Code:    domainErrors.ErrClientServiceUnavailable.Code,
				Message: fmt.Sprintf("client service returned status %d", statusCode),
			}
		}
	}

	// Validate client is active
	if !success.Active {
		span.AddEvent("client is inactive", observability.Any("client_id", clientID))
		return nil, domainErrors.ErrClientInactive
	}

	span.AddEvent("client retrieved successfully",
		observability.Any("client_id", success.ID),
		observability.Any("client_name", success.Name),
	)

	return success, nil
}
