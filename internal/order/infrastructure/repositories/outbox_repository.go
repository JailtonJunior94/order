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

func (r *outboxRepository) FindAll(ctx context.Context, wasPublished bool) ([]*entities.Outbox, error) {
	ctx, span := r.o11y.Start(ctx, "outbox_repository.find_all")
	defer span.End()

	query := `select
				id,
				event_name,
				was_published,
				published_at,
				payload,
				created_at
			  from
				outbox o
			  where
				o.was_published = $1`

	rows, err := r.tx.QueryContext(ctx, query, wasPublished)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, "error find all outbox", o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}
	defer rows.Close()

	var outboxes []*entities.Outbox
	for rows.Next() {
		var outbox entities.Outbox
		err := rows.Scan(
			&outbox.ID.Value,
			&outbox.EventName,
			&outbox.WasPublished,
			&outbox.PublishedAt.Time,
			&outbox.Payload,
			&outbox.CreatedAt,
		)
		if err != nil {
			span.AddAttributes(ctx, o11y.Error, "error scan row", o11y.Attributes{Key: "error", Value: err})
			return nil, err
		}
		outboxes = append(outboxes, &outbox)
	}
	return outboxes, nil
}

func (r *outboxRepository) Update(ctx context.Context, outbox *entities.Outbox) error {
	ctx, span := r.o11y.Start(ctx, "outbox_repository.update")
	defer span.End()

	query := `update
				outbox
			  set
				was_published = $1,
				published_at = $2
			  where
				id = $3`

	_, err := r.tx.ExecContext(
		ctx,
		query,
		outbox.WasPublished,
		outbox.PublishedAt.Time,
		outbox.ID.Value,
	)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, "error update outbox", o11y.Attributes{Key: "error", Value: err})
		return err
	}
	return nil
}
