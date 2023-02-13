package main

import (
	"context"

	"github.com/lucasmls/dd/internal/orders/adapters/repositories"
	resolvers "github.com/lucasmls/dd/internal/orders/ports/grpc"

	"github.com/lucasmls/dd/pkg/grpc"
	"github.com/lucasmls/dd/pkg/protog"
	"go.uber.org/zap"
	gGRPC "google.golang.org/grpc"
)

func main() {
	ctx := context.Background()

	logger, _ := zap.NewProduction()
	defer logger.Sync()

	sLogger := logger.Sugar()

	orderRepository := repositories.MustNewInMemoryOrdersRepository(
		logger,
		10,
	)

	ordersResolver := resolvers.MustNewOrdersResolver(
		sLogger,
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
