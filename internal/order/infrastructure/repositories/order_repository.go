package repositories

import (
	"context"
	"database/sql"

	"github.com/jailtonjunior94/outbox/internal/order/domain/entities"
	"github.com/jailtonjunior94/outbox/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/outbox/pkg/o11y"
)

type orderRepository struct {
	db   *sql.DB
	tx   *sql.Tx
	o11y o11y.Observability
}

func NewOrderRepository(db *sql.DB, tx *sql.Tx, o11y o11y.Observability) interfaces.OrderRepository {
	return &orderRepository{
		db:   db,
		tx:   tx,
		o11y: o11y,
	}
}

func (r *orderRepository) Insert(ctx context.Context, order *entities.Order) error {
	ctx, span := r.o11y.Start(ctx, "order_repository.insert")
	defer span.End()

	query := `INSERT INTO
				orders (id, status, created_at, updated_at, deleted_at)
			  VALUES
				($1, $2, $3, $4, $5)`

	_, err := r.db.ExecContext(
		ctx,
		query,
		order.ID,
		order.Status,
		order.CreatedAt,
		order.UpdatedAt,
		order.DeletedAt,
	)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, "error insert order", o11y.Attributes{Key: "error", Value: err})
		return err
	}
	return nil
}

func (r *orderRepository) InsertItems(ctx context.Context, items []*entities.OrderItem) error {
	ctx, span := r.o11y.Start(ctx, "order_repository.insert_items")
	defer span.End()

	query := `INSERT INTO
				order_items (
					id,
					order_id,
					product_id,
					quantity,
					price,
					created_at,
					updated_at,
					deleted_at
					)
				VALUES
					($1, $2,$3, $4, $5, $6, $7, $8)`

	for _, item := range items {
		_, err := r.db.ExecContext(
			ctx,
			query,
			item.ID,
			item.OrderID,
			item.ProductName,
			item.Quantity,
			item.Price,
			item.CreatedAt,
			item.UpdatedAt,
			item.DeletedAt,
		)
		if err != nil {
			span.AddAttributes(ctx, o11y.Error, "error insert order item", o11y.Attributes{Key: "error", Value: err})
			return err
		}
	}

	return nil
}
