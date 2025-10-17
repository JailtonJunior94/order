package usecase

import (
	"context"

	"github.com/jailtonjunior94/order/internal/order/domain/dtos"
	"github.com/jailtonjunior94/order/internal/order/domain/factories"
	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/pkg/database"
	"github.com/jailtonjunior94/order/pkg/database/uow"
	"github.com/jailtonjunior94/order/pkg/o11y"
)

type (
	CreateOrderUseCase interface {
		Execute(ctx context.Context, input *dtos.OrderInput) (*dtos.OrderOutput, error)
	}

	createOrderUseCase struct {
		uow               uow.UnitOfWork
		telemetry         o11y.Telemetry
		repositoryFactory interfaces.RepositoryFactory
	}
)

func NewCreateOrderUseCase(
	uow uow.UnitOfWork,
	telemetry o11y.Telemetry,
	repositoryFactory interfaces.RepositoryFactory,
) CreateOrderUseCase {
	return &createOrderUseCase{
		uow:               uow,
		telemetry:         telemetry,
		repositoryFactory: repositoryFactory,
	}
}

func (c *createOrderUseCase) Execute(ctx context.Context, input *dtos.OrderInput) (*dtos.OrderOutput, error) {
	ctx, span := c.telemetry.Tracer().Start(ctx, "create_order_usecase.execute")
	defer span.End()

	newOrder, err := factories.CreateOrder(input)
	if err != nil {
		span.AddEvent("error creating order", o11y.Attribute{Key: "error", Value: err})
		return nil, err
	}

	err = c.uow.Do(ctx, func(ctx context.Context, db database.DBTX) error {
		orderRepository := c.repositoryFactory.OrderRepository(db, c.telemetry)
		if err := orderRepository.Insert(ctx, newOrder); err != nil {
			span.AddEvent("error inserting order", o11y.Attribute{Key: "error", Value: err})
			return err
		}

		if err := orderRepository.InsertItems(ctx, newOrder.Items); err != nil {
			span.AddEvent("error inserting items", o11y.Attribute{Key: "error", Value: err})
			return err
		}
		return nil
	})

	if err != nil {
		span.AddEvent("error creating order", o11y.Attribute{Key: "error", Value: err})
		return nil, err
	}

	c.telemetry.Metrics().AddCounter(ctx, "order_created_total", 1, nil)
	c.telemetry.Logger().Info(ctx, "order_created", o11y.Field{Key: "id", Value: newOrder.ID.String()})
	return dtos.NewOrderOutput(newOrder.ID.String(), newOrder.Status.String()), nil
}
