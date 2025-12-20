package worker

import (
	"context"
	"log"
	"time"

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
	resource, err := o11y.NewServiceResource(ctx, ioc.Config.O11yConfig.OrderAPI, ioc.Config.O11yConfig.ServiceVersion, ioc.Config.Environment)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	tracer, tracerShutdown, err := o11y.NewTracerWithOptions(
		ctx,
		o11y.WithTracerEndpoint(ioc.Config.O11yConfig.ExporterEndpoint),
		o11y.WithTracerServiceName(ioc.Config.O11yConfig.OrderWorker),
		o11y.WithTracerResource(resource),
		o11y.WithTracerInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to create tracer: %v", err)
	}

	metrics, metricsShutdown, err := o11y.NewMetricsWithOptions(
		ctx,
		o11y.WithMetricsEndpoint(ioc.Config.O11yConfig.ExporterEndpoint),
		o11y.WithMetricsServiceName(ioc.Config.O11yConfig.OrderWorker),
		o11y.WithMetricsResource(resource),
		o11y.WithMetricsInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to create metrics: %v", err)
	}

	logger, loggerShutdown, err := o11y.NewLoggerWithOptions(
		ctx,
		o11y.WithLoggerEndpoint(ioc.Config.O11yConfig.ExporterEndpointHTTP),
		o11y.WithLoggerServiceName(ioc.Config.O11yConfig.OrderWorker),
		o11y.WithLoggerResource(resource),
		o11y.WithLoggerInsecure(),
	)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}

	telemetry, err := o11y.NewTelemetry(tracer, metrics, logger, tracerShutdown, metricsShutdown, loggerShutdown)
	if err != nil {
		log.Fatalf("failed to create telemetry: %v", err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := telemetry.Shutdown(shutdownCtx); err != nil {
			log.Printf("telemetry shutdown error: %v", err)
		}
	}()

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
