package configs

import (
	"strings"

	"github.com/spf13/viper"
)

type (
	Config struct {
		DBConfig     DBConfig     `mapstructure:",squash"`
		HTTPConfig   HTTPConfig   `mapstructure:",squash"`
		O11yConfig   O11yConfig   `mapstructure:",squash"`
		KafkaConfig  KafkaConfig  `mapstructure:",squash"`
		WorkerConfig WorkerConfig `mapstructure:",squash"`
	}

	DBConfig struct {
		Driver         string `mapstructure:"DB_DRIVER"`
		Host           string `mapstructure:"DB_HOST"`
		Port           string `mapstructure:"DB_PORT"`
		User           string `mapstructure:"DB_USER"`
		Password       string `mapstructure:"DB_PASSWORD"`
		Name           string `mapstructure:"DB_NAME"`
		DBMaxIdleConns int    `mapstructure:"DB_MAX_IDLE_CONNS"`
		MigratePath    string `mapstructure:"MIGRATE_PATH"`
	}

	HTTPConfig struct {
		Port string `mapstructure:"HTTP_PORT"`
	}

	O11yConfig struct {
		ServiceName      string `mapstructure:"OTEL_SERVICE_NAME"`
		ServiceVersion   string `mapstructure:"OTEL_SERVICE_VERSION"`
		ExporterEndpoint string `mapstructure:"OTEL_EXPORTER_OTLP_ENDPOINT"`
	}

	KafkaConfig struct {
		Brokers                []string `mapstructure:"KAFKA_BROKERS"`
		Order                  string   `mapstructure:"KAFKA_ORDER_TOPIC"`
		OrderPartitions        int      `mapstructure:"KAFKA_ORDER_NUM_PARTITIONS"`
		OrderReplicationFactor int      `mapstructure:"KAFKA_ORDER_REPLICATION_FACTOR"`
		OrderDLQ               string   `mapstructure:"KAFKA_ORDER_DLQ_TOPIC"`
		OrderGroupID           string   `mapstructure:"KAFKA_ORDER_GROUP_ID"`
	}

	WorkerConfig struct {
		CronExpression string `mapstructure:"WORKER_CRON"`
	}
)

func LoadConfig(path string) (*Config, error) {
	var config *Config

	viper.AddConfigPath(path)
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	err := viper.ReadInConfig()
	if err != nil {
		return nil, err
	}

	err = viper.Unmarshal(&config)
	if err != nil {
		return nil, err
	}

	if brokers := viper.GetString("KAFKA_BROKERS"); brokers != "" {
		config.KafkaConfig.Brokers = strings.Split(brokers, ",")
	}

	return config, nil
}
