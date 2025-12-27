package worker

import (
	"context"
	"log"
	"strconv"
	"time"

	"github.com/jailtonjunior94/order/internal/order"
	"github.com/jailtonjunior94/order/pkg/bundle"
	"github.com/jailtonjunior94/order/pkg/observability"
	"github.com/jailtonjunior94/order/pkg/observability/otel"

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
		ServiceName:     ioc.Config.O11yConfig.OrderWorker,
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
			log.Fatalf("failed to close database: %v", err)
		}
	}()

	/* Order */
	publishEventHandler := order.RegisterPublishEventHandler(ioc, o11y)

	jobs := cron.New()

	if _, err = jobs.AddFunc(ioc.Config.WorkerConfig.CronExpression, publishEventHandler.Handle); err != nil {
		log.Printf("failed to add cron job: %v", err)
		return
	}

	jobs.Run()
}
