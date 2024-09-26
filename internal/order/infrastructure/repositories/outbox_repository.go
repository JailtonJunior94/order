package repositories

import (
	"context"
	"database/sql"

	"github.com/jailtonjunior94/order/internal/order/domain/entities"
	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/pkg/o11y"
)

type outboxRepository struct {
	db   *sql.DB
	tx   *sql.Tx
	o11y o11y.Observability
}

func NewOutboxRepository(db *sql.DB, tx *sql.Tx, o11y o11y.Observability) interfaces.OutboxRepository {
	return &outboxRepository{
		db:   db,
		tx:   tx,
		o11y: o11y,
	}
}

func (r *outboxRepository) Insert(ctx context.Context, outbox *entities.Outbox) error {
	ctx, span := r.o11y.Start(ctx, "outbox_repository.insert")
	defer span.End()
	query := `insert into
				outbox (id, event_name, was_published, published_at, payload, created_at)
			  values
				($1, $2, $3, $4, $5, $6)`

	_, err := r.tx.ExecContext(
		ctx,
		query,
		outbox.ID.Value,
		outbox.EventName,
		outbox.WasPublished,
		outbox.PublishedAt.Time,
		outbox.Payload,
		outbox.CreatedAt,
	)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, "error insert outbox", o11y.Attributes{Key: "error", Value: err})
		return err
	}
	return nil
}
