package usecase

import (
	"context"
	"errors"

	"github.com/jailtonjunior94/order/internal/order/domain/dtos"
	"github.com/jailtonjunior94/order/internal/order/domain/factories"
	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/pkg/database/uow"
	"github.com/jailtonjunior94/order/pkg/o11y"
)

const (
	OrderRepository = "OrderRepository"
)

var (
	ErrInvalidRepositoryType = errors.New("invalid repository type")
)

type (
	CreateOrderUseCase interface {
		Execute(ctx context.Context, input *dtos.OrderInput) (*dtos.OrderOutput, error)
	}

	createOrderUseCase struct {
		uow  uow.UnitOfWork
		o11y o11y.Observability
	}
)

func NewCreateOrderUseCase(
	uow uow.UnitOfWork,
	o11y o11y.Observability,
) CreateOrderUseCase {
	return &createOrderUseCase{
		uow:  uow,
		o11y: o11y,
	}
}

func (c *createOrderUseCase) Execute(ctx context.Context, input *dtos.OrderInput) (*dtos.OrderOutput, error) {
	ctx, span := c.o11y.Start(ctx, "create_order_usecase.execute")
	defer span.End()

	newOrder, err := factories.CreateOrder(input)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, "error create order", o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}

	err = c.uow.Do(ctx, func(ctx context.Context, tx uow.TX) error {
		orderRepository, err := c.getOrderRepository(tx)
		if err != nil {
			span.AddAttributes(ctx, o11y.Error, "error get order repository", o11y.Attributes{Key: "error", Value: err})
			return err
		}

		if err := orderRepository.Insert(ctx, newOrder); err != nil {
			span.AddAttributes(ctx, o11y.Error, "error insert order", o11y.Attributes{Key: "error", Value: err})
			return err
		}

		if err := orderRepository.InsertItems(ctx, newOrder.Items); err != nil {
			span.AddAttributes(ctx, o11y.Error, "error insert items", o11y.Attributes{Key: "error", Value: err})
			return err
		}
		return nil
	})

	if err != nil {
		span.AddAttributes(ctx, o11y.Error, "error create order", o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}
	return dtos.NewOrderOutput(newOrder.ID.String(), newOrder.Status.String()), nil
}

func (c *createOrderUseCase) getOrderRepository(tx uow.TX) (interfaces.OrderRepository, error) {
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
