package repositories

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/lucasmls/dd/internal/pkg/protog"
	"go.uber.org/zap"
)

var (
	ErrInvalidStorageSize  = errors.New("invalid-storage-size")
	ErrStorageLimitReached = errors.New("storage-limit-reached")
)

type InMemoryOrdersRepository struct {
	logger      *zap.Logger
	storageSize int
	storage     map[string]*protog.Order
}

// NewInMemoryOrdersRepository creates a new InMemoryOrdersRepository.
func NewInMemoryOrdersRepository(
	logger *zap.Logger,
	storageSize int,
) (InMemoryOrdersRepository, error) {
	if storageSize == 0 {
		return InMemoryOrdersRepository{}, ErrInvalidStorageSize
	}

	return InMemoryOrdersRepository{
		logger:      logger,
		storageSize: storageSize,
		storage:     make(map[string]*protog.Order, storageSize),
	}, nil
}

// MustNewInMemoryOrdersRepository creates a new InMemoryOrdersRepository.
// It panics if any error is found.
func MustNewInMemoryOrdersRepository(
	logger *zap.Logger,
	storageSize int,
) InMemoryOrdersRepository {
	repo, err := NewInMemoryOrdersRepository(logger, storageSize)
	if err != nil {
		panic(err)
	}

	return repo
}

func (r InMemoryOrdersRepository) Create(
	ctx context.Context,
	order *protog.Order,
) (*protog.Order, error) {
	if len(r.storage) == r.storageSize {
		return nil, ErrStorageLimitReached
	}

	if order.Id == "" {
		order.Id = uuid.New().String()
	}

	r.storage[order.Id] = order
	return order, nil
}
