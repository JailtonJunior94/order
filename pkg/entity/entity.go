package entity

import (
	"time"

	"github.com/jailtonjunior94/order/pkg/vos"
)

type Base struct {
	ID        vos.UUID
	CreatedAt time.Time
	UpdatedAt vos.NullableTime
}
