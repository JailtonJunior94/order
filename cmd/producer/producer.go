package main

import (
	"context"
	"log"

	"github.com/jailtonjunior94/outbox/configs"
	kafkaBuilder "github.com/jailtonjunior94/outbox/pkg/messaging/kafka"

	"github.com/segmentio/kafka-go"
)

func main() {
	ctx := context.Background()
	config, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	createTopics(ctx, config)
	produceMessages(config.KafkaBrokers[0], config.KafkaFinacialTopics[0])
}

func createTopics(ctx context.Context, config *configs.Config) {
	conn, err := kafka.DialContext(ctx, "tcp", config.KafkaBrokers[0])
	if err != nil {
		log.Fatal("failed to dial leader:", err)
	}
	defer conn.Close()

	topics := []*kafkaBuilder.TopicConfig{}
	for _, topic := range config.KafkaFinacialTopics {
		topics = append(topics, kafkaBuilder.NewTopicConfig(topic, 1, 1))
	}

	_ = kafkaBuilder.NewKafkaBuilder(conn).
		DeclareTopics(topics...).
		Build()
}

func produceMessages(kafkaURL, topic string) {
	writer := kafka.NewWriter(kafka.WriterConfig{
		Brokers: []string{kafkaURL},
		Topic:   topic,
	})
	defer writer.Close()

	messages := []kafka.Message{
		{Value: []byte("Welcome")},
		{Value: []byte("to")},
		{Value: []byte("Kafka")},
		{Value: []byte("with")},
		{Value: []byte("Go")},
	}

	err := writer.WriteMessages(context.Background(), messages...)
	if err != nil {
		log.Fatal("failed to write messages:", err)
	}
	log.Println("messages written successfully")
}
