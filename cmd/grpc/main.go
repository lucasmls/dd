package main

import (
	"context"

	"github.com/lucasmls/dd/internal/orders/adapters/repositories"
	resolvers "github.com/lucasmls/dd/internal/orders/ports/grpc"

	"github.com/lucasmls/dd/pkg/grpc"
	"github.com/lucasmls/dd/pkg/protog"
	"github.com/lucasmls/dd/pkg/trace"
	"go.uber.org/zap"
	gGRPC "google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	sLogger := logger.Sugar()

	otelCollectorGrpcClient := grpc.MustNewClient("localhost:4317", sLogger, []gGRPC.DialOption{})
	otelCollectorGrpcConnection := otelCollectorGrpcClient.MustConnect(ctx)

	otlpTraceProvider := trace.MustNewOtlpProvider("dd", "1.0.0", 1.0, otelCollectorGrpcConnection)
	otlpTracer, flush := otlpTraceProvider.MustTracer(ctx)
	defer flush(ctx)

	orderRepository := repositories.MustNewInMemoryOrdersRepository(
		sLogger,
		otlpTracer,
		10,
	)

	ordersResolver := resolvers.MustNewOrdersResolver(
		sLogger,
		otlpTracer,
		orderRepository,
	)

	grpcServer := grpc.MustNewServer(
		logger,
		3001,
		nil,
		func(server gGRPC.ServiceRegistrar) {
			protog.RegisterOrdersServiceServer(server, ordersResolver)
		},
	)

	if err := grpcServer.Run(ctx); err != nil {
		logger.Fatal("failed to run gRPC server", zap.Error(err))
	}
}
