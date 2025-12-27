package interfaces

import "context"

// ClientService represents the interface for client consultation service.
type ClientService interface {
	GetClient(ctx context.Context, clientID string) (*ClientDTO, error)
}

// ClientDTO represents the client data transfer object.
type ClientDTO struct {
	ID     string `json:"id"`
	Name   string `json:"name"`
	Email  string `json:"email"`
	Active bool   `json:"active"`
}
