package kafka

import (
	"context"

	"github.com/jailtonjunior94/order/pkg/observability"

	"github.com/segmentio/kafka-go"
	"go.opentelemetry.io/otel/propagation"
)

type (
	KafkaClient interface {
		Produce(ctx context.Context, topic string, headers map[string]string, message *Message) error
	}

	kafkaClient struct {
		client    *kafka.Writer
		o11y observability.Observability
	}

	Message struct {
		Key   []byte
		Value []byte
	}
)

func NewKafkaClient(
	broker string,
	o11y observability.Observability,
) KafkaClient {
	client := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Balancer: &kafka.LeastBytes{},
	}
	return &kafkaClient{o11y: o11y, client: client}
}

func (k *kafkaClient) Produce(ctx context.Context, topic string, headers map[string]string, message *Message) error {
	ctx, span := k.o11y.Tracer().Start(ctx, "producer.produce")
	defer span.End()

	messageKafka := kafka.Message{
		Topic: topic,
		Key:   message.Key,
		Value: message.Value,
	}

	tracingHeader := map[string][]string{}
	propagator := propagation.NewCompositeTextMapPropagator(propagation.TraceContext{}, propagation.Baggage{})
	propagator.Inject(ctx, propagation.HeaderCarrier(tracingHeader))

	if traceParent, ok := tracingHeader["Traceparent"]; ok && len(traceParent) > 0 {
		headers["traceID"] = traceParent[0]
	}

	// TODO: Re-implement correlation ID
	// Propagate correlation ID if present in context
	// if correlationID := ...FromContext(ctx); correlationID != "" {
	// 	headers["correlationID"] = correlationID.String()
	// }

	for key, value := range headers {
		messageKafka.Headers = append(messageKafka.Headers, kafka.Header{
			Key:   key,
			Value: []byte(value),
		})
	}

	if err := k.client.WriteMessages(ctx, messageKafka); err != nil {
		return err
	}
	return nil
}
