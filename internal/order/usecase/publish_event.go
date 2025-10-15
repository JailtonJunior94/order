package usecase

import (
	"context"

	"github.com/jailtonjunior94/order/configs"
	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/pkg/database"
	"github.com/jailtonjunior94/order/pkg/database/uow"
	"github.com/jailtonjunior94/order/pkg/messaging/kafka"
	"github.com/jailtonjunior94/order/pkg/o11y"
)

type (
	PublishEventUseCase interface {
		Execute(ctx context.Context) error
	}

	publishEventUseCase struct {
		config            *configs.Config
		uow               uow.UnitOfWork
		brokerClient      kafka.KafkaClient
		telemetry         o11y.Telemetry
		repositoryFactory interfaces.RepositoryFactory
	}
)

func NewPublishEventUseCase(
	uow uow.UnitOfWork,
	config *configs.Config,
	telemetry o11y.Telemetry,
	brokerClient kafka.KafkaClient,
	repositoryFactory interfaces.RepositoryFactory,
) PublishEventUseCase {
	return &publishEventUseCase{
		uow:               uow,
		config:            config,
		telemetry:         telemetry,
		brokerClient:      brokerClient,
		repositoryFactory: repositoryFactory,
	}
}

func (c *publishEventUseCase) Execute(ctx context.Context) error {
	ctx, span := c.telemetry.Tracer().Start(ctx, "publish_event_usecase.execute")
	defer span.End()

	return c.uow.Do(ctx, func(ctx context.Context, db database.DBTX) error {
		outboxRepository := c.repositoryFactory.OutboxRepository(db, c.telemetry)
		eventsToPublish, err := outboxRepository.FindAll(ctx, false)
		if err != nil {
			span.AddEvent("error find all events to publish", o11y.Attribute{Key: "error", Value: err})
			return err
		}

		for _, event := range eventsToPublish {
			headers := map[string]string{"event_name": event.EventName}
			message := &kafka.Message{
				Key:   []byte(event.ID.String()),
				Value: []byte(event.Payload),
			}

			if err := c.brokerClient.Produce(ctx, c.config.KafkaConfig.Order, headers, message); err != nil {
				span.AddEvent("error produce event", o11y.Attribute{Key: "error", Value: err})
				return err
			}

			if err := outboxRepository.Update(ctx, event.MarkAsPublished()); err != nil {
				span.AddEvent("error update status event", o11y.Attribute{Key: "error", Value: err})
			}
		}

		return nil
	})
}
