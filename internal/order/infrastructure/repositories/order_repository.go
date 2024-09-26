package repositories

import (
	"context"
	"database/sql"

	"github.com/jailtonjunior94/order/internal/order/domain/entities"
	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/internal/order/domain/vos"
	"github.com/jailtonjunior94/order/pkg/o11y"
	sharedVos "github.com/jailtonjunior94/order/pkg/vos"
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

func (r *orderRepository) FindAll(ctx context.Context, status vos.Status) ([]*entities.Order, error) {
	ctx, span := r.o11y.Start(ctx, "order_repository.find_all")
	defer span.End()

	query := `select 
				id,
				status,
				created_at,
				updated_at
			  from
				orders
			  where
				status = $1`

	rows, err := r.tx.QueryContext(ctx, query, status.String())
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, "error find all orders", o11y.Attributes{Key: "error", Value: err})
		return nil, err
	}
	defer rows.Close()

	var orders []*entities.Order
	for rows.Next() {
		var order entities.Order
		err := rows.Scan(
			&order.ID.Value,
			&order.Status,
			&order.CreatedAt,
			&order.UpdatedAt.Time,
		)
		if err != nil {
			span.AddAttributes(ctx, o11y.Error, "error scan row", o11y.Attributes{Key: "error", Value: err})
			return nil, err
		}
		orders = append(orders, &order)
	}
	return orders, nil
}

func (r *orderRepository) Find(ctx context.Context, orderID sharedVos.UUID) (*entities.Order, error) {
	ctx, span := r.o11y.Start(ctx, "order_repository.find")
	defer span.End()

	query := `select
				id,
				status,
				created_at,
				updated_at
			  from
				orders
			  where
				id = $1`

	var order entities.Order
	err := r.tx.QueryRowContext(ctx, query, orderID.String()).Scan(
		&order.ID.Value,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt.Time,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			span.AddAttributes(ctx, o11y.Ok, "order found", o11y.Attributes{Key: "order_id", Value: orderID.String()})
			return nil, nil
		}
		span.AddAttributes(ctx, o11y.Error, "error find order", o11y.Attributes{Key: "order_id", Value: orderID.String()})
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) Insert(ctx context.Context, order *entities.Order) error {
	ctx, span := r.o11y.Start(ctx, "order_repository.insert")
	defer span.End()

	query := `insert into
				orders (id, status, created_at, updated_at)
			  values
				($1, $2, $3, $4)`

	_, err := r.tx.ExecContext(
		ctx,
		query,
		order.ID.Value,
		order.Status.String(),
		order.CreatedAt,
		order.UpdatedAt.Time,
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

	query := `insert into
				order_items (
					id,
					order_id,
					product_name,
					quantity,
					price,
					created_at,
					updated_at
					)
				values
					($1, $2, $3, $4, $5, $6, $7)`

	for _, item := range items {
		_, err := r.tx.ExecContext(
			ctx,
			query,
			item.ID.Value,
			item.OrderID.Value,
			item.ProductName,
			item.Quantity,
			item.Price,
			item.CreatedAt,
			item.UpdatedAt.Time,
		)
		if err != nil {
			span.AddAttributes(ctx, o11y.Error, "error insert order item", o11y.Attributes{Key: "error", Value: err})
			return err
		}
	}

	return nil
}

func (r *orderRepository) Update(ctx context.Context, order *entities.Order) error {
	ctx, span := r.o11y.Start(ctx, "order_repository.update")
	defer span.End()

	query := `update
				orders
			  set
				status = $1,
				updated_at = $2
			  where
				id = $3`

	_, err := r.tx.ExecContext(
		ctx,
		query,
		order.Status.String(),
		order.UpdatedAt.Time,
		order.ID.Value,
	)
	if err != nil {
		span.AddAttributes(ctx, o11y.Error, "error update order", o11y.Attributes{Key: "error", Value: err})
		return err
	}
	return nil
}
