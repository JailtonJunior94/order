package job

import (
	"context"

	"github.com/jailtonjunior94/order/internal/order/usecase"
	"github.com/jailtonjunior94/order/pkg/observability"
)

type PublishEventHandler struct {
	o11y          observability.Observability
	publishEvent usecase.PublishEventUseCase
}

func NewPublishEventHandler(
	o11y observability.Observability,
	publishEvent usecase.PublishEventUseCase,
) *PublishEventHandler {
	return &PublishEventHandler{
		o11y:          o11y,
		publishEvent: publishEvent,
	}
}

func (h *PublishEventHandler) Handle() {
	ctx, span := h.o11y.Tracer().Start(context.Background(), "publish_event_handler.handle")
	defer span.End()

	if err := h.publishEvent.Execute(ctx); err != nil {
		span.AddEvent("error publish event", observability.Any("error", err))
		return
	}
	span.AddEvent("event published successfully", observability.Any("status", "ok"))
}
