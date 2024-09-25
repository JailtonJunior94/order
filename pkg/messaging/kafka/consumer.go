package kafka

import (
	"context"
	"log"
	"time"

	"github.com/jailtonjunior94/outbox/pkg/o11y"
	"go.opentelemetry.io/otel/propagation"

	"github.com/cenkalti/backoff/v4"
	"github.com/segmentio/kafka-go"
)

type (
	ConsumerOptions func(consumer *consumer)
	ConsumeHandler  func(ctx context.Context, body []byte) error

	Consumer interface {
		Consume(ctx context.Context, handler ConsumeHandler) error
	}

	consumer struct {
		retries    int
		maxRetries int
		topic      string
		groupID    string
		brokers    []string
		reader     *kafka.Reader
		handler    ConsumeHandler
		backoff    backoff.BackOff
		retryChan  chan kafka.Message
		o11y       o11y.Observability
	}
)

func NewConsumer(o11y o11y.Observability, options ...ConsumerOptions) Consumer {
	consumer := &consumer{o11y: o11y}
	for _, opt := range options {
		opt(consumer)
	}
	return consumer
}

func (c *consumer) Consume(ctx context.Context, handler ConsumeHandler) error {
	go func() {
		for {
			msg, err := c.reader.ReadMessage(ctx)
			if err != nil {
				log.Fatal("failed to read message:", err)
				continue
			}

			tracingHeader := map[string][]string{}
			for _, header := range msg.Headers {
				if header.Key == "traceID" {
					tracingHeader["Traceparent"] = []string{string(header.Value)}
					break
				}
			}

			propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
			ctx = propagator.Extract(ctx, propagation.HeaderCarrier(tracingHeader))

			if err := c.dispatcher(ctx, msg, handler); err != nil {
				log.Fatal("failed to dispatch message:", err)
				continue
			}
		}
	}()
	return nil
}

func WithTopic(name string) ConsumerOptions {
	return func(consumer *consumer) {
		consumer.topic = name
	}
}

func WithBrokers(brokers []string) ConsumerOptions {
	return func(consumer *consumer) {
		consumer.brokers = brokers
	}
}

func WithGroupID(groupID string) ConsumerOptions {
	return func(consumer *consumer) {
		consumer.groupID = groupID
	}
}

func WithMaxRetries(maxRetries int) ConsumerOptions {
	return func(consumer *consumer) {
		consumer.maxRetries = maxRetries
	}
}

func WithRetryChan(sizeChan int) ConsumerOptions {
	return func(consumer *consumer) {
		consumer.retryChan = make(chan kafka.Message, sizeChan)
	}
}

func WithBackoff(backoff backoff.BackOff) ConsumerOptions {
	return func(consumer *consumer) {
		consumer.backoff = backoff
	}
}

func WithReader() ConsumerOptions {
	return func(consumer *consumer) {
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:        consumer.brokers,
			GroupID:        consumer.groupID,
			Topic:          consumer.topic,
			MinBytes:       10e3,
			MaxBytes:       10e6,
			CommitInterval: 0,
			StartOffset:    kafka.FirstOffset,
		})
		consumer.reader = reader
	}
}

func WithHandler(handler ConsumeHandler) ConsumerOptions {
	return func(consumer *consumer) {
		consumer.handler = handler
	}
}

func (c *consumer) dispatcher(ctx context.Context, message kafka.Message, handler ConsumeHandler) error {
	ctx, span := c.o11y.Start(ctx, "consumer.consume")
	defer span.End()

	err := handler(ctx, message.Value)
	if err != nil {
		c.retries++
		return c.retry(ctx, message)
	}

	if err := c.reader.CommitMessages(ctx, message); err != nil {
		return err
	}
	return nil
}

func (c *consumer) retry(ctx context.Context, message kafka.Message) error {
	c.retryChan <- message
	go func() {
		for msg := range c.retryChan {
			if c.retries >= c.maxRetries {
				log.Printf("max retries reached for message %s", msg.Value)
				break
			}

			if err := c.dispatcher(ctx, msg, c.handler); err != nil {
				time.Sleep(c.backoff.NextBackOff())
				continue
			}
			break
		}
	}()
	return nil
}
