package worker

import (
	"context"
	"log"

	"github.com/jailtonjunior94/order/internal/order"
	"github.com/jailtonjunior94/order/pkg/bundle"
	"github.com/jailtonjunior94/order/pkg/o11y"

	"github.com/robfig/cron/v3"
)

type worker struct {
}

func NewWorkers() *worker {
	return &worker{}
}

func (w *worker) Run() {
	ctx := context.Background()
	ioc := bundle.NewContainer(ctx)

	/* Observability */
	resource, err := o11y.NewServiceResource(ctx, ioc.Config.O11yConfig.OrderWorker, ioc.Config.O11yConfig.ServiceVersion, ioc.Config.Environment)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	metrics, shutdown, err := o11y.NewMetrics(ctx, ioc.Config.O11yConfig.ExporterEndpoint, ioc.Config.O11yConfig.OrderWorker, resource)
	if err != nil {
		log.Fatalf("failed to create metrics: %v", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown metrics: %v", err)
		}
	}()

	tracer, shutdown, err := o11y.NewTracer(ctx, ioc.Config.O11yConfig.ExporterEndpoint, ioc.Config.O11yConfig.OrderWorker, resource)
	if err != nil {
		log.Fatalf("failed to create tracer: %v", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown tracer: %v", err)
		}
	}()

	logger, shutdown, err := o11y.NewLogger(ctx, tracer, ioc.Config.O11yConfig.ExporterEndpointHTTP, ioc.Config.O11yConfig.OrderWorker, resource)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown logger: %v", err)
		}
	}()

	telemetry, err := o11y.NewTelemetry(tracer, metrics, logger)
	if err != nil {
		log.Fatalf("failed to create telemetry: %v", err)
	}

	/* Close DBConnection */
	defer func() {
		if err := ioc.DB.Close(); err != nil {
			log.Fatalf("failed to close database: %v", err)
		}
	}()

	/* Order */
	publishEventHandler := order.RegisterPublishEventHandler(ioc, telemetry)

	jobs := cron.New()

	_, err = jobs.AddFunc(ioc.Config.WorkerConfig.CronExpression, publishEventHandler.Handle)
	if err != nil {
		log.Fatalf("failed to add cron job: %v", err)
	}

	jobs.Run()
}
