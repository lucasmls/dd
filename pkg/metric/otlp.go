package metric

import (
	"context"
	"errors"

	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/metric"

	gGRPC "google.golang.org/grpc"
)

var (
	ErrMissingApplicationName = errors.New("metric: misssing application name")
)

type OtlpMetricProvider struct {
	applicationName             string
	applicationVersion          string
	otelCollectorGprcConnection *gGRPC.ClientConn
}

func NewOtlpMetricProvider(
	applicationName string,
	applicationVersion string,
	otelCollectorGprcConnection *gGRPC.ClientConn,
) (*OtlpMetricProvider, error) {
	if applicationName == "" {
		return nil, ErrMissingApplicationName
	}

	if applicationVersion == "" {
		applicationVersion = "unknown_version"
	}

	return &OtlpMetricProvider{
		applicationName:             applicationName,
		applicationVersion:          applicationVersion,
		otelCollectorGprcConnection: otelCollectorGprcConnection,
	}, nil
}

func MustNewOtlpMetricProvider(
	applicationName string,
	applicationVersion string,
	otelCollectorGprcConnection *gGRPC.ClientConn,
) *OtlpMetricProvider {
	otlpProvider, err := NewOtlpMetricProvider(
		applicationName,
		applicationVersion,
		otelCollectorGprcConnection,
	)
	if err != nil {
		panic(err)
	}

	return otlpProvider
}

func (p OtlpMetricProvider) Meter(ctx context.Context) (metric.Meter, func(context.Context) error, error) {
	otlpExporter, err := otlpmetricgrpc.New(
		ctx,
		otlpmetricgrpc.WithGRPCConn(p.otelCollectorGprcConnection),
	)
	if err != nil {
		return nil, nil, err
	}

	metricClient, err := NewMetricClient(
		p.applicationName,
		p.applicationVersion,
		otlpExporter,
	)
	if err != nil {
		return nil, nil, err
	}

	meter, flush := metricClient.Meter(ctx)

	return meter, flush, nil
}

func (p OtlpMetricProvider) MustMeter(ctx context.Context) (metric.Meter, func(context.Context) error) {
	meter, flush, err := p.Meter(ctx)
	if err != nil {
		panic(err)
	}

	return meter, flush
}
