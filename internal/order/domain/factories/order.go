package factories

import (
	"github.com/jailtonjunior94/order/internal/order/domain/dtos"
	"github.com/jailtonjunior94/order/internal/order/domain/entities"
	"github.com/jailtonjunior94/order/pkg/vos"
)

func CreateOrder(input *dtos.OrderInput) (*entities.Order, error) {
	orderID, err := vos.NewUUID()
	if err != nil {
		return nil, err
	}

	order := entities.NewOrder()
	order.ID = orderID

	for _, item := range input.Items {
		orderItem := entities.NewOrderItem(order.ID, item.ProductName, item.Price, item.Quantity)
		orderItemID, err := vos.NewUUID()
		if err != nil {
			return nil, err
		}
		orderItem.ID = orderItemID
		order.Items = append(order.Items, orderItem)
	}

	return order, nil
}
