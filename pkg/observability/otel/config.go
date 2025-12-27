package otel

import (
	"context"
	"fmt"

	"github.com/jailtonjunior94/order/pkg/observability"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.21.0"
)

// Config holds the configuration for the OpenTelemetry provider.
type Config struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	OTLPEndpoint   string

	// Trace configuration
	TraceSampleRate float64 // 0.0 to 1.0, default 1.0 (always sample)

	// Log configuration
	LogLevel  observability.LogLevel
	LogFormat observability.LogFormat

	// Resource attributes (optional)
	ResourceAttributes map[string]string
}

// DefaultConfig returns a configuration with sensible defaults.
func DefaultConfig(serviceName string) *Config {
	return &Config{
		ServiceName:     serviceName,
		ServiceVersion:  "unknown",
		Environment:     "development",
		OTLPEndpoint:    "localhost:4317",
		TraceSampleRate: 1.0,
		LogLevel:        observability.LogLevelInfo,
		LogFormat:       observability.LogFormatJSON,
	}
}

// Provider implements the observability.Observability interface using OpenTelemetry.
type Provider struct {
	config          *Config
	tracerProvider  *sdktrace.TracerProvider
	meterProvider   *sdkmetric.MeterProvider
	tracer          *otelTracer
	logger          *otelLogger
	metrics         *otelMetrics
	shutdownFuncs   []func(context.Context) error
}

// NewProvider creates and initializes a new OpenTelemetry provider.
func NewProvider(ctx context.Context, config *Config) (*Provider, error) {
	if config == nil {
		return nil, fmt.Errorf("config cannot be nil")
	}

	provider := &Provider{
		config:        config,
		shutdownFuncs: make([]func(context.Context) error, 0),
	}

	// Create resource with service information
	res, err := provider.createResource()
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Initialize tracer provider
	if err := provider.initTracerProvider(ctx, res); err != nil {
		return nil, fmt.Errorf("failed to initialize tracer provider: %w", err)
	}

	// Initialize meter provider
	if err := provider.initMeterProvider(ctx, res); err != nil {
		return nil, fmt.Errorf("failed to initialize meter provider: %w", err)
	}

	// Set global propagator for trace context propagation
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Initialize components
	provider.tracer = newOtelTracer(provider.tracerProvider.Tracer(config.ServiceName))
	provider.logger = newOtelLogger(config.LogLevel, config.LogFormat, config.ServiceName)
	provider.metrics = newOtelMetrics(provider.meterProvider.Meter(config.ServiceName))

	return provider, nil
}

// createResource creates an OTLP resource with service information.
func (p *Provider) createResource() (*resource.Resource, error) {
	attrs := []resource.Option{
		resource.WithAttributes(
			semconv.ServiceName(p.config.ServiceName),
			semconv.ServiceVersion(p.config.ServiceVersion),
			semconv.DeploymentEnvironment(p.config.Environment),
		),
	}

	// Add custom resource attributes
	if len(p.config.ResourceAttributes) > 0 {
		customAttrs := make([]attribute.KeyValue, 0, len(p.config.ResourceAttributes))
		for k, v := range p.config.ResourceAttributes {
			customAttrs = append(customAttrs, attribute.String(k, v))
		}
		attrs = append(attrs, resource.WithAttributes(customAttrs...))
	}

	return resource.New(
		context.Background(),
		attrs...,
	)
}

// initTracerProvider initializes the OpenTelemetry tracer provider.
func (p *Provider) initTracerProvider(ctx context.Context, res *resource.Resource) error {
	// Create OTLP exporter
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(p.config.OTLPEndpoint),
		otlptracegrpc.WithInsecure(), // Use WithTLSCredentials in production
	)
	if err != nil {
		return fmt.Errorf("failed to create trace exporter: %w", err)
	}

	// Create sampler based on sample rate
	var sampler sdktrace.Sampler
	if p.config.TraceSampleRate >= 1.0 {
		sampler = sdktrace.AlwaysSample()
	} else if p.config.TraceSampleRate <= 0.0 {
		sampler = sdktrace.NeverSample()
	} else {
		sampler = sdktrace.TraceIDRatioBased(p.config.TraceSampleRate)
	}

	// Create tracer provider with batch span processor for performance
	p.tracerProvider = sdktrace.NewTracerProvider(
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sampler),
		sdktrace.WithBatcher(exporter), // Batch for high throughput
	)

	// Set as global tracer provider
	otel.SetTracerProvider(p.tracerProvider)

	// Register shutdown function
	p.shutdownFuncs = append(p.shutdownFuncs, p.tracerProvider.Shutdown)

	return nil
}

// initMeterProvider initializes the OpenTelemetry meter provider.
func (p *Provider) initMeterProvider(ctx context.Context, res *resource.Resource) error {
	// Create OTLP exporter
	exporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithEndpoint(p.config.OTLPEndpoint),
		otlpmetricgrpc.WithInsecure(), // Use WithTLSCredentials in production
	)
	if err != nil {
		return fmt.Errorf("failed to create metrics exporter: %w", err)
	}

	// Create meter provider with periodic reader for performance
	p.meterProvider = sdkmetric.NewMeterProvider(
		sdkmetric.WithResource(res),
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter)),
	)

	// Set as global meter provider
	otel.SetMeterProvider(p.meterProvider)

	// Register shutdown function
	p.shutdownFuncs = append(p.shutdownFuncs, p.meterProvider.Shutdown)

	return nil
}

// Tracer returns the OpenTelemetry tracer.
func (p *Provider) Tracer() observability.Tracer {
	return p.tracer
}

// Logger returns the OpenTelemetry logger.
func (p *Provider) Logger() observability.Logger {
	return p.logger
}

// Metrics returns the OpenTelemetry metrics recorder.
func (p *Provider) Metrics() observability.Metrics {
	return p.metrics
}

// Shutdown gracefully shuts down the provider, flushing any pending telemetry.
func (p *Provider) Shutdown(ctx context.Context) error {
	var errs []error
	for _, shutdown := range p.shutdownFuncs {
		if err := shutdown(ctx); err != nil {
			errs = append(errs, err)
		}
	}

	if len(errs) > 0 {
		return fmt.Errorf("errors during shutdown: %v", errs)
	}

	return nil
}
