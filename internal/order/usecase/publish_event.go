package usecase

import (
	"context"

	"github.com/jailtonjunior94/order/configs"
	"github.com/jailtonjunior94/order/pkg/database/uow"
	"github.com/jailtonjunior94/order/pkg/messaging/kafka"
	"github.com/jailtonjunior94/order/pkg/o11y"
)

type (
	PublishEventUseCase interface {
		Execute(ctx context.Context) error
	}

	publishEventUseCase struct {
		config       *configs.Config
		uow          uow.UnitOfWork
		brokerClient kafka.KafkaClient
		o11y         o11y.Observability
	}
)

func NewPublishEventUseCase(
	config *configs.Config,
	uow uow.UnitOfWork,
	brokerClient kafka.KafkaClient,
	o11y o11y.Observability,
) PublishEventUseCase {
	return &publishEventUseCase{
		uow:          uow,
		o11y:         o11y,
		config:       config,
		brokerClient: brokerClient,
	}
}

func (c *publishEventUseCase) Execute(ctx context.Context) error {
	ctx, span := c.o11y.Start(ctx, "publish_event_usecase.execute")
	defer span.End()

	return c.uow.Do(ctx, func(ctx context.Context, tx uow.TX) error {
		outboxRepository, err := GetOutboxRepository(tx)
		if err != nil {
			span.AddAttributes(ctx, o11y.Error, "error get outbox repository", o11y.Attributes{Key: "error", Value: err})
			return err
		}

		eventsToPublish, err := outboxRepository.FindAll(ctx, false)
		if err != nil {
			span.AddAttributes(ctx, o11y.Error, "error find all events to publish", o11y.Attributes{Key: "error", Value: err})
			return err
		}

		for _, event := range eventsToPublish {
			headers := map[string]string{"event_name": event.EventName}
			message := &kafka.Message{
				Key:   []byte(event.ID.String()),
				Value: []byte(event.Payload),
			}

			if err := c.brokerClient.Produce(ctx, c.config.KafkaConfig.Order, headers, message); err != nil {
				span.AddAttributes(ctx, o11y.Error, "error produce event", o11y.Attributes{Key: "error", Value: err})
				return err
			}

			if err := outboxRepository.Update(ctx, event.MarkAsPublished()); err != nil {
				span.AddAttributes(ctx, o11y.Error, "error update status event", o11y.Attributes{Key: "error", Value: err})
			}
		}

		return nil
	})
}
