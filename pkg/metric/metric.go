package metric

import (
	"context"

	"go.opentelemetry.io/otel/metric"
	metricSdk "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type MetricClient struct {
	applicationName    string
	applicationVersion string
	exporter           metricSdk.Exporter
}

func NewMetricClient(
	applicationName string,
	applicationVersion string,
	exporter metricSdk.Exporter,
) (*MetricClient, error) {
	if applicationName == "" {
		return nil, ErrMissingApplicationName
	}

	if applicationVersion == "" {
		applicationVersion = "unknown_version"
	}

	return &MetricClient{
		applicationName:    applicationName,
		applicationVersion: applicationVersion,
		exporter:           exporter,
	}, nil
}

func MustNewMetricClient(
	applicationName string,
	applicationVersion string,
	exporter metricSdk.Exporter,
) *MetricClient {
	client, err := NewMetricClient(
		applicationName,
		applicationVersion,
		exporter,
	)
	if err != nil {
		panic(err)
	}

	return client
}

func (c MetricClient) Meter(ctx context.Context) (metric.Meter, func(context.Context) error) {
	metricsOptions := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceNameKey.String(c.applicationName),
		semconv.ServiceVersionKey.String(c.applicationVersion),
	)

	meterProvider := metricSdk.NewMeterProvider(
		metricSdk.WithResource(metricsOptions),
		metricSdk.WithReader(metricSdk.NewPeriodicReader(c.exporter)),
	)

	return meterProvider.Meter(c.applicationName), meterProvider.ForceFlush
}
