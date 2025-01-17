package utils

import (
	"fmt"

	"github.com/sirupsen/logrus"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"

	"context"

	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

var (
	OTLP_COLLECTOR_HOST=GetEnv("OTLP_COLLECTOR_HOST","localhost")
	OTLP_COLLECTOR_PORT=GetEnv("OTLP_COLLECTOR_PORT","4318")
)
// Initialize tracing and return a cleanup function
func InitTracer(serviceName string, logger *logrus.Logger) func() {
	// Create OTLP exporter to send trace data to the collector
	exporter, err := otlptracehttp.New(context.Background(),
		otlptracehttp.WithEndpoint(fmt.Sprintf("%s:%s",OTLP_COLLECTOR_HOST,OTLP_COLLECTOR_PORT)), // Use the HTTP endpoint of the collector
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		logger.Fatalf("Failed to create OTLP exporter: %v", err)
	}

	tp := trace.NewTracerProvider(
		trace.WithBatcher(exporter),
		trace.WithResource(resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		)),
	)

	otel.SetTracerProvider(tp)
	
	otel.SetTextMapPropagator(
        propagation.NewCompositeTextMapPropagator(
            propagation.TraceContext{},
            propagation.Baggage{},
        ),
    )

	return func() {
		_ = tp.Shutdown(context.Background())
	}
}
