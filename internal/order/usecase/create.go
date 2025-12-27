package usecase

import (
	"context"

	"github.com/jailtonjunior94/order/internal/order/domain/dtos"
	domainErrors "github.com/jailtonjunior94/order/internal/order/domain/errors"
	"github.com/jailtonjunior94/order/internal/order/domain/factories"
	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/pkg/database"
	"github.com/jailtonjunior94/order/pkg/database/uow"
	"github.com/jailtonjunior94/order/pkg/observability"
)

type (
	CreateOrderUseCase interface {
		Execute(ctx context.Context, input *dtos.OrderInput) (*dtos.OrderOutput, error)
	}

	createOrderUseCase struct {
		uow               uow.UnitOfWork
		o11y               observability.Observability
		repositoryFactory interfaces.RepositoryFactory
		clientService     interfaces.ClientService
	}
)

func NewCreateOrderUseCase(
	uow uow.UnitOfWork,
	o11y observability.Observability,
	repositoryFactory interfaces.RepositoryFactory,
	clientService interfaces.ClientService,
) CreateOrderUseCase {
	return &createOrderUseCase{
		uow:               uow,
		o11y:               o11y,
		repositoryFactory: repositoryFactory,
		clientService:     clientService,
	}
}

func (c *createOrderUseCase) Execute(ctx context.Context, input *dtos.OrderInput) (*dtos.OrderOutput, error) {
	ctx, span := c.o11y.Tracer().Start(ctx, "create_order_usecase.execute")
	defer span.End()

	// Validate client ID
	if input.ClientID == "" {
		span.AddEvent("empty client_id", observability.String("error", "client_id is required"))
		return nil, domainErrors.ErrEmptyClientID
	}

	// Get and validate client
	client, err := c.clientService.GetClient(ctx, input.ClientID)
	if err != nil {
		span.AddEvent("error getting client", observability.Error(err))
		return nil, err
	}

	span.AddEvent("client validated successfully",
		observability.String("client_id", client.ID),
		observability.String("client_name", client.Name),
	)

	newOrder, err := factories.CreateOrder(input)
	if err != nil {
		span.AddEvent("error creating order", observability.Error(err))
		return nil, err
	}

	err = c.uow.Do(ctx, func(ctx context.Context, db database.DBTX) error {
		orderRepository := c.repositoryFactory.OrderRepository(db, c.o11y)
		if err := orderRepository.Insert(ctx, newOrder); err != nil {
			span.AddEvent("error inserting order", observability.Error(err))
			return err
		}

		if err := orderRepository.InsertItems(ctx, newOrder.Items); err != nil {
			span.AddEvent("error inserting items", observability.Error(err))
			return err
		}
		return nil
	})

	if err != nil {
		span.AddEvent("error creating order", observability.Error(err))
		return nil, err
	}

	counter := c.o11y.Metrics().Counter("order_created_total", "Total orders created", "1")
	counter.Add(ctx, 1)

	c.o11y.Logger().Info(ctx, "order_created", observability.String("id", newOrder.ID.String()))
	return dtos.NewOrderOutput(newOrder.ID.String(), newOrder.Status.String()), nil
}
