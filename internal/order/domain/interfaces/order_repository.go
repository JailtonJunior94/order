package interfaces

import (
	"context"

	"github.com/jailtonjunior94/order/internal/order/domain/entities"
)

type OrderRepository interface {
	Insert(ctx context.Context, order *entities.Order) error
	InsertItems(ctx context.Context, items []*entities.OrderItem) error
}
