package events

import "github.com/jailtonjunior94/order/internal/order/domain/vos"

type OrderPaid struct {
	OrderID string  `json:"order_id"`
	Amount  float64 `json:"amount"`
	Status  string  `json:"status"`
}

func NewOrderPaid(orderID string, amount float64) *OrderPaid {
	return &OrderPaid{
		OrderID: orderID,
		Amount:  amount,
		Status:  vos.StatusPaid.String(),
	}
}
