package grpc_port

import (
	"context"
	"errors"

	"github.com/lucasmls/dd/internal/orders"
	"github.com/lucasmls/dd/internal/orders/adapters/repositories"
	iProtog "github.com/lucasmls/dd/internal/pkg/protog"
	"github.com/lucasmls/dd/pkg/protog"
	otelCodes "go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// OrdersResolver ...
type OrdersResolver struct {
	logger           *zap.SugaredLogger
	tracer           trace.Tracer
	ordersRepository orders.Repository

	protog.UnimplementedOrdersServiceServer
}

// NewOrdersResolver ...
func NewOrdersResolver(
	logger *zap.SugaredLogger,
	tracer trace.Tracer,
	ordersRepository orders.Repository,
) (OrdersResolver, error) {

	return OrdersResolver{
		logger:           logger,
		tracer:           tracer,
		ordersRepository: ordersRepository,
	}, nil
}

// MustNewOrdersResolver ...
func MustNewOrdersResolver(
	logger *zap.SugaredLogger,
	tracer trace.Tracer,
	ordersRepository orders.Repository,
) OrdersResolver {

	ordersResolver, err := NewOrdersResolver(
		logger,
		tracer,
		ordersRepository,
	)
	if err != nil {
		panic(err)
	}

	return ordersResolver
}

var (
	InternalServerError = status.Error(codes.Internal, "Internal server error")
)

func (r OrdersResolver) Send(
	ctx context.Context,
	req *protog.SendOrderRequest,
) (*protog.SendOrderResponse, error) {
	ctx, span := r.tracer.Start(ctx, "OrdersResolver.Send")
	defer span.End()

	order := &iProtog.Order{
		Amount: req.Amount,
		Quote:  req.Quote,
	}

	order, err := r.ordersRepository.Create(ctx, order)
	if err != nil {
		if errors.Is(err, repositories.ErrStorageLimitReached) {
			r.logger.Warnw(
				"orders storage limit reached",
				zap.Error(err),
			)

			span.SetStatus(otelCodes.Error, err.Error())

			return nil, status.Error(codes.ResourceExhausted, err.Error())
		}

		r.logger.Error("failed to store sent order", zap.Error(err))

		return nil, InternalServerError
	}

	return &protog.SendOrderResponse{
		Id: order.Id,
	}, nil
}
