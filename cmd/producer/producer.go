package main

import (
	"context"
	"log"

	"github.com/jailtonjunior94/outbox/configs"
	messagingKafka "github.com/jailtonjunior94/outbox/pkg/messaging/kafka"
	"github.com/jailtonjunior94/outbox/pkg/o11y"

	"github.com/segmentio/kafka-go"
)

func main() {
	ctx := context.Background()
	config, err := configs.LoadConfig(".")
	if err != nil {
		log.Fatal(err)
	}

	observability := o11y.NewObservability(
		o11y.WithServiceName(config.ServiceName),
		o11y.WithServiceVersion("1.0.0"),
		o11y.WithResource(),
		o11y.WithLoggerProvider(ctx, config.OtelExporterEndpoint),
		o11y.WithTracerProvider(ctx, config.OtelExporterEndpoint),
		o11y.WithMeterProvider(ctx, config.OtelExporterEndpoint),
	)

	createTopics(ctx, config)

	producer := newProducer(
		messagingKafka.NewKafkaClient(observability, config.KafkaBrokers[0], config.KafkaFinacialTopics[0]),
	)

	if err := producer.Produce(ctx, config.KafkaFinacialTopics[0], []byte("key"), []byte("value")); err != nil {
		log.Fatal(err)
	}
}

func createTopics(ctx context.Context, config *configs.Config) {
	conn, err := kafka.DialContext(ctx, "tcp", config.KafkaBrokers[0])
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}
	defer conn.Close()

	topics := []*messagingKafka.TopicConfig{}
	for _, topic := range config.KafkaFinacialTopics {
		topics = append(topics, messagingKafka.NewTopicConfig(topic, 1, 1))
	}

	_ = messagingKafka.NewKafkaBuilder(conn).
		DeclareTopics(topics...).
		Build()
}

type producer struct {
	client messagingKafka.KafkaClient
}

func newProducer(client messagingKafka.KafkaClient) *producer {
	return &producer{client: client}
}

func (p *producer) Produce(ctx context.Context, topic string, key, value []byte) error {
	headers := map[string]string{"key": string(key)}
	message := &messagingKafka.Message{
		Key:   key,
		Value: value,
	}

	return p.client.Produce(ctx, topic, headers, message)
}
