package main

// import (
// 	"context"
// 	"errors"
// 	"log"
// 	"os"
// 	"os/signal"
// 	"syscall"
// 	"time"

// 	"github.com/jailtonjunior94/outbox/configs"
// 	kafkaConsumer "github.com/jailtonjunior94/outbox/pkg/messaging/kafka"
// 	"github.com/jailtonjunior94/outbox/pkg/o11y"

// 	"github.com/cenkalti/backoff/v4"
// )

// func main() {
// 	config, err := configs.LoadConfig(".")
// 	if err != nil {
// 		log.Fatal(err)
// 	}

// 	ctx, cancel := context.WithCancel(context.Background())
// 	defer cancel()

// 	sigChan := make(chan os.Signal, 1)
// 	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

// 	go func() {
// 		sig := <-sigChan
// 		log.Printf("Received signal: %s. Shutting down gracefully...", sig)
// 		cancel()
// 	}()

// 	observability := o11y.NewObservability(
// 		o11y.WithServiceName(config.ServiceName),
// 		o11y.WithServiceVersion("1.0.0"),
// 		o11y.WithResource(),
// 		o11y.WithLoggerProvider(ctx, config.OtelExporterEndpoint),
// 		o11y.WithTracerProvider(ctx, config.OtelExporterEndpoint),
// 		o11y.WithMeterProvider(ctx, config.OtelExporterEndpoint),
// 	)

// 	backoff := backoff.NewExponentialBackOff()
// 	backoff.MaxElapsedTime = time.Second * 1

// 	consumer := kafkaConsumer.NewConsumer(
// 		observability,
// 		kafkaConsumer.WithBrokers(config.KafkaBrokers),
// 		kafkaConsumer.WithGroupID("meu-grupo-consumidor-1"),
// 		kafkaConsumer.WithTopic(config.KafkaFinacialTopics[0]),
// 		kafkaConsumer.WithMaxRetries(3),
// 		kafkaConsumer.WithRetryChan(1000),
// 		kafkaConsumer.WithBackoff(backoff),
// 		kafkaConsumer.WithReader(),
// 		kafkaConsumer.WithHandler(handlerMessage),
// 	)

// 	go func() {
// 		if err := consumer.Consume(ctx, handlerMessage); err != nil {
// 			log.Printf("Error consuming messages: %v", err)
// 			cancel()
// 		}
// 	}()

// 	defer func() {
// 		if r := recover(); r != nil {
// 			log.Printf("Recovered from panic: %v", r)
// 		}
// 	}()

// 	<-ctx.Done()
// 	log.Println("Consumer has been shut down.")
// }

// func handlerMessage(ctx context.Context, body []byte) error {
// 	log.Println("Message received: ", string(body))
// 	if string(body) == "error" {
// 		return errors.New("deu ruim")
// 	}
// 	return nil
// }
