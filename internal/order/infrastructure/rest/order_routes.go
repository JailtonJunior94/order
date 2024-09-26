package rest

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

type (
	Routes     func(orderRoute *orderRoute)
	orderRoute struct {
		CreateOrderHandler func(w http.ResponseWriter, r *http.Request)
		MarkAsPaidHandler  func(w http.ResponseWriter, r *http.Request)
	}
)

func NewOrderRoute(router *chi.Mux, orderRoutes ...Routes) *orderRoute {
	route := &orderRoute{}
	for _, orderRoute := range orderRoutes {
		orderRoute(route)
	}
	route.Register(router)
	return route
}

func (u *orderRoute) Register(router *chi.Mux) {
	router.Route("/api/v1/orders", func(r chi.Router) {
		r.Post("/", u.CreateOrderHandler)
		r.Patch("/{id}", u.MarkAsPaidHandler)
	})
}

func WithCreateOrderHandler(handler func(w http.ResponseWriter, r *http.Request)) Routes {
	return func(orderRoute *orderRoute) {
		orderRoute.CreateOrderHandler = handler
	}
}

func WithMarkAsPaidHandler(handler func(w http.ResponseWriter, r *http.Request)) Routes {
	return func(orderRoute *orderRoute) {
		orderRoute.MarkAsPaidHandler = handler
	}
}
