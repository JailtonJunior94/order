package entities

import (
	"time"

	"github.com/jailtonjunior94/outbox/internal/order/domain/vos"
	"github.com/jailtonjunior94/outbox/pkg/entity"
	sharedVos "github.com/jailtonjunior94/outbox/pkg/vos"
)

type Order struct {
	entity.Base
	Status vos.Status
	Items  []*OrderItem
}

func NewOrder() *Order {
	return &Order{
		Status: vos.StatusPending,
		Base: entity.Base{
			CreatedAt: time.Now().UTC(),
		},
	}
}

func (o *Order) MarkAsPaid() {
	o.Status = vos.StatusPaid
	o.UpdatedAt = sharedVos.NewNullableTime(time.Now().UTC())
}

func (o *Order) AddItems(items []*OrderItem) {
	o.Items = items
}
