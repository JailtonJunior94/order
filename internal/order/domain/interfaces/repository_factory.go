package interfaces

import (
	"github.com/JailtonJunior94/devkit-go/pkg/o11y"
	"github.com/jailtonjunior94/order/pkg/database"
)

type RepositoryFactory interface {
	OrderRepository(db database.DBTX, o11y o11y.Observability) OrderRepository
	OutboxRepository(db database.DBTX, o11y o11y.Observability) OutboxRepository
}
