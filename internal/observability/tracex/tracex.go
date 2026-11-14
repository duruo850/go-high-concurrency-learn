package tracex

import (
	"context"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/TripleCGame/go-high-concurrency-learn/internal/conf"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

const serviceName = "go-high-concurrency-learn"

var (
	enabled       bool
	tp            *sdktrace.TracerProvider
	activeEndpoint string
)

// Init configures OTel TracerProvider + OTLP/HTTP exporter.
// Empty endpoint = disabled (noop). Env OTEL_EXPORTER_OTLP_ENDPOINT overrides config.
func Init(c *conf.Trace) (bool, error) {
	endpoint := ""
	sample := 1.0
	if c != nil {
		endpoint = strings.TrimSpace(c.Endpoint)
		if c.SampleRatio > 0 && c.SampleRatio <= 1 {
			sample = c.SampleRatio
		}
	}
	if v := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT"); v != "" {
		endpoint = strings.TrimSpace(v)
	}
	if endpoint == "" {
		return false, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	exp, err := otlptracehttp.New(ctx,
		otlptracehttp.WithEndpointURL(normalizeEndpoint(endpoint)),
		otlptracehttp.WithInsecure(),
	)
	if err != nil {
		return false, fmt.Errorf("otlp exporter: %w", err)
	}

	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion("1.0.0"),
		),
	)
	if err != nil {
		return false, fmt.Errorf("otel resource: %w", err)
	}

	tp = sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exp),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(sdktrace.TraceIDRatioBased(sample))),
	)
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))
	activeEndpoint = endpoint
	enabled = true
	return true, nil
}

// Endpoint returns the effective OTLP endpoint (empty if disabled).
func Endpoint() string { return activeEndpoint }

func normalizeEndpoint(ep string) string {
	ep = strings.TrimRight(ep, "/")
	if strings.HasSuffix(ep, "/v1/traces") {
		return ep
	}
	return ep + "/v1/traces"
}

func Enabled() bool { return enabled }

func Tracer() trace.Tracer {
	return otel.Tracer(serviceName)
}

func Shutdown(ctx context.Context) error {
	if tp == nil {
		return nil
	}
	err := tp.Shutdown(ctx)
	tp = nil
	enabled = false
	activeEndpoint = ""
	return err
}

// TraceID returns hex trace id from ctx, or empty.
func TraceID(ctx context.Context) string {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return ""
	}
	return sc.TraceID().String()
}

// Start is a thin wrapper for demo / handlers.
func Start(ctx context.Context, name string, attrs ...attribute.KeyValue) (context.Context, trace.Span) {
	return Tracer().Start(ctx, name, trace.WithAttributes(attrs...))
}
