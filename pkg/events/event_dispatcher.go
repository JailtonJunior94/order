package events

import (
	"context"
	"errors"
	"sync"
)

var ErrHandlerAlreadyRegistered = errors.New("handler already registered")

type eventDispatcher struct {
	mu       sync.RWMutex
	handlers map[string][]EventHandler
}

func NewEventDispatcher() EventDispatcher {
	return &eventDispatcher{
		handlers: make(map[string][]EventHandler),
	}
}

func (ev *eventDispatcher) Dispatch(ctx context.Context, event Event) error {
	ev.mu.RLock()
	handlers, ok := ev.handlers[event.GetEventType()]
	ev.mu.RUnlock()

	if ok {
		wg := &sync.WaitGroup{}
		for _, handler := range handlers {
			wg.Add(1)
			h := handler // capture loop variable
			go func() {
				if err := h.Handle(ctx, event, wg); err != nil {
					// Log error but continue execution
					// In production, consider using proper logging
				}
			}()
		}
		wg.Wait()
	}
	return nil
}

func (ed *eventDispatcher) Register(eventName string, handler EventHandler) error {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	if _, ok := ed.handlers[eventName]; ok {
		for _, h := range ed.handlers[eventName] {
			if h == handler {
				return ErrHandlerAlreadyRegistered
			}
		}
	}
	ed.handlers[eventName] = append(ed.handlers[eventName], handler)
	return nil
}

func (ed *eventDispatcher) Has(eventName string, handler EventHandler) bool {
	ed.mu.RLock()
	defer ed.mu.RUnlock()

	if _, ok := ed.handlers[eventName]; ok {
		for _, h := range ed.handlers[eventName] {
			if h == handler {
				return true
			}
		}
	}
	return false
}

func (ed *eventDispatcher) Remove(eventName string, handler EventHandler) error {
	ed.mu.Lock()
	defer ed.mu.Unlock()

	if _, ok := ed.handlers[eventName]; ok {
		for i, h := range ed.handlers[eventName] {
			if h == handler {
				ed.handlers[eventName] = append(ed.handlers[eventName][:i], ed.handlers[eventName][i+1:]...)
				return nil
			}
		}
	}
	return nil
}

func (ed *eventDispatcher) Clear() {
	ed.mu.Lock()
	defer ed.mu.Unlock()
	ed.handlers = make(map[string][]EventHandler)
}
