package usecase

import (
	"context"

	"github.com/jailtonjunior94/order/configs"
	"github.com/jailtonjunior94/order/internal/order/domain/entities"
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

	// Step 1: Fetch unpublished events in a read-only transaction
	eventsToPublish, err := c.fetchUnpublishedEvents(ctx)
	if err != nil {
		span.AddEvent("error fetching unpublished events", o11y.Attribute{Key: "error", Value: err})
		return err
	}

	if len(eventsToPublish) == 0 {
		c.telemetry.Logger().Info(ctx, "no events to publish")
		return nil
	}

	c.telemetry.Logger().Info(ctx, "events_to_publish",
		o11y.Field{Key: "count", Value: len(eventsToPublish)})

	// Step 2: Publish each event to Kafka and mark as published (OUTSIDE the fetch transaction)
	// This prevents database rollback from affecting already-published Kafka messages
	successCount := 0
	for _, event := range eventsToPublish {
		if err := c.publishAndMarkAsPublished(ctx, event); err != nil {
			span.AddEvent("error publishing event",
				o11y.Attribute{Key: "error", Value: err},
				o11y.Attribute{Key: "event_id", Value: event.ID.String()})

			// Log error but continue with other events to maximize throughput
			c.telemetry.Logger().Error(ctx, err, "failed to publish event",
				o11y.Field{Key: "event_id", Value: event.ID.String()})

			// Return error to trigger retry of failed events
			return err
		}
		successCount++
	}

	c.telemetry.Metrics().AddCounter(ctx, "outbox_events_published_total", int64(successCount), nil)
	c.telemetry.Logger().Info(ctx, "events published successfully",
		o11y.Field{Key: "count", Value: successCount})

	return nil
}

// fetchUnpublishedEvents retrieves all unpublished events in a read-only transaction
func (c *publishEventUseCase) fetchUnpublishedEvents(ctx context.Context) ([]*entities.Outbox, error) {
	var events []*entities.Outbox

	err := c.uow.Do(ctx, func(ctx context.Context, db database.DBTX) error {
		outboxRepository := c.repositoryFactory.OutboxRepository(db, c.telemetry)

		fetchedEvents, err := outboxRepository.FindAll(ctx, false)
		if err != nil {
			return err
		}

		events = fetchedEvents
		return nil
	})

	if err != nil {
		return nil, err
	}

	return events, nil
}

// publishAndMarkAsPublished publishes a single event to Kafka and marks it as published
// This is done in a separate transaction to ensure atomicity per event
func (c *publishEventUseCase) publishAndMarkAsPublished(ctx context.Context, event *entities.Outbox) error {
	ctx, span := c.telemetry.Tracer().Start(ctx, "publish_event_usecase.publish_and_mark")
	defer span.End()

	span.AddEvent("publishing event",
		o11y.Attribute{Key: "event_id", Value: event.ID.String()},
		o11y.Attribute{Key: "event_name", Value: event.EventName})

	// Step 1: Publish to Kafka FIRST (outside transaction)
	headers := map[string]string{"event_name": event.EventName}
	message := &kafka.Message{
		Key:   []byte(event.ID.String()),
		Value: []byte(event.Payload),
	}

	if err := c.brokerClient.Produce(ctx, c.config.KafkaConfig.Order, headers, message); err != nil {
		span.AddEvent("kafka produce failed", o11y.Attribute{Key: "error", Value: err})
		c.telemetry.Metrics().AddCounter(ctx, "outbox_kafka_publish_errors_total", int64(1),
			map[string]string{"event_name": event.EventName})
		return err
	}

	span.AddEvent("kafka produce succeeded")
	c.telemetry.Metrics().AddCounter(ctx, "outbox_kafka_published_total", int64(1),
		map[string]string{"event_name": event.EventName})

	// Step 2: Mark as published in a NEW separate transaction
	// If this fails, the event will be republished (idempotency required in consumer)
	err := c.uow.Do(ctx, func(ctx context.Context, db database.DBTX) error {
		outboxRepository := c.repositoryFactory.OutboxRepository(db, c.telemetry)

		if err := outboxRepository.Update(ctx, event.MarkAsPublished()); err != nil {
			span.AddEvent("database update failed", o11y.Attribute{Key: "error", Value: err})
			return err
		}

		return nil
	})

	if err != nil {
		// WARNING: Event was published to Kafka but NOT marked as published in DB
		// This will cause re-publication on next execution
		// Consumer MUST be idempotent to handle duplicate messages
		c.telemetry.Logger().Warn(ctx, "event published to kafka but failed to mark as published - will retry",
			o11y.Field{Key: "event_id", Value: event.ID.String()},
			o11y.Field{Key: "event_name", Value: event.EventName})

		c.telemetry.Metrics().AddCounter(ctx, "outbox_mark_published_errors_total", int64(1),
			map[string]string{"event_name": event.EventName})

		return err
	}

	span.AddEvent("marked as published")
	return nil
}
