package order

import (
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
	repositoryFactory := repositories.NewRepositoryFactory()
	createOrderUseCase := usecase.NewCreateOrderUseCase(uow, ioc.Observability, repositoryFactory)
	markAsPaidUseCaseUseCase := usecase.NewMarkAsPaidUseCase(uow, ioc.Observability, repositoryFactory)

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
	repositoryFactory := repositories.NewRepositoryFactory()
	brokerClient := kafka.NewKafkaClient(ioc.Config.KafkaConfig.Brokers[0], ioc.Observability)
	publishEventUseCase := usecase.NewPublishEventUseCase(uow, ioc.Config, ioc.Observability, brokerClient, repositoryFactory)
	return job.NewPublishEventHandler(ioc.Observability, publishEventUseCase)
}
