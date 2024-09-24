package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/jailtonjunior94/outbox/configs"
	kafkaConsumer "github.com/jailtonjunior94/outbox/pkg/messaging/kafka"
	"github.com/jailtonjunior94/outbox/pkg/o11y"

	"github.com/cenkalti/backoff/v4"
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

	backoff := backoff.NewExponentialBackOff()
	backoff.MaxElapsedTime = time.Second * 1

	consumer := kafkaConsumer.NewConsumer(
		observability,
		kafkaConsumer.WithBrokers(config.KafkaBrokers),
		kafkaConsumer.WithGroupID("meu-grupo-consumidor-1"),
		kafkaConsumer.WithTopic(config.KafkaFinacialTopics[0]),
		kafkaConsumer.WithMaxRetries(3),
		kafkaConsumer.WithRetryChan(1000),
		kafkaConsumer.WithBackoff(backoff),
		kafkaConsumer.WithReader(),
		kafkaConsumer.WithHandler(handlerMessage),
	)
	consumer.Consume(ctx, handlerMessage)

	forever := make(chan bool)
	<-forever
}

func handlerMessage(ctx context.Context, body []byte) error {
	log.Println("Message received: ", string(body))
	if string(body) == "error" {
		return errors.New("deu ruim")
	}
	return nil
}
