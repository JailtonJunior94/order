package events

import (
	"context"
	"time"

	"github.com/jailtonjunior94/order/pkg/vos"
)

type Event interface {
	GetEventType() string
	GetDateTime() time.Time
	GetPayload() any
	SetPayload(payload any)
	SetKey(key vos.UUID)
	GetKey() []byte
}

type EventDispatcher interface {
	Register(eventType string, handler EventHandler) error
	Dispatch(ctx context.Context, event Event) error
	Remove(eventType string, handler EventHandler) error
	Has(eventType string, handler EventHandler) bool
	Clear()
}

type EventHandler interface {
	Handle(ctx context.Context, event Event) error
}
