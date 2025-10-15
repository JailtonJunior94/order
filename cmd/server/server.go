package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/jailtonjunior94/order/internal/order"
	"github.com/jailtonjunior94/order/pkg/bundle"
	"github.com/jailtonjunior94/order/pkg/o11y"
	"github.com/jailtonjunior94/order/pkg/responses"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type apiServer struct {
}

func NewApiServer() *apiServer {
	return &apiServer{}
}

func (s *apiServer) Run() {
	ctx := context.Background()
	ioc := bundle.NewContainer(ctx)

	/* Observability */
	resource, err := o11y.NewServiceResource(ctx, ioc.Config.O11yConfig.OrderAPI, ioc.Config.O11yConfig.ServiceVersion, ioc.Config.Environment)
	if err != nil {
		log.Fatalf("failed to create resource: %v", err)
	}

	metrics, shutdown, err := o11y.NewMetrics(ctx, ioc.Config.O11yConfig.ExporterEndpoint, ioc.Config.O11yConfig.OrderAPI, resource)
	if err != nil {
		log.Fatalf("failed to create metrics: %v", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown metrics: %v", err)
		}
	}()

	tracer, shutdown, err := o11y.NewTracer(ctx, ioc.Config.O11yConfig.ExporterEndpoint, ioc.Config.O11yConfig.OrderAPI, resource)
	if err != nil {
		log.Fatalf("failed to create tracer: %v", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown tracer: %v", err)
		}
	}()

	logger, shutdown, err := o11y.NewLogger(ctx, tracer, ioc.Config.O11yConfig.ExporterEndpointHTTP, ioc.Config.O11yConfig.OrderAPI, resource)
	if err != nil {
		log.Fatalf("failed to create logger: %v", err)
	}
	defer func() {
		if err := shutdown(ctx); err != nil {
			log.Fatalf("failed to shutdown logger: %v", err)
		}
	}()

	telemetry, err := o11y.NewTelemetry(tracer, metrics, logger)
	if err != nil {
		log.Fatalf("failed to create telemetry: %v", err)
	}

	/* Close DBConnection */
	defer func() {
		if err := ioc.DB.Close(); err != nil {
			log.Fatal(err)
		}
	}()

	router := chi.NewRouter()
	router.Use(
		middleware.RealIP,
		middleware.RequestID,
		middleware.SetHeader("Content-Type", "application/json"),
		middleware.AllowContentType("application/json", "application/x-www-form-urlencoded"),
	)

	router.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		if err := ioc.DB.Ping(); err != nil {
			responses.Error(w, http.StatusInternalServerError, "database error connection failed or database is not running")
			return
		}
		responses.JSON(w, http.StatusOK, map[string]any{"status": "ok"})
	})

	/* Order */
	order.RegisterOrderModule(ioc, telemetry, router)

	/* Graceful shutdown */
	server := http.Server{
		ReadTimeout:       time.Duration(10) * time.Second,
		ReadHeaderTimeout: time.Duration(10) * time.Second,
		Handler:           router,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", ioc.Config.HTTPConfig.Port))
	if err != nil {
		log.Fatal(err)
	}

	go func() {
		if err := server.Serve(listener); err != nil && err != http.ErrServerClosed {
			log.Fatal(err)
		}
	}()

	s.gracefulShutdown(&server)
}

func (s *apiServer) gracefulShutdown(server *http.Server) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctxShutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := server.Shutdown(ctxShutdown); err != nil {
		log.Fatal(err)
	}
}
