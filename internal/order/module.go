package order

import (
	"database/sql"

	"github.com/jailtonjunior94/order/internal/order/infrastructure/job"
	"github.com/jailtonjunior94/order/internal/order/infrastructure/repositories"
	"github.com/jailtonjunior94/order/internal/order/infrastructure/rest"
	"github.com/jailtonjunior94/order/internal/order/usecase"
	"github.com/jailtonjunior94/order/pkg/bundle"
	unitOfWork "github.com/jailtonjunior94/order/pkg/database/uow"
	"github.com/jailtonjunior94/order/pkg/messaging/kafka"

	"github.com/go-chi/chi/v5"
)

func RegisterOrderModule(ioc *bundle.Container, router *chi.Mux) {
	uow := unitOfWork.NewUnitOfWork(ioc.DB)
	uow.Register("OrderRepository", func(tx *sql.Tx) unitOfWork.Repository {
		return repositories.NewOrderRepository(ioc.DB, tx, ioc.Observability)
	})

	uow.Register("OutboxRepository", func(tx *sql.Tx) unitOfWork.Repository {
		return repositories.NewOutboxRepository(ioc.DB, tx, ioc.Observability)
	})

	createOrderUseCase := usecase.NewCreateOrderUseCase(uow, ioc.Observability)
	markAsPaidUseCaseUseCase := usecase.NewMarkAsPaidUseCase(uow, ioc.Observability)

	orderHandler := rest.NewUserHandler(
		ioc.Observability,
		createOrderUseCase,
		markAsPaidUseCaseUseCase,
	)

	rest.NewOrderRoute(router,
		rest.WithCreateOrderHandler(orderHandler.Create),
		rest.WithMarkAsPaidHandler(orderHandler.MarkAsPaid),
	)
}

func RegisterPublishEventHandler(ioc *bundle.Container) *job.PublishEventHandler {
	uow := unitOfWork.NewUnitOfWork(ioc.DB)
	uow.Register("OutboxRepository", func(tx *sql.Tx) unitOfWork.Repository {
		return repositories.NewOutboxRepository(ioc.DB, tx, ioc.Observability)
	})

	brokeClient := kafka.NewKafkaClient(ioc.Config.KafkaConfig.Brokers[0], ioc.Observability)
	publishEventUseCase := usecase.NewPublishEventUseCase(ioc.Config, uow, brokeClient, ioc.Observability)
	return job.NewPublishEventHandler(ioc.Observability, publishEventUseCase)
}
