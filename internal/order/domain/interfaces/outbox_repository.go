package interfaces

import (
	"context"

	"github.com/jailtonjunior94/order/internal/order/domain/entities"
)

type OutboxRepository interface {
	Insert(ctx context.Context, outbox *entities.Outbox) error
}
