package orders

import (
	"context"

	"github.com/lucasmls/dd/internal/pkg/protog"
)

type Repository interface {
	Create(ctx context.Context, order *protog.Order) (*protog.Order, error)
}
