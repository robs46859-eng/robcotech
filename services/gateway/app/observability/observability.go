package observability

import (
	"context"
	"log/slog"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// Config holds observability configuration
type Config struct {
	ServiceName string
	Version     string
	LogLevel    string
	OTLPEndpoint string
}

// Observability holds all observability components
type Observability struct {
	Logger   *slog.Logger
	Tracer   trace.Tracer
	TracerProvider *sdktrace.TracerProvider
}

// New creates a new Observability instance
func New(cfg Config) (*Observability, error) {
	// Setup logger
	level := parseLogLevel(cfg.LogLevel)
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{
		Level: level,
	}))

	// Setup tracer - use default tracer provider
	var tracerProvider *sdktrace.TracerProvider
	if cfg.OTLPEndpoint != "" {
		exporter, err := otlptracehttp.New(context.Background())
		if err != nil {
			logger.Error("failed to create OTLP exporter", "error", err)
		} else {
			tracerProvider = sdktrace.NewTracerProvider(
				sdktrace.WithBatcher(exporter),
			)
		}
	}

	if tracerProvider == nil {
		// Use noop tracer provider
		tracerProvider = sdktrace.NewTracerProvider()
	}

	otel.SetTracerProvider(tracerProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	tracer := tracerProvider.Tracer(cfg.ServiceName)

	return &Observability{
		Logger:         logger,
		Tracer:         tracer,
		TracerProvider: tracerProvider,
	}, nil
}

// Shutdown cleans up observability resources
func (o *Observability) Shutdown() error {
	if o.TracerProvider != nil {
		return o.TracerProvider.Shutdown(nil)
	}
	return nil
}

// RecordRequestDuration records the duration of a request
func (o *Observability) RecordRequestDuration(path, method string, statusCode int, duration time.Duration) {
	// TODO: Implement metrics recording with OpenTelemetry metrics
}

// RecordRequestCount records a request count
func (o *Observability) RecordRequestCount(path, method string, statusCode int) {
	// TODO: Implement metrics recording with OpenTelemetry metrics
}

func parseLogLevel(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
