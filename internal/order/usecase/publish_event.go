package usecase

import (
	"context"

	"github.com/jailtonjunior94/order/configs"
	"github.com/jailtonjunior94/order/internal/order/domain/entities"
	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/pkg/database"
	"github.com/jailtonjunior94/order/pkg/database/uow"
	"github.com/jailtonjunior94/order/pkg/messaging/kafka"
	"github.com/jailtonjunior94/order/pkg/observability"
)

type (
	PublishEventUseCase interface {
		Execute(ctx context.Context) error
	}

	publishEventUseCase struct {
		config            *configs.Config
		uow               uow.UnitOfWork
		brokerClient      kafka.KafkaClient
		o11y               observability.Observability
		repositoryFactory interfaces.RepositoryFactory
	}
)

func NewPublishEventUseCase(
	uow uow.UnitOfWork,
	config *configs.Config,
	o11y observability.Observability,
	brokerClient kafka.KafkaClient,
	repositoryFactory interfaces.RepositoryFactory,
) PublishEventUseCase {
	return &publishEventUseCase{
		uow:               uow,
		config:            config,
		o11y:               o11y,
		brokerClient:      brokerClient,
		repositoryFactory: repositoryFactory,
	}
}

func (c *publishEventUseCase) Execute(ctx context.Context) error {
	ctx, span := c.o11y.Tracer().Start(ctx, "publish_event_usecase.execute")
	defer span.End()

	// Step 1: Fetch unpublished events in a read-only transaction
	eventsToPublish, err := c.fetchUnpublishedEvents(ctx)
	if err != nil {
		span.AddEvent("error fetching unpublished events", observability.Error(err))
		return err
	}

	if len(eventsToPublish) == 0 {
		c.o11y.Logger().Info(ctx, "no events to publish")
		return nil
	}

	c.o11y.Logger().Info(ctx, "events_to_publish",
		observability.Any("count", len(eventsToPublish)))

	// Step 2: Publish each event to Kafka and mark as published (OUTSIDE the fetch transaction)
	// This prevents database rollback from affecting already-published Kafka messages
	successCount := 0
	for _, event := range eventsToPublish {
		if err := c.publishAndMarkAsPublished(ctx, event); err != nil {
			span.AddEvent("error publishing event",
				observability.Any("error", err),
				observability.Any("event_id", event.ID.String()))

			// Log error but continue with other events to maximize throughput
			c.o11y.Logger().Error(ctx, "failed to publish event",
				observability.Error(err),
				observability.String("event_id", event.ID.String()))

			// Return error to trigger retry of failed events
			return err
		}
		successCount++
	}

	counter := c.o11y.Metrics().Counter("outbox_events_published_total", "Total outbox events published", "1")
	counter.Add(ctx, int64(successCount))

	c.o11y.Logger().Info(ctx, "events published successfully",
		observability.Int("count", successCount))

	return nil
}

// fetchUnpublishedEvents retrieves all unpublished events in a read-only transaction.
func (c *publishEventUseCase) fetchUnpublishedEvents(ctx context.Context) ([]*entities.Outbox, error) {
	var events []*entities.Outbox

	err := c.uow.Do(ctx, func(ctx context.Context, db database.DBTX) error {
		outboxRepository := c.repositoryFactory.OutboxRepository(db, c.o11y)

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

// publishAndMarkAsPublished publishes a single event to Kafka and marks it as published.
// This is done in a separate transaction to ensure atomicity per event.
func (c *publishEventUseCase) publishAndMarkAsPublished(ctx context.Context, event *entities.Outbox) error {
	ctx, span := c.o11y.Tracer().Start(ctx, "publish_event_usecase.publish_and_mark")
	defer span.End()

	span.AddEvent("publishing event",
		observability.String("event_id", event.ID.String()),
		observability.String("event_name", event.EventName))

	// Step 1: Publish to Kafka FIRST (outside transaction)
	headers := map[string]string{"event_name": event.EventName}
	message := &kafka.Message{
		Key:   []byte(event.ID.String()),
		Value: []byte(event.Payload),
	}

	if err := c.brokerClient.Produce(ctx, c.config.KafkaConfig.Order, headers, message); err != nil {
		span.AddEvent("kafka produce failed", observability.Error(err))

		errorCounter := c.o11y.Metrics().Counter("outbox_kafka_publish_errors_total", "Total Kafka publish errors", "1")
		errorCounter.Add(ctx, 1, observability.String("event_name", event.EventName))
		return err
	}

	span.AddEvent("kafka produce succeeded")

	publishedCounter := c.o11y.Metrics().Counter("outbox_kafka_published_total", "Total events published to Kafka", "1")
	publishedCounter.Add(ctx, 1, observability.String("event_name", event.EventName))

	// Step 2: Mark as published in a NEW separate transaction
	// If this fails, the event will be republished (idempotency required in consumer)
	err := c.uow.Do(ctx, func(ctx context.Context, db database.DBTX) error {
		outboxRepository := c.repositoryFactory.OutboxRepository(db, c.o11y)

		if err := outboxRepository.Update(ctx, event.MarkAsPublished()); err != nil {
			span.AddEvent("database update failed", observability.Error(err))
			return err
		}

		return nil
	})

	if err != nil {
		// WARNING: Event was published to Kafka but NOT marked as published in DB
		// This will cause re-publication on next execution
		// Consumer MUST be idempotent to handle duplicate messages
		c.o11y.Logger().Warn(ctx, "event published to kafka but failed to mark as published - will retry",
			observability.String("event_id", event.ID.String()),
			observability.String("event_name", event.EventName))

		markErrorCounter := c.o11y.Metrics().Counter("outbox_mark_published_errors_total", "Total mark as published errors", "1")
		markErrorCounter.Add(ctx, 1, observability.String("event_name", event.EventName))

		return err
	}

	span.AddEvent("marked as published")
	return nil
}
