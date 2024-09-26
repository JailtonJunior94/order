package consumer

import (
	"context"
	"errors"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jailtonjunior94/order/configs"
	"github.com/jailtonjunior94/order/pkg/bundle"
	kafkaConsumer "github.com/jailtonjunior94/order/pkg/messaging/kafka"
	"github.com/segmentio/kafka-go"

	"github.com/cenkalti/backoff/v4"
)

type consumer struct {
}

func NewConsumer() *consumer {
	return &consumer{}
}

func (c *consumer) Run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	go func() {
		sig := <-sigChan
		log.Printf("Received signal: %s. Shutting down gracefully...", sig)
		cancel()
	}()
	ioc := bundle.NewContainer(ctx)

	/* Observability */
	tracerProvider := ioc.Observability.TracerProvider()
	defer func() {
		if err := tracerProvider.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	meterProvider := ioc.Observability.MeterProvider()
	defer func() {
		if err := meterProvider.Shutdown(ctx); err != nil {
			log.Fatal(err)
		}
	}()

	/* Close DBConnection */
	defer func() {
		if err := ioc.DB.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	backoff := backoff.NewExponentialBackOff()
	backoff.MaxElapsedTime = time.Second * 1

	c.declareTopics(ioc.Config)

	consumer := kafkaConsumer.NewConsumer(
		ioc.Observability,
		kafkaConsumer.WithBrokers(ioc.Config.KafkaConfig.Brokers),
		kafkaConsumer.WithGroupID(ioc.Config.KafkaConfig.OrderGroupID),
		kafkaConsumer.WithTopic(ioc.Config.KafkaConfig.Order),
		kafkaConsumer.WithMaxRetries(3),
		kafkaConsumer.WithRetryChan(1000),
		kafkaConsumer.WithBackoff(backoff),
		kafkaConsumer.WithReader(),
		kafkaConsumer.WithHandler(handlerMessage),
	)

	go func() {
		if err := consumer.Consume(ctx, handlerMessage); err != nil {
			log.Printf("Error consuming messages: %v", err)
			cancel()
		}
	}()

	defer func() {
		if r := recover(); r != nil {
			log.Printf("Recovered from panic: %v", r)
		}
	}()

	<-ctx.Done()
	log.Println("Consumer has been shut down.")
}

func (c *consumer) declareTopics(config *configs.Config) {
	conn, err := kafka.Dial("tcp", config.KafkaConfig.Brokers[0])
	if err != nil {
		panic(err.Error())
	}
	defer conn.Close()

	kafkaConsumer.NewKafkaBuilder(conn).DeclareTopics(
		kafkaConsumer.NewTopicConfig(
			config.KafkaConfig.Order,
			config.KafkaConfig.OrderPartitions,
			config.KafkaConfig.OrderReplicationFactor,
		),
		kafkaConsumer.NewTopicConfig(
			config.KafkaConfig.OrderDLQ,
			config.KafkaConfig.OrderPartitions,
			config.KafkaConfig.OrderReplicationFactor,
		),
	).Build()
}

func handlerMessage(ctx context.Context, body []byte) error {
	log.Println("Message received: ", string(body))
	if string(body) == "error" {
		return errors.New("deu ruim")
	}
	return nil
}
