package rest

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/jailtonjunior94/order/internal/order/domain/dtos"
	domainErrors "github.com/jailtonjunior94/order/internal/order/domain/errors"
	"github.com/jailtonjunior94/order/internal/order/usecase"
	"github.com/jailtonjunior94/order/pkg/observability"
	"github.com/jailtonjunior94/order/pkg/responses"
	"github.com/jailtonjunior94/order/pkg/vos"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	o11y               observability.Observability
	createUseCase     usecase.CreateOrderUseCase
	markAsPaidUseCase usecase.MarkAsPaidUseCase
}

func NewUserHandler(
	o11y observability.Observability,
	createUseCase usecase.CreateOrderUseCase,
	markAsPaidUseCase usecase.MarkAsPaidUseCase,
) *UserHandler {
	return &UserHandler{
		o11y:               o11y,
		createUseCase:     createUseCase,
		markAsPaidUseCase: markAsPaidUseCase,
	}
}

// mapDomainErrorToHTTPStatus maps domain errors to HTTP status codes.
func mapDomainErrorToHTTPStatus(err error) int {
	switch {
	case errors.Is(err, domainErrors.ErrClientNotFound):
		return http.StatusNotFound
	case errors.Is(err, domainErrors.ErrClientInactive):
		return http.StatusForbidden
	case errors.Is(err, domainErrors.ErrEmptyClientID):
		return http.StatusBadRequest
	case errors.Is(err, domainErrors.ErrClientServiceTimeout):
		return http.StatusGatewayTimeout
	case errors.Is(err, domainErrors.ErrClientServiceUnavailable):
		return http.StatusServiceUnavailable
	default:
		return http.StatusInternalServerError
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "order_handler.create")
	defer span.End()

	var input *dtos.OrderInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		span.AddEvent("error decoding request body", observability.Any("error", err))
		responses.Error(w, http.StatusUnprocessableEntity, "unprocessable Entity")
		return
	}

	output, err := h.createUseCase.Execute(ctx, input)
	if err != nil {
		span.AddEvent("error creating order", observability.Any("error", err))
		statusCode := mapDomainErrorToHTTPStatus(err)
		responses.Error(w, statusCode, err.Error())
		return
	}
	responses.JSON(w, http.StatusCreated, output)
}

func (h *UserHandler) MarkAsPaid(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "order_handler.mark_as_paid")
	defer span.End()

	orderIDParam := chi.URLParam(r, "id")
	if orderIDParam == "" {
		responses.Error(w, http.StatusUnprocessableEntity, "order_id is required")
		return
	}

	orderID, err := vos.NewUUIDFromString(orderIDParam)
	if err != nil {
		responses.Error(w, http.StatusUnprocessableEntity, "order id is invalid")
		return
	}

	output, err := h.markAsPaidUseCase.Execute(ctx, orderID)
	if err != nil {
		span.AddEvent("error updating order", observability.Any("error", err))
		responses.Error(w, http.StatusBadRequest, "error updating order")
		return
	}
	responses.JSON(w, http.StatusOK, output)
}
