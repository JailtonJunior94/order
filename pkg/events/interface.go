package events

import (
	"context"
	"sync"
	"time"

	"github.com/jailtonjunior94/order/pkg/vos"
)

type (
	EventHandler interface {
		Handle(ctx context.Context, event Event, wg *sync.WaitGroup) error
	}

	EventDispatcher interface {
		Register(eventType string, handler EventHandler) error
		Dispatch(ctx context.Context, event Event) error
		Remove(eventType string, handler EventHandler) error
		Has(eventType string, handler EventHandler) bool
		Clear()
	}

	Event interface {
		GetEventType() string
		GetDateTime() time.Time
		GetPayload() any
		SetPayload(payload any)
		SetKey(key vos.UUID)
		GetKey() []byte
	}
)
