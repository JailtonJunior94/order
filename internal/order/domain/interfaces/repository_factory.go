package interfaces

import (
	"github.com/jailtonjunior94/order/pkg/database"
	"github.com/jailtonjunior94/order/pkg/o11y"
)

type RepositoryFactory interface {
	OrderRepository(db database.DBTX, telemetry o11y.Telemetry) OrderRepository
	OutboxRepository(db database.DBTX, telemetry o11y.Telemetry) OutboxRepository
}
