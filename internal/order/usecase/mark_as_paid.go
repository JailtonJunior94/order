package usecase

import (
	"context"

	"github.com/jailtonjunior94/order/internal/order/domain/dtos"
	"github.com/jailtonjunior94/order/internal/order/domain/entities"
	"github.com/jailtonjunior94/order/internal/order/domain/events"
	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/pkg/database"
	"github.com/jailtonjunior94/order/pkg/database/uow"
	"github.com/jailtonjunior94/order/pkg/observability"
	"github.com/jailtonjunior94/order/pkg/vos"
)

const (
	OrderPaidEvent = "order_paid"
)

type (
	MarkAsPaidUseCase interface {
		Execute(ctx context.Context, orderID vos.UUID) (*dtos.OrderOutput, error)
	}

	markAsPaidUseCase struct {
		uow               uow.UnitOfWork
		o11y               observability.Observability
		repositoryFactory interfaces.RepositoryFactory
	}
)

func NewMarkAsPaidUseCase(
	uow uow.UnitOfWork,
	o11y observability.Observability,
	repositoryFactory interfaces.RepositoryFactory,
) MarkAsPaidUseCase {
	return &markAsPaidUseCase{
		uow:               uow,
		o11y:               o11y,
		repositoryFactory: repositoryFactory,
	}
}

func (u *markAsPaidUseCase) Execute(ctx context.Context, orderID vos.UUID) (*dtos.OrderOutput, error) {
	ctx, span := u.o11y.Tracer().Start(ctx, "create_order_usecase.execute")
	defer span.End()

	var orderUpdated *entities.Order
	err := u.uow.Do(ctx, func(ctx context.Context, db database.DBTX) error {
		orderRepository := u.repositoryFactory.OrderRepository(db, u.o11y)
		outboxRepository := u.repositoryFactory.OutboxRepository(db, u.o11y)

		order, err := orderRepository.Find(ctx, orderID)
		if err != nil {
			span.AddEvent("error find order", observability.Any("error", err))
			return err
		}

		if order == nil {
			span.AddEvent("error order not found", observability.Any("error", err))
			return err
		}

		orderUpdated = order.MarkAsPaid()
		if err := orderRepository.Update(ctx, order); err != nil {
			span.AddEvent("error update order", observability.Any("error", err))
			return err
		}

		outbox, err := entities.NewOutbox(order.ID, OrderPaidEvent, events.NewOrderPaid(order.ID.String(), order.Total()))
		if err != nil {
			span.AddEvent("error create outbox", observability.Any("error", err))
			return err
		}

		if err := outboxRepository.Insert(ctx, outbox); err != nil {
			span.AddEvent("error insert outbox", observability.Any("error", err))
			return err
		}

		return nil
	})

	if err != nil {
		span.AddEvent("error mark as paid order", observability.Any("error", err))
		return nil, err
	}
	return dtos.NewOrderOutput(orderUpdated.ID.String(), orderUpdated.Status.String()), nil
}
