package grpc_port

import (
	"context"

	"github.com/lucasmls/dd/internal/orders"
	iProtog "github.com/lucasmls/dd/internal/pkg/protog"
	"github.com/lucasmls/dd/pkg/protog"
	"go.uber.org/zap"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

// OrdersResolver ...
type OrdersResolver struct {
	logger           *zap.Logger
	ordersRepository orders.Repository

	protog.UnimplementedOrdersServiceServer
}

// NewOrdersResolver ...
func NewOrdersResolver(
	logger *zap.Logger,
	ordersRepository orders.Repository,
) (OrdersResolver, error) {

	return OrdersResolver{
		logger:           logger,
		ordersRepository: ordersRepository,
	}, nil
}

// MustNewOrdersResolver ...
func MustNewOrdersResolver(
	logger *zap.Logger,
	ordersRepository orders.Repository,
) OrdersResolver {

	ordersResolver, err := NewOrdersResolver(
		logger,
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
	order := &iProtog.Order{
		Amount: req.Amount,
		Quote:  req.Quote,
	}

	order, err := r.ordersRepository.Create(ctx, order)
	if err != nil {
		return nil, InternalServerError
	}

	return &protog.SendOrderResponse{
		Id: order.Id,
	}, nil
}
