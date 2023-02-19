package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/lucasmls/dd/internal/pkg/protog"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	ErrInvalidStorageSize  = errors.New("invalid-storage-size")
	ErrStorageLimitReached = errors.New("storage-limit-reached")
)

type InMemoryOrdersRepository struct {
	logger      *zap.SugaredLogger
	tracer      trace.Tracer
	storageSize int
	storage     map[string]*protog.Order
}

// NewInMemoryOrdersRepository creates a new InMemoryOrdersRepository.
func NewInMemoryOrdersRepository(
	logger *zap.SugaredLogger,
	tracer trace.Tracer,
	storageSize int,
) (InMemoryOrdersRepository, error) {
	if storageSize == 0 {
		return InMemoryOrdersRepository{}, ErrInvalidStorageSize
	}

	return InMemoryOrdersRepository{
		logger:      logger,
		tracer:      tracer,
		storageSize: storageSize,
		storage:     make(map[string]*protog.Order, storageSize),
	}, nil
}

// MustNewInMemoryOrdersRepository creates a new InMemoryOrdersRepository.
// It panics if any error is found.
func MustNewInMemoryOrdersRepository(
	logger *zap.SugaredLogger,
	tracer trace.Tracer,
	storageSize int,
) InMemoryOrdersRepository {
	repo, err := NewInMemoryOrdersRepository(logger, tracer, storageSize)
	if err != nil {
		panic(err)
	}

	return repo
}

func (r InMemoryOrdersRepository) Create(
	ctx context.Context,
	order *protog.Order,
) (*protog.Order, error) {
	ctx, span := r.tracer.Start(ctx, "InMemoryOrdersRepository.Create")
	defer span.End()

	if len(r.storage) == r.storageSize {
		return nil, ErrStorageLimitReached
	}

	if order.Id == "" {
		order.Id = uuid.New().String()
	}

	r.storage[order.Id] = order
	return order, nil
}
