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
	ctx, span := r.tracer.Start(
		ctx,
		"OrdersResolver.Send",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	order := &iProtog.Order{
		Amount: req.Amount,
		Quote:  req.Quote,
	}

	order, err := r.ordersRepository.Create(ctx, order)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())

		if errors.Is(err, repositories.ErrStorageLimitReached) {
			r.logger.Warnw(
				"orders storage limit reached",
				zap.Error(err),
			)

			return nil, status.Error(codes.ResourceExhausted, err.Error())
		}

		r.logger.Error("failed to store sent order", zap.Error(err))

		return nil, InternalServerError
	}

	return &protog.SendOrderResponse{
		Id: order.Id,
	}, nil
}

func (r OrdersResolver) Find(
	ctx context.Context,
	req *protog.FindOrderRequest,
) (*protog.FindOrderResponse, error) {
	ctx, span := r.tracer.Start(
		ctx,
		"OrdersResolver.Find",
		trace.WithSpanKind(trace.SpanKindServer),
	)
	defer span.End()

	order, err := r.ordersRepository.Find(ctx, req.Id)
	if err != nil {
		span.SetStatus(otelCodes.Error, err.Error())

		if errors.Is(err, repositories.ErrOrderNotFound) {
			r.logger.Error(
				"order not found",
				zap.Error(err),
				zap.String("id", req.Id),
			)

			return nil, status.Error(codes.NotFound, err.Error())
		}

		r.logger.Error(
			"failed to find order",
			zap.Error(err),
			zap.String("id", req.Id),
		)

		return nil, InternalServerError
	}

	return &protog.FindOrderResponse{
		Id:     order.Id,
		Amount: order.Amount,
		Quote:  order.Quote,
	}, nil
}
