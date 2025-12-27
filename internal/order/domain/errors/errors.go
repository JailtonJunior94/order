package errors

import "fmt"

type DomainError struct {
	Code    string
	Message string
	Err     error
}

func (e *DomainError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("%s: %v", e.Message, e.Err)
	}
	return e.Message
}

func (e *DomainError) Unwrap() error {
	return e.Err
}

// Client service errors.
var (
	ErrClientNotFound = &DomainError{
		Code:    "CLIENT_NOT_FOUND",
		Message: "client not found",
	}
	ErrClientInactive = &DomainError{
		Code:    "CLIENT_INACTIVE",
		Message: "client is inactive",
	}
	ErrClientServiceTimeout = &DomainError{
		Code:    "CLIENT_SERVICE_TIMEOUT",
		Message: "client service timeout",
	}
	ErrClientServiceUnavailable = &DomainError{
		Code:    "CLIENT_SERVICE_UNAVAILABLE",
		Message: "client service unavailable",
	}
)

// Validation errors.
var (
	ErrEmptyClientID = &DomainError{
		Code:    "EMPTY_CLIENT_ID",
		Message: "client_id is required",
	}
)
