package entities

import (
	"encoding/json"
	"time"

	"github.com/jailtonjunior94/order/pkg/entity"
	"github.com/jailtonjunior94/order/pkg/vos"
)

type Outbox struct {
	entity.Base
	EventName    string
	WasPublished bool
	PublishedAt  vos.NullableTime
	Payload      string
}

func NewOutbox(id vos.UUID, eventName string, payload any) (*Outbox, error) {
	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, err
	}

	return &Outbox{
		EventName: eventName,
		Payload:   string(jsonPayload),
		Base: entity.Base{
			ID:        id,
			CreatedAt: time.Now().UTC(),
		},
	}, nil
}

func (o *Outbox) MarkAsPublished() *Outbox {
	o.WasPublished = true
	o.PublishedAt = vos.NewNullableTime(time.Now().UTC())
	return o
}
