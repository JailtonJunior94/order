package usecase

import (
	"context"

	"github.com/jailtonjunior94/order/internal/order/domain/dtos"
	"github.com/jailtonjunior94/order/internal/order/domain/factories"
	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/pkg/database/uow"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"go.opentelemetry.io/otel/metric"
)

type (
	CreateOrderUseCase interface {
		Execute(ctx context.Context, input *dtos.OrderInput) (*dtos.OrderOutput, error)
	}

	createOrderUseCase struct {
		uow             uow.UnitOfWork
		o11y            o11y.Observability
		metrics         *createOrderMetrics
		orderRepository interfaces.OrderRepository
	}

	createOrderMetrics struct {
		orderCounter metric.Int64Counter
	}
)

func NewCreateOrderUseCase(
	o11y o11y.Observability,
	uow uow.UnitOfWork,
	orderRepository interfaces.OrderRepository,
) CreateOrderUseCase {
	uc := &createOrderUseCase{
		uow:             uow,
		o11y:            o11y,
		orderRepository: orderRepository,
	}
	uc.addMetrics()
	return uc
}

func (c *createOrderUseCase) Execute(ctx context.Context, input *dtos.OrderInput) (*dtos.OrderOutput, error) {
	ctx, span := c.o11y.Start(ctx, "create_order_usecase.execute")
	defer span.End()

	newOrder, err := factories.CreateOrder(input)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, "error create order", o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}

	err = c.uow.Do(ctx, func(ctx context.Context) error {
		if err := c.orderRepository.Insert(ctx, newOrder); err != nil {
			span.AddAttributes(ctx, o11y.Error, "error insert order", o11y.Attributes{Key: "error", Value: err})
			return err
		}

		if err := c.orderRepository.InsertItems(ctx, newOrder.Items); err != nil {
			span.AddAttributes(ctx, o11y.Error, "error insert items", o11y.Attributes{Key: "error", Value: err})
			return err
		}
		return nil
	})

	if err != nil {
		span.AddAttributes(ctx, o11y.Error, "error create order", o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}

	c.metrics.orderCounter.Add(ctx, 1)
	return dtos.NewOrderOutput(newOrder.ID.String(), newOrder.Status.String()), nil
}

func (c *createOrderUseCase) addMetrics() (*createOrderUseCase, error) {
	orderCounter, err := c.o11y.Meter().Int64Counter("order_created_total", metric.WithDescription("Total number of orders created"))
	if err != nil {
		return nil, err
	}
	c.metrics = &createOrderMetrics{orderCounter: orderCounter}
	return c, nil
}
