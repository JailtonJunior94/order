package job

import (
	"context"

	"github.com/jailtonjunior94/order/internal/order/usecase"
	"github.com/jailtonjunior94/order/pkg/o11y"
)

type PublishEventHandler struct {
	o11y         o11y.Observability
	publishEvent usecase.PublishEventUseCase
}

func NewPublishEventHandler(
	o11y o11y.Observability,
	publishEvent usecase.PublishEventUseCase,
) *PublishEventHandler {
	return &PublishEventHandler{
		o11y:         o11y,
		publishEvent: publishEvent,
	}
}

func (h *PublishEventHandler) Handle() {
	ctx, span := h.o11y.Start(context.Background(), "publish_event_handler.handle")
	defer span.End()

	if err := h.publishEvent.Execute(ctx); err != nil {
		span.AddAttributes(ctx, o11y.Error, "error publish event", o11y.Attributes{Key: "error", Value: err})
		return
	}
	span.AddAttributes(ctx, o11y.Ok, "")
}
