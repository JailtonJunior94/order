package consumer

import (
	"context"
	"errors"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jailtonjunior94/order/configs"
	"github.com/jailtonjunior94/order/pkg/bundle"
	kafkaConsumer "github.com/jailtonjunior94/order/pkg/messaging/kafka"
	"github.com/jailtonjunior94/order/pkg/observability"
	"github.com/jailtonjunior94/order/pkg/observability/otel"

	"github.com/cenkalti/backoff/v4"
	"github.com/segmentio/kafka-go"
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
	// Parse configuration with defaults
	logLevel := observability.LogLevelInfo
	if ioc.Config.O11yConfig.LogLevel != "" {
		logLevel = observability.LogLevel(ioc.Config.O11yConfig.LogLevel)
	}

	logFormat := observability.LogFormatJSON
	if ioc.Config.O11yConfig.LogFormat == "text" {
		logFormat = observability.LogFormatText
	}

	sampleRate := 1.0
	if ioc.Config.O11yConfig.TraceSampleRate != "" {
		if rate, err := strconv.ParseFloat(ioc.Config.O11yConfig.TraceSampleRate, 64); err == nil {
			sampleRate = rate
		}
	}

	// Create observability provider
	config := &otel.Config{
		ServiceName:     ioc.Config.O11yConfig.OrderConsumer,
		ServiceVersion:  ioc.Config.O11yConfig.ServiceVersion,
		Environment:     ioc.Config.Environment,
		OTLPEndpoint:    ioc.Config.O11yConfig.ExporterEndpoint,
		OTLPProtocol:    otel.OTLPProtocol(ioc.Config.O11yConfig.Protocol),
		TraceSampleRate: sampleRate,
		LogLevel:        logLevel,
		LogFormat:       logFormat,
	}

	o11y, err := otel.NewProvider(ctx, config)
	if err != nil {
		log.Fatalf("failed to initialize observability: %v", err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := o11y.Shutdown(shutdownCtx); err != nil {
			log.Printf("observability shutdown error: %v", err)
		}
	}()

	/* Close DBConnection */
	defer func() {
		if err := ioc.DB.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	backoff := backoff.NewExponentialBackOff()
	backoff.MaxElapsedTime = time.Second * 1

	c.declareTopics(ioc.Config)

	consumer := kafkaConsumer.NewConsumer(
	 o11y,
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
	defer func() {
		if err := conn.Close(); err != nil {
			log.Printf("failed to close kafka connection: %v", err)
		}
	}()

	if err := kafkaConsumer.NewKafkaBuilder(conn).DeclareTopics(
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
	).Build(); err != nil {
		panic(fmt.Sprintf("failed to declare topics: %v", err))
	}
}

func handlerMessage(ctx context.Context, body []byte) error {
	log.Println("Message received: ", string(body))
	if string(body) == "error" {
		return errors.New("deu ruim")
	}
	return nil
}
