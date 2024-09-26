package usecase

import (
	"context"

	"github.com/jailtonjunior94/order/internal/order/domain/dtos"
	"github.com/jailtonjunior94/order/internal/order/domain/entities"
	"github.com/jailtonjunior94/order/internal/order/domain/events"
	"github.com/jailtonjunior94/order/pkg/database/uow"
	"github.com/jailtonjunior94/order/pkg/o11y"
	"github.com/jailtonjunior94/order/pkg/vos"
)

const (
	OrderPaidEvent   = "order_paid"
	OutboxRepository = "OutboxRepository"
)

type (
	MarkAsPaidUseCase interface {
		Execute(ctx context.Context, orderID vos.UUID) (*dtos.OrderOutput, error)
	}

	markAsPaidUseCase struct {
		uow  uow.UnitOfWork
		o11y o11y.Observability
	}
)

func NewMarkAsPaidUseCase(
	uow uow.UnitOfWork,
	o11y o11y.Observability,
) MarkAsPaidUseCase {
	return &markAsPaidUseCase{
		uow:  uow,
		o11y: o11y,
	}
}

func (u *markAsPaidUseCase) Execute(ctx context.Context, orderID vos.UUID) (*dtos.OrderOutput, error) {
	ctx, span := u.o11y.Start(ctx, "create_order_usecase.execute")
	defer span.End()

	var orderUpdated *entities.Order
	err := u.uow.Do(ctx, func(ctx context.Context, tx uow.TX) error {
		orderRepository, err := GetOrderRepository(tx)
		if err != nil {
			span.AddAttributes(ctx, o11y.Error, "error get order repository", o11y.Attributes{Key: "error", Value: err})
			return err
		}

		outboxRepository, err := GetOutboxRepository(tx)
		if err != nil {
			span.AddAttributes(ctx, o11y.Error, "error get outbox repository", o11y.Attributes{Key: "error", Value: err})
			return err
		}

		order, err := orderRepository.Find(ctx, orderID)
		if err != nil {
			span.AddAttributes(ctx, o11y.Error, "error find order", o11y.Attributes{Key: "error", Value: err})
			return err
		}

		if order == nil {
			span.AddAttributes(ctx, o11y.Error, "error order not found", o11y.Attributes{Key: "error", Value: err})
			return err
		}

		orderUpdated = order.MarkAsPaid()
		if err := orderRepository.Update(ctx, order); err != nil {
			span.AddAttributes(ctx, o11y.Error, "error update order", o11y.Attributes{Key: "error", Value: err})
			return err
		}

		outbox, err := entities.NewOutbox(order.ID, OrderPaidEvent, events.NewOrderPaid(order.ID.String(), order.Total()))
		if err != nil {
			span.AddAttributes(ctx, o11y.Error, "error create outbox", o11y.Attributes{Key: "error", Value: err})
			return err
		}

		if err := outboxRepository.Insert(ctx, outbox); err != nil {
			span.AddAttributes(ctx, o11y.Error, "error insert outbox", o11y.Attributes{Key: "error", Value: err})
			return err
		}

		return nil
	})

	if err != nil {
		span.AddAttributes(ctx, o11y.Error, "error mark as paid order", o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}
	return dtos.NewOrderOutput(orderUpdated.ID.String(), orderUpdated.Status.String()), nil
}
