package interfaces

import (
	"github.com/jailtonjunior94/order/pkg/database"
	"github.com/jailtonjunior94/order/pkg/observability"
)

type RepositoryFactory interface {
	OrderRepository(db database.DBTX, o11y observability.Observability) OrderRepository
	OutboxRepository(db database.DBTX, o11y observability.Observability) OutboxRepository
}
