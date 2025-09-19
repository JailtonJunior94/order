package repositories

import (
	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/pkg/database"

	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
)

type repositoryFactory struct{}

func NewRepositoryFactory() *repositoryFactory {
	return &repositoryFactory{}
}

func (f *repositoryFactory) OrderRepository(db database.DBTX, o11y o11y.Observability) interfaces.OrderRepository {
	return NewOrderRepository(db, o11y)
}

func (f *repositoryFactory) OutboxRepository(db database.DBTX, o11y o11y.Observability) interfaces.OutboxRepository {
	return NewOutboxRepository(db, o11y)
}
