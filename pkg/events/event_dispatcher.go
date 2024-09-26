package events

import (
	"context"
	"errors"
)

var ErrHandlerAlreadyRegistered = errors.New("handler already registered")

type eventDispatcher struct {
	handlers map[string][]EventHandler
}

func NewEventDispatcher() EventDispatcher {
	return &eventDispatcher{
		handlers: make(map[string][]EventHandler),
	}
}

func (ev *eventDispatcher) Dispatch(ctx context.Context, event Event) error {
	if handlers, ok := ev.handlers[event.GetEventType()]; ok {
		for _, handler := range handlers {
			if err := handler.Handle(ctx, event); err != nil {
				return err
			}
			continue
		}
	}
	return nil
}

func (ed *eventDispatcher) Register(eventName string, handler EventHandler) error {
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
	ed.handlers = make(map[string][]EventHandler)
}
