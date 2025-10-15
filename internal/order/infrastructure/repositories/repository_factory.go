package repositories

import (
	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/pkg/database"
	"github.com/jailtonjunior94/order/pkg/o11y"
)

type repositoryFactory struct{}

func NewRepositoryFactory() *repositoryFactory {
	return &repositoryFactory{}
}

func (f *repositoryFactory) OrderRepository(db database.DBTX, telemetry o11y.Telemetry) interfaces.OrderRepository {
	return NewOrderRepository(db, telemetry)
}

func (f *repositoryFactory) OutboxRepository(db database.DBTX, telemetry o11y.Telemetry) interfaces.OutboxRepository {
	return NewOutboxRepository(db, telemetry)
}
