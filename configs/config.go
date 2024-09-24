package configs

import (
	"strings"

	"github.com/spf13/viper"
)

type Config struct {
	ServiceName          string   `mapstructure:"SERVICE_NAME"`
	KafkaBrokers         []string `mapstructure:"KAFKA_BROKERS"`
	OtelExporterEndpoint string   `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	KafkaFinacialTopics  []string `mapstructure:"KAFKA_FINACIAL_TOPICS"`
}

func LoadConfig(path string) (*Config, error) {
	var cfg *Config

	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&cfg)
	if err != nil {
		return nil, err
	}

	if brokers := viper.GetString("KAFKA_BROKERS"); brokers != "" {
		cfg.KafkaBrokers = strings.Split(brokers, ",")
	}

	if financialTopics := viper.GetString("KAFKA_FINACIAL_TOPICS"); financialTopics != "" {
		cfg.KafkaFinacialTopics = strings.Split(financialTopics, ",")
	}

	return cfg, nil
}
