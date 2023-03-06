package trace

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	traceSdk "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
	"go.opentelemetry.io/otel/trace"
)

var (
	ErrMissingApplicationName = errors.New("trace: misssing application name")
)

type TraceClient struct {
	applicationName    string
	applicationVersion string
	exporter           traceSdk.SpanExporter

	// TraceRatio indicates how often the system should collect traces.
	// Use it with caution: It may overload the system and also be too expensive to mantain its value too high in a high throughput system
	// Values vary between 0 and 1, with 0 meaning No Sampling and 1 meaning Always Sampling.
	// Values lower than 0 are treated as 0 and values greater than 1 are treated as 1.
	traceRatio float64
}

func NewTraceClient(
	applicationName string,
	applicationVersion string,
	exporter traceSdk.SpanExporter,
	traceRatio float64,
) (*TraceClient, error) {
	if applicationName == "" {
		return nil, ErrMissingApplicationName
	}

	if applicationVersion == "" {
		applicationVersion = "unknown_version"
	}

	return &TraceClient{
		applicationName:    applicationName,
		applicationVersion: applicationVersion,
		exporter:           exporter,
		traceRatio:         traceRatio,
	}, nil
}

func MustNewTraceClient(
	applicationName string,
	applicationVersion string,
	exporter traceSdk.SpanExporter,
	traceRatio float64,
) *TraceClient {
	client, err := NewTraceClient(
		applicationName,
		applicationVersion,
		exporter,
		traceRatio,
	)
	if err != nil {
		panic(err)
	}

	return client
}

func (c TraceClient) Tracer(ctx context.Context) (trace.Tracer, func(context.Context) error) {
	tOpts := []traceSdk.TracerProviderOption{
		traceSdk.WithSampler(traceSdk.TraceIDRatioBased(c.traceRatio)),
		traceSdk.WithBatcher(c.exporter),
		traceSdk.WithResource(
			resource.NewWithAttributes(
				semconv.SchemaURL,
				semconv.ServiceNameKey.String(c.applicationName),
				semconv.ServiceVersionKey.String(c.applicationVersion),
			),
		),
	}

	traceProvider := traceSdk.NewTracerProvider(tOpts...)

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return traceProvider.Tracer(c.applicationName), traceProvider.ForceFlush
}
