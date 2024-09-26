package order

import (
	"database/sql"

	"github.com/jailtonjunior94/order/internal/order/infrastructure/repositories"
	"github.com/jailtonjunior94/order/internal/order/infrastructure/rest"
	"github.com/jailtonjunior94/order/internal/order/usecase"
	"github.com/jailtonjunior94/order/pkg/bundle"
	unitOfWork "github.com/jailtonjunior94/order/pkg/database/uow"

	"github.com/go-chi/chi/v5"
)

func RegisterOrderModule(ioc *bundle.Container, router *chi.Mux) {
	uow := unitOfWork.NewUnitOfWork(ioc.DB)
	uow.Register("OrderRepository", func(tx *sql.Tx) unitOfWork.Repository {
		return repositories.NewOrderRepository(ioc.DB, tx, ioc.Observability)
	})
	createOrderUseCase := usecase.NewCreateOrderUseCase(uow, ioc.Observability)
	orderHandler := rest.NewUserHandler(ioc.Observability, createOrderUseCase)
	rest.NewOrderRoute(router, rest.WithCreateOrderHandler(orderHandler.Create))
}
