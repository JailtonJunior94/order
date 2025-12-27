package repositories

import (
	"context"

	"github.com/jailtonjunior94/order/internal/order/domain/entities"
	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/pkg/database"
	"github.com/jailtonjunior94/order/pkg/observability"
)

type outboxRepository struct {
	db        database.DBTX
	o11y observability.Observability
}

func NewOutboxRepository(db database.DBTX, o11y observability.Observability) interfaces.OutboxRepository {
	return &outboxRepository{
		db:        db,
		o11y: o11y,
	}
}

func (r *outboxRepository) Insert(ctx context.Context, outbox *entities.Outbox) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "outbox_repository.insert")
	defer span.End()
	query := `insert into
				outbox (id, event_name, was_published, published_at, payload, created_at)
			  values
				($1, $2, $3, $4, $5, $6)`

	_, err := r.db.ExecContext(
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
		span.AddEvent("error insert outbox", observability.Any("error", err))
		return err
	}
	return nil
}

func (r *outboxRepository) FindAll(ctx context.Context, wasPublished bool) ([]*entities.Outbox, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "outbox_repository.find_all")
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

	rows, err := r.db.QueryContext(ctx, query, wasPublished)
	if err != nil {
		span.AddEvent("error find all outbox", observability.Any("error", err))
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			span.AddEvent("error closing rows", observability.Any("error", err))
		}
	}()

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
			span.AddEvent("error scan row", observability.Any("error", err))
			return nil, err
		}
		outboxes = append(outboxes, &outbox)
	}
	return outboxes, nil
}

func (r *outboxRepository) Update(ctx context.Context, outbox *entities.Outbox) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "outbox_repository.update")
	defer span.End()

	query := `update
				outbox
			  set
				was_published = $1,
				published_at = $2
			  where
				id = $3`

	_, err := r.db.ExecContext(
		ctx,
		query,
		outbox.WasPublished,
		outbox.PublishedAt.Time,
		outbox.ID.Value,
	)
	if err != nil {
		span.AddEvent("error update outbox", observability.Any("error", err))
		return err
	}
	return nil
}
