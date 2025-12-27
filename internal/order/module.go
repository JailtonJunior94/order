package order

import (
	"time"

	"github.com/jailtonjunior94/order/internal/order/infrastructure/job"
	"github.com/jailtonjunior94/order/internal/order/infrastructure/repositories"
	"github.com/jailtonjunior94/order/internal/order/infrastructure/rest"
	"github.com/jailtonjunior94/order/internal/order/infrastructure/services"
	"github.com/jailtonjunior94/order/internal/order/usecase"
	"github.com/jailtonjunior94/order/pkg/bundle"
	unitOfWork "github.com/jailtonjunior94/order/pkg/database/uow"
	httpclient "github.com/jailtonjunior94/order/pkg/http-client"
	"github.com/jailtonjunior94/order/pkg/messaging/kafka"
	"github.com/jailtonjunior94/order/pkg/observability"

	"github.com/go-chi/chi/v5"
)

func RegisterOrderModule(ioc *bundle.Container, o11y observability.Observability, router *chi.Mux) {
	uow := unitOfWork.NewUnitOfWork(ioc.DB)
	repositoryFactory := repositories.NewRepositoryFactory()

	// Parse timeout
	timeout, err := time.ParseDuration(ioc.Config.ClientServiceConfig.Timeout)
	if err != nil {
		timeout = 30 * time.Second // default
	}

	// Create base HTTP client with timeout
	baseClient := httpclient.NewHTTPClientWithOptions(
		httpclient.WithTimeout(timeout),
	)

	// Parse retry configuration
	initialDelay, err := time.ParseDuration(ioc.Config.ClientServiceConfig.InitialRetryDelay)
	if err != nil {
		initialDelay = 100 * time.Millisecond
	}

	maxDelay, err := time.ParseDuration(ioc.Config.ClientServiceConfig.MaxRetryDelay)
	if err != nil {
		maxDelay = 2 * time.Second
	}

	retryableStatusCodes := httpclient.ParseRetryableStatusCodes(ioc.Config.ClientServiceConfig.RetryableStatusCodes)
	if len(retryableStatusCodes) == 0 {
		// Default retryable status codes
		retryableStatusCodes = map[int]bool{
			408: true, 429: true, 500: true, 502: true, 503: true, 504: true,
		}
	}

	maxRetries := ioc.Config.ClientServiceConfig.MaxRetries
	if maxRetries == 0 {
		maxRetries = 3
	}

	// Wrap client with retry logic
	retryClient := httpclient.NewRetryClient(baseClient, httpclient.RetryConfig{
		MaxRetries:           maxRetries,
		InitialDelay:         initialDelay,
		MaxDelay:             maxDelay,
		RetryableStatusCodes: retryableStatusCodes,
		BackoffMultiplier:    2.0,
	}, o11y)

	// Create client service
	clientService := services.NewClientService(retryClient, o11y, ioc.Config.ClientServiceConfig)

	// Create use cases
	createOrderUseCase := usecase.NewCreateOrderUseCase(uow, o11y, repositoryFactory, clientService)
	markAsPaidUseCaseUseCase := usecase.NewMarkAsPaidUseCase(uow, o11y, repositoryFactory)

	orderHandler := rest.NewUserHandler(
	 o11y,
		createOrderUseCase,
		markAsPaidUseCaseUseCase,
	)

	rest.NewOrderRoute(router,
		rest.WithCreateOrderHandler(orderHandler.Create),
		rest.WithMarkAsPaidHandler(orderHandler.MarkAsPaid),
	)
}

func RegisterPublishEventHandler(ioc *bundle.Container, o11y observability.Observability) *job.PublishEventHandler {
	uow := unitOfWork.NewUnitOfWork(ioc.DB)
	repositoryFactory := repositories.NewRepositoryFactory()
	brokerClient := kafka.NewKafkaClient(ioc.Config.KafkaConfig.Brokers[0], o11y)
	publishEventUseCase := usecase.NewPublishEventUseCase(uow, ioc.Config, o11y, brokerClient, repositoryFactory)
	return job.NewPublishEventHandler(o11y, publishEventUseCase)
}
