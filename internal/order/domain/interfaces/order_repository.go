package interfaces

import (
	"context"

	"github.com/jailtonjunior94/order/internal/order/domain/entities"
	"github.com/jailtonjunior94/order/internal/order/domain/vos"
	sharedVos "github.com/jailtonjunior94/order/pkg/vos"
)

type OrderRepository interface {
	Update(ctx context.Context, order *entities.Order) error
	Insert(ctx context.Context, order *entities.Order) error
	InsertItems(ctx context.Context, items []*entities.OrderItem) error
	Find(ctx context.Context, orderID sharedVos.UUID) (*entities.Order, error)
	FindAll(ctx context.Context, status vos.Status) ([]*entities.Order, error)
}
