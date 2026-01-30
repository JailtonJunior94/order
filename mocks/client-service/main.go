package main

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ClientDTO represents the client data transfer object.
type ClientDTO struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Active bool   `json:"active"`
}

// ErrorResponse represents an error response.
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message"`
}

// Mock data
var clients = map[string]ClientDTO{
	"550e8400-e29b-41d4-a716-446655440000": {
		ID:     "550e8400-e29b-41d4-a716-446655440000",
		Name:   "John Doe",
		Email:  "john.doe@example.com",
		Active: true,
	},
	"550e8400-e29b-41d4-a716-446655440001": {
		ID:     "550e8400-e29b-41d4-a716-446655440001",
		Name:   "Jane Smith",
		Email:  "jane.smith@example.com",
		Active: true,
	},
	"inactive-client-id": {
		ID:     "inactive-client-id",
		Name:   "Jane Inactive",
		Email:  "jane.inactive@example.com",
		Active: false,
	},
}

func main() {
	r := chi.NewRouter()

	r.Use(middleware.Logger)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(60 * time.Second))

	r.Get("/clients/{clientID}", getClient)
	r.Get("/health", healthCheck)

	log.Println("Client Service Mock starting on :8081")
	log.Println("Available endpoints:")
	log.Println("  GET /clients/{clientID} - Get client by ID")
	log.Println("  GET /health - Health check")
	log.Println("")
	log.Println("Special client IDs for testing:")
	log.Println("  550e8400-e29b-41d4-a716-446655440000 - Active client (John Doe)")
	log.Println("  550e8400-e29b-41d4-a716-446655440001 - Active client (Jane Smith)")
	log.Println("  inactive-client-id - Inactive client")
	log.Println("  non-existent-id - Not found (404)")
	log.Println("  server-error-id - Internal error (500)")
	log.Println("  service-error-id - Service unavailable (503)")
	log.Println("  timeout-id - Gateway timeout (504)")
	log.Println("  rate-limited-id - Too many requests (429)")
	log.Println("  slow-* - Delayed response (2s)")

	if err := http.ListenAndServe(":8081", r); err != nil {
		log.Fatal(err)
	}
}

func getClient(w http.ResponseWriter, r *http.Request) {
	clientID := chi.URLParam(r, "clientID")
	w.Header().Set("Content-Type", "application/json")

	// Simulate slow response
	if strings.HasPrefix(clientID, "slow-") {
		time.Sleep(2 * time.Second)
		clientID = strings.TrimPrefix(clientID, "slow-")
	}

	// Special error cases
	switch clientID {
	case "server-error-id":
		w.WriteHeader(http.StatusInternalServerError)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "INTERNAL_ERROR",
			Message: "An unexpected error occurred",
		})
		return

	case "service-error-id":
		w.WriteHeader(http.StatusServiceUnavailable)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "SERVICE_UNAVAILABLE",
			Message: "The client service is temporarily unavailable",
		})
		return

	case "timeout-id":
		w.WriteHeader(http.StatusGatewayTimeout)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "GATEWAY_TIMEOUT",
			Message: "The upstream service did not respond in time",
		})
		return

	case "rate-limited-id":
		w.Header().Set("Retry-After", "30")
		w.WriteHeader(http.StatusTooManyRequests)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "RATE_LIMITED",
			Message: "Too many requests. Please retry after 30 seconds",
		})
		return
	}

	// Check if client exists
	client, exists := clients[clientID]
	if !exists {
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(ErrorResponse{
			Error:   "CLIENT_NOT_FOUND",
			Message: "Client with the specified ID was not found",
		})
		return
	}

	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(client)
}

func healthCheck(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status": "healthy",
	})
}
