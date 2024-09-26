package entities

import (
	"time"

	"github.com/jailtonjunior94/outbox/pkg/entity"
	"github.com/jailtonjunior94/outbox/pkg/vos"
)

type OrderItem struct {
	entity.Base
	OrderID     vos.UUID
	ProductName string
	Price       float64
	Quantity    uint
}

func NewOrderItem(orderID vos.UUID, productName string, price float64, quantity uint) *OrderItem {
	return &OrderItem{
		OrderID:     orderID,
		ProductName: productName,
		Price:       price,
		Quantity:    quantity,
		Base: entity.Base{
			CreatedAt: time.Now().UTC(),
		},
	}
}
