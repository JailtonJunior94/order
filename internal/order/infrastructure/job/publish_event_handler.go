package job

import (
	"context"

	"github.com/jailtonjunior94/order/internal/order/usecase"
	"github.com/jailtonjunior94/order/pkg/o11y"
)

type PublishEventHandler struct {
	telemetry    o11y.Telemetry
	publishEvent usecase.PublishEventUseCase
}

func NewPublishEventHandler(
	telemetry o11y.Telemetry,
	publishEvent usecase.PublishEventUseCase,
) *PublishEventHandler {
	return &PublishEventHandler{
		telemetry:    telemetry,
		publishEvent: publishEvent,
	}
}

func (h *PublishEventHandler) Handle() {
	ctx, span := h.telemetry.Tracer().Start(context.Background(), "publish_event_handler.handle")
	defer span.End()

	if err := h.publishEvent.Execute(ctx); err != nil {
		span.AddEvent("error publish event", o11y.Attribute{Key: "error", Value: err})
		return
	}
	span.AddEvent("event published successfully", o11y.Attribute{Key: "status", Value: "ok"})
}
