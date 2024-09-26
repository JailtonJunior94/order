package entities

import (
	"time"

	"github.com/jailtonjunior94/order/internal/order/domain/vos"
	"github.com/jailtonjunior94/order/pkg/entity"
	sharedVos "github.com/jailtonjunior94/order/pkg/vos"
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

func (o *Order) MarkAsPaid() *Order {
	o.Status = vos.StatusPaid
	o.UpdatedAt = sharedVos.NewNullableTime(time.Now().UTC())
	return o
}

func (o *Order) AddItems(items []*OrderItem) {
	o.Items = items
}

func (o *Order) Total() float64 {
	var total float64
	for _, item := range o.Items {
		total += item.Price * float64(item.Quantity)
	}
	return total
}
