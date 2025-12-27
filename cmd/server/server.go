package server

import (
	"context"
	"fmt"
	"log"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/jailtonjunior94/order/internal/order"
	"github.com/jailtonjunior94/order/pkg/bundle"
	"github.com/jailtonjunior94/order/pkg/observability"
	"github.com/jailtonjunior94/order/pkg/observability/otel"
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
	// Parse configuration with defaults
	logLevel := observability.LogLevelInfo
	if ioc.Config.O11yConfig.LogLevel != "" {
		logLevel = observability.LogLevel(ioc.Config.O11yConfig.LogLevel)
	}

	logFormat := observability.LogFormatJSON
	if ioc.Config.O11yConfig.LogFormat == "text" {
		logFormat = observability.LogFormatText
	}

	sampleRate := 1.0
	if ioc.Config.O11yConfig.TraceSampleRate != "" {
		if rate, err := strconv.ParseFloat(ioc.Config.O11yConfig.TraceSampleRate, 64); err == nil {
			sampleRate = rate
		}
	}

	// Create observability provider
	config := &otel.Config{
		ServiceName:     ioc.Config.O11yConfig.OrderAPI,
		ServiceVersion:  ioc.Config.O11yConfig.ServiceVersion,
		Environment:     ioc.Config.Environment,
		OTLPEndpoint:    ioc.Config.O11yConfig.ExporterEndpoint,
		TraceSampleRate: sampleRate,
		LogLevel:        logLevel,
		LogFormat:       logFormat,
	}

	o11y, err := otel.NewProvider(ctx, config)
	if err != nil {
		log.Fatalf("failed to initialize observability: %v", err)
	}

	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		if err := o11y.Shutdown(shutdownCtx); err != nil {
			log.Printf("observability shutdown error: %v", err)
		}
	}()

	/* Close DBConnection */
	defer func() {
		if err := ioc.DB.Close(); err != nil {
			log.Printf("failed to close database: %v", err)
		}
	}()

	router := chi.NewRouter()
	router.Use(
		middleware.RealIP,
		middleware.RequestID,
		// TODO: Create new correlation middleware for observability
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
	order.RegisterOrderModule(ioc, o11y, router)

	/* Graceful shutdown */
	server := http.Server{
		ReadTimeout:       time.Duration(10) * time.Second,
		ReadHeaderTimeout: time.Duration(10) * time.Second,
		Handler:           router,
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", ioc.Config.HTTPConfig.Port))
	if err != nil {
		log.Printf("failed to create listener: %v", err)
		return
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
		log.Printf("server shutdown error: %v", err)
	}
}
