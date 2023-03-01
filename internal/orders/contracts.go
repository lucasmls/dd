package orders

import (
	"context"

	"github.com/lucasmls/dd/internal/pkg/protog"
)

type Repository interface {
	Find(ctx context.Context, id string) (*protog.Order, error)
	Create(ctx context.Context, order *protog.Order) (*protog.Order, error)
}
