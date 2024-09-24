package kafka

import (
	"context"

	"github.com/jailtonjunior94/outbox/pkg/o11y"
	"github.com/segmentio/kafka-go"
)

type (
	KafkaClient interface {
		Produce(ctx context.Context, topic string, headers map[string]string, message *Message) error
	}

	kafkaClient struct {
		client *kafka.Writer
		o11y   o11y.Observability
	}

	Message struct {
		Key   []byte
		Value []byte
	}
)

func NewKafkaClient(
	o11y o11y.Observability,
	broker string,
	topic string,
) KafkaClient {
	client := &kafka.Writer{
		Addr:     kafka.TCP(broker),
		Topic:    topic,
		Balancer: &kafka.LeastBytes{},
	}
	return &kafkaClient{o11y: o11y, client: client}
}

func (k *kafkaClient) Produce(ctx context.Context, topic string, headers map[string]string, message *Message) error {
	ctx, span := k.o11y.Start(ctx, "producer.Produce")
	defer span.End()

	messageKafka := kafka.Message{
		Topic: topic,
		Key:   message.Key,
		Value: message.Value,
	}

	span.AddAttributes(ctx, o11y.Ok, "producer.Produce",
		o11y.Attributes{Key: "messaging.system", Value: "kafka"},
		o11y.Attributes{Key: "messaging.destination", Value: topic},
		o11y.Attributes{Key: "messaging.kafka.message_key", Value: string(messageKafka.Key)},
	)

	for key, value := range headers {
		messageKafka.Headers = append(messageKafka.Headers, kafka.Header{
			Key:   key,
			Value: []byte(value),
		})
	}

	return k.client.WriteMessages(ctx, messageKafka)
}
