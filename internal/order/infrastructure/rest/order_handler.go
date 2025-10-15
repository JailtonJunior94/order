package rest

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/order/internal/order/domain/dtos"
	"github.com/jailtonjunior94/order/internal/order/usecase"
	"github.com/jailtonjunior94/order/pkg/o11y"
	"github.com/jailtonjunior94/order/pkg/responses"
	"github.com/jailtonjunior94/order/pkg/vos"

	"github.com/go-chi/chi/v5"
)

type UserHandler struct {
	telemetry         o11y.Telemetry
	createUseCase     usecase.CreateOrderUseCase
	markAsPaidUseCase usecase.MarkAsPaidUseCase
}

func NewUserHandler(
	telemetry o11y.Telemetry,
	createUseCase usecase.CreateOrderUseCase,
	markAsPaidUseCase usecase.MarkAsPaidUseCase,
) *UserHandler {
	return &UserHandler{
		telemetry:         telemetry,
		createUseCase:     createUseCase,
		markAsPaidUseCase: markAsPaidUseCase,
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.telemetry.Tracer().Start(r.Context(), "order_handler.create")
	defer span.End()

	var input *dtos.OrderInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		span.AddEvent("error decoding request body", o11y.Attribute{Key: "error", Value: err})
		responses.Error(w, http.StatusUnprocessableEntity, "unprocessable Entity")
		return
	}

	output, err := h.createUseCase.Execute(ctx, input)
	if err != nil {
		span.AddEvent("error creating order", o11y.Attribute{Key: "error", Value: err})
		responses.Error(w, http.StatusBadRequest, "error creating order")
		return
	}
	responses.JSON(w, http.StatusCreated, output)
}

func (h *UserHandler) MarkAsPaid(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.telemetry.Tracer().Start(r.Context(), "order_handler.mark_as_paid")
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
		span.AddEvent("error updating order", o11y.Attribute{Key: "error", Value: err})
		responses.Error(w, http.StatusBadRequest, "error updating order")
		return
	}
	responses.JSON(w, http.StatusOK, output)
}
