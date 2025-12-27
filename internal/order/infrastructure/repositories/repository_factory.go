package repositories

import (
	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/pkg/database"
	"github.com/jailtonjunior94/order/pkg/observability"
)

type repositoryFactory struct{}

func NewRepositoryFactory() *repositoryFactory {
	return &repositoryFactory{}
}

func (f *repositoryFactory) OrderRepository(db database.DBTX, o11y observability.Observability) interfaces.OrderRepository {
	return NewOrderRepository(db, o11y)
}

func (f *repositoryFactory) OutboxRepository(db database.DBTX, o11y observability.Observability) interfaces.OutboxRepository {
	return NewOutboxRepository(db, o11y)
}
