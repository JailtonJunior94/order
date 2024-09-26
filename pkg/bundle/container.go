package bundle

import (
	"context"
	"database/sql"

	"github.com/jailtonjunior94/order/configs"
	"github.com/jailtonjunior94/order/pkg/database/postgres"
	"github.com/jailtonjunior94/order/pkg/o11y"
)

type Container struct {
	DB            *sql.DB
	Config        *configs.Config
	Observability o11y.Observability
}

func NewContainer(ctx context.Context) *Container {
	config, err := configs.LoadConfig(".")
	if err != nil {
		panic(err)
	}

	db, err := postgres.NewPostgresDatabase(config)
	if err != nil {
		panic(err)
	}

	observability := o11y.NewObservability(
		o11y.WithServiceName(config.O11yConfig.ServiceName),
		o11y.WithServiceVersion(config.O11yConfig.ServiceVersion),
		o11y.WithResource(),
		o11y.WithLoggerProvider(ctx, config.O11yConfig.ExporterEndpoint),
		o11y.WithTracerProvider(ctx, config.O11yConfig.ExporterEndpoint),
		o11y.WithMeterProvider(ctx, config.O11yConfig.ExporterEndpoint),
	)

	return &Container{
		DB:            db,
		Config:        config,
		Observability: observability,
	}
}
