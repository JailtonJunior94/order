package interfaces

import (
	"context"

	"github.com/jailtonjunior94/order/internal/order/domain/entities"
)

type OutboxRepository interface {
	Insert(ctx context.Context, outbox *entities.Outbox) error
	Update(ctx context.Context, outbox *entities.Outbox) error
	FindAll(ctx context.Context, wasPublished bool) ([]*entities.Outbox, error)
}
