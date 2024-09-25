package rest

import (
	"encoding/json"
	"net/http"

	"github.com/jailtonjunior94/outbox/internal/order/domain/dtos"
	"github.com/jailtonjunior94/outbox/internal/order/usecase"
	"github.com/jailtonjunior94/outbox/pkg/o11y"
	"github.com/jailtonjunior94/outbox/pkg/responses"
)

type UserHandler struct {
	o11y          o11y.Observability
	createUseCase usecase.CreateOrderUseCase
}

func NewUserHandler(
	o11y o11y.Observability,
	createUseCase usecase.CreateOrderUseCase,
) *UserHandler {
	return &UserHandler{
		o11y:          o11y,
		createUseCase: createUseCase,
	}
}

func (h *UserHandler) Create(w http.ResponseWriter, r *http.Request) {
	ctx, span := h.o11y.Tracer().Start(r.Context(), "order_handler.create")
	defer span.End()

	var input *dtos.OrderInput
	err := json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		span.RecordError(err)
		responses.Error(w, http.StatusUnprocessableEntity, "Unprocessable Entity")
		return
	}

	output, err := h.createUseCase.Execute(ctx, input)
	if err != nil {
		span.RecordError(err)
		responses.Error(w, http.StatusBadRequest, "error creating order")
		return
	}
	responses.JSON(w, http.StatusCreated, output)
}
