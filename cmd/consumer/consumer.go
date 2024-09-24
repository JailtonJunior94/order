package main

import (
	"context"
	"errors"
	"log"
	"time"

	"github.com/cenkalti/backoff/v4"
	"github.com/jailtonjunior94/outbox/configs"
	kafkaConsumer "github.com/jailtonjunior94/outbox/pkg/messaging/kafka"
)

func main() {
	ctx := context.Background()
	config, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	backoff := backoff.NewExponentialBackOff()
	backoff.MaxElapsedTime = time.Second * 1

	consumer := kafkaConsumer.NewConsumer(
		kafkaConsumer.WithBrokers(config.KafkaBrokers),
		kafkaConsumer.WithGroupID("meu-grupo-consumer-1"),
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
