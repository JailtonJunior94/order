package otel

import (
	"context"
	"io"
	"log/slog"
	"os"

	"github.com/jailtonjunior94/order/pkg/observability"
	"go.opentelemetry.io/otel/trace"
)

// otelLogger implements observability.Logger using slog with OpenTelemetry trace context.
type otelLogger struct {
	logger      *slog.Logger
	level       observability.LogLevel
	format      observability.LogFormat
	serviceName string
	fields      []observability.Field
}

// newOtelLogger creates a new logger with the specified level and format.
func newOtelLogger(level observability.LogLevel, format observability.LogFormat, serviceName string) *otelLogger {
	return &otelLogger{
		logger:      createSlogLogger(level, format, os.Stdout),
		level:       level,
		format:      format,
		serviceName: serviceName,
		fields:      make([]observability.Field, 0),
	}
}

// createSlogLogger creates a slog logger with the specified configuration.
func createSlogLogger(level observability.LogLevel, format observability.LogFormat, output io.Writer) *slog.Logger {
	slogLevel := convertLogLevel(level)

	opts := &slog.HandlerOptions{
		Level: slogLevel,
	}

	var handler slog.Handler
	if format == observability.LogFormatJSON {
		handler = slog.NewJSONHandler(output, opts)
	} else {
		handler = slog.NewTextHandler(output, opts)
	}

	return slog.New(handler)
}

// convertLogLevel converts observability.LogLevel to slog.Level.
func convertLogLevel(level observability.LogLevel) slog.Level {
	switch level {
	case observability.LogLevelDebug:
		return slog.LevelDebug
	case observability.LogLevelInfo:
		return slog.LevelInfo
	case observability.LogLevelWarn:
		return slog.LevelWarn
	case observability.LogLevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

// Debug logs a debug-level message.
func (l *otelLogger) Debug(ctx context.Context, msg string, fields ...observability.Field) {
	l.log(ctx, slog.LevelDebug, msg, fields...)
}

// Info logs an info-level message.
func (l *otelLogger) Info(ctx context.Context, msg string, fields ...observability.Field) {
	l.log(ctx, slog.LevelInfo, msg, fields...)
}

// Warn logs a warning-level message.
func (l *otelLogger) Warn(ctx context.Context, msg string, fields ...observability.Field) {
	l.log(ctx, slog.LevelWarn, msg, fields...)
}

// Error logs an error-level message.
func (l *otelLogger) Error(ctx context.Context, msg string, fields ...observability.Field) {
	l.log(ctx, slog.LevelError, msg, fields...)
}

// log is the internal logging method that adds trace context and structured fields.
func (l *otelLogger) log(ctx context.Context, level slog.Level, msg string, fields ...observability.Field) {
	// Combine permanent fields with call-specific fields
	allFields := make([]observability.Field, 0, len(l.fields)+len(fields)+3)
	allFields = append(allFields, l.fields...)
	allFields = append(allFields, fields...)

	// Extract trace context from the context
	span := trace.SpanFromContext(ctx)
	if span.SpanContext().IsValid() {
		allFields = append(allFields,
			observability.String("trace_id", span.SpanContext().TraceID().String()),
			observability.String("span_id", span.SpanContext().SpanID().String()),
		)
	}

	// Add service name
	allFields = append(allFields, observability.String("service", l.serviceName))

	// Convert fields to slog.Attr
	attrs := make([]slog.Attr, 0, len(allFields))
	for _, field := range allFields {
		attrs = append(attrs, convertFieldToSlogAttr(field))
	}

	// Log with slog
	l.logger.LogAttrs(ctx, level, msg, attrs...)
}

// With creates a child logger with additional fields.
func (l *otelLogger) With(fields ...observability.Field) observability.Logger {
	return &otelLogger{
		logger:      l.logger,
		level:       l.level,
		format:      l.format,
		serviceName: l.serviceName,
		fields:      append(l.fields, fields...),
	}
}

// convertFieldToSlogAttr converts an observability.Field to a slog.Attr.
func convertFieldToSlogAttr(field observability.Field) slog.Attr {
	switch v := field.Value.(type) {
	case string:
		return slog.String(field.Key, v)
	case int:
		return slog.Int(field.Key, v)
	case int64:
		return slog.Int64(field.Key, v)
	case float64:
		return slog.Float64(field.Key, v)
	case bool:
		return slog.Bool(field.Key, v)
	case error:
		return slog.String(field.Key, v.Error())
	default:
		return slog.Any(field.Key, v)
	}
}
