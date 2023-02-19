package trace

import (
	"context"

	otelOtlpGrpcExporter "go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/trace"
	gGRPC "google.golang.org/grpc"
)

type OtlpTracerProvider struct {
	applicationName             string
	applicationVersion          string
	traceRatio                  float64
	otelCollectorGprcConnection *gGRPC.ClientConn
}

func NewOtlpTracerProvider(
	applicationName string,
	applicationVersion string,
	traceRatio float64,
	otelCollectorGprcConnection *gGRPC.ClientConn,
) (*OtlpTracerProvider, error) {
	if applicationName == "" {
		return nil, MissingApplicationNameErr
	}

	if applicationVersion == "" {
		applicationVersion = "Unknown"
	}

	return &OtlpTracerProvider{
		applicationName:             applicationName,
		applicationVersion:          applicationVersion,
		traceRatio:                  traceRatio,
		otelCollectorGprcConnection: otelCollectorGprcConnection,
	}, nil
}

func MustNewOtlpProvider(
	applicationName string,
	applicationVersion string,
	traceRatio float64,
	otelCollectorGprcConnection *gGRPC.ClientConn,
) *OtlpTracerProvider {
	otlpProvider, err := NewOtlpTracerProvider(applicationName, applicationVersion, traceRatio, otelCollectorGprcConnection)
	if err != nil {
		panic(err)
	}

	return otlpProvider
}

func (c OtlpTracerProvider) Tracer(ctx context.Context) (trace.Tracer, func(context.Context) error, error) {
	otlpTraceExporter, err := otelOtlpGrpcExporter.New(
		ctx,
		otelOtlpGrpcExporter.WithGRPCConn(c.otelCollectorGprcConnection),
	)
	if err != nil {
		return nil, nil, err
	}

	traceClient, err := NewTraceClient(
		c.applicationName,
		c.applicationVersion,
		otlpTraceExporter,
		c.traceRatio,
	)
	if err != nil {
		return nil, nil, err
	}

	tracer, flush := traceClient.Tracer(ctx)

	return tracer, flush, nil
}

func (c OtlpTracerProvider) MustTracer(ctx context.Context) (trace.Tracer, func(context.Context) error) {
	tracer, flush, err := c.Tracer(ctx)
	if err != nil {
		panic(err)
	}

	return tracer, flush
}
