package order

import (
	"github.com/jailtonjunior94/order/internal/order/infrastructure/job"
	"github.com/jailtonjunior94/order/internal/order/infrastructure/repositories"
	"github.com/jailtonjunior94/order/internal/order/infrastructure/rest"
	"github.com/jailtonjunior94/order/internal/order/usecase"
	"github.com/jailtonjunior94/order/pkg/bundle"
	unitOfWork "github.com/jailtonjunior94/order/pkg/database/uow"
	"github.com/jailtonjunior94/order/pkg/messaging/kafka"
	"github.com/jailtonjunior94/order/pkg/o11y"

	"github.com/go-chi/chi/v5"
)

func RegisterOrderModule(ioc *bundle.Container, telemetry o11y.Telemetry, router *chi.Mux) {
	uow := unitOfWork.NewUnitOfWork(ioc.DB)
	repositoryFactory := repositories.NewRepositoryFactory()
	createOrderUseCase := usecase.NewCreateOrderUseCase(uow, telemetry, repositoryFactory)
	markAsPaidUseCaseUseCase := usecase.NewMarkAsPaidUseCase(uow, telemetry, repositoryFactory)

	orderHandler := rest.NewUserHandler(
		telemetry,
		createOrderUseCase,
		markAsPaidUseCaseUseCase,
	)

	rest.NewOrderRoute(router,
		rest.WithCreateOrderHandler(orderHandler.Create),
		rest.WithMarkAsPaidHandler(orderHandler.MarkAsPaid),
	)
}

func RegisterPublishEventHandler(ioc *bundle.Container, telemetry o11y.Telemetry) *job.PublishEventHandler {
	uow := unitOfWork.NewUnitOfWork(ioc.DB)
	repositoryFactory := repositories.NewRepositoryFactory()
	brokerClient := kafka.NewKafkaClient(ioc.Config.KafkaConfig.Brokers[0], telemetry)
	publishEventUseCase := usecase.NewPublishEventUseCase(uow, ioc.Config, telemetry, brokerClient, repositoryFactory)
	return job.NewPublishEventHandler(telemetry, publishEventUseCase)
}
