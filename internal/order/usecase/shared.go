package usecase

import (
	"errors"

	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/pkg/database/uow"
)

const (
	OrderRepository = "OrderRepository"
)

var (
	ErrInvalidRepositoryType = errors.New("invalid repository type")
)

func GetOrderRepository(tx uow.TX) (interfaces.OrderRepository, error) {
	repository, err := tx.Get(OrderRepository)
	if err != nil {
		return nil, err
	}

	orderRepository, ok := repository.(interfaces.OrderRepository)
	if !ok {
		return nil, ErrInvalidRepositoryType
	}
	return orderRepository, nil
}

func GetOutboxRepository(tx uow.TX) (interfaces.OutboxRepository, error) {
	repository, err := tx.Get(OutboxRepository)
	if err != nil {
		return nil, err
	}

	outboxRepository, ok := repository.(interfaces.OutboxRepository)
	if !ok {
		return nil, ErrInvalidRepositoryType
	}
	return outboxRepository, nil
}
