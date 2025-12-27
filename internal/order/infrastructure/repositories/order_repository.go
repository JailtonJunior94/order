package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/jailtonjunior94/order/internal/order/domain/entities"
	"github.com/jailtonjunior94/order/internal/order/domain/interfaces"
	"github.com/jailtonjunior94/order/internal/order/domain/vos"
	"github.com/jailtonjunior94/order/pkg/database"
	"github.com/jailtonjunior94/order/pkg/observability"
	sharedVos "github.com/jailtonjunior94/order/pkg/vos"
)

type orderRepository struct {
	db        database.DBTX
	o11y observability.Observability
}

func NewOrderRepository(db database.DBTX, o11y observability.Observability) interfaces.OrderRepository {
	return &orderRepository{
		db:        db,
		o11y: o11y,
	}
}

func (r *orderRepository) FindAll(ctx context.Context, status vos.Status) ([]*entities.Order, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "order_repository.find_all")
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

	rows, err := r.db.QueryContext(ctx, query, status.String())
	if err != nil {
		span.AddEvent("error find all orders", observability.Any("error", err))
		return nil, err
	}
	defer func() {
		if err := rows.Close(); err != nil {
			span.AddEvent("error closing rows", observability.Any("error", err))
		}
	}()

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
			span.AddEvent("error scan row", observability.Any("error", err))
			return nil, err
		}
		orders = append(orders, &order)
	}
	return orders, nil
}

func (r *orderRepository) Find(ctx context.Context, orderID sharedVos.UUID) (*entities.Order, error) {
	ctx, span := r.o11y.Tracer().Start(ctx, "order_repository.find")
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
	err := r.db.QueryRowContext(ctx, query, orderID.String()).Scan(
		&order.ID.Value,
		&order.Status,
		&order.CreatedAt,
		&order.UpdatedAt.Time,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			span.AddEvent("order found", observability.Any("order_id", orderID.String()))
			return nil, nil
		}
		span.AddEvent("error find order", observability.Any("order_id", orderID.String()))
		return nil, err
	}
	return &order, nil
}

func (r *orderRepository) Insert(ctx context.Context, order *entities.Order) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "order_repository.insert")
	defer span.End()

	query := `insert into
				orders (id, status, created_at, updated_at)
			  values
				($1, $2, $3, $4)`

	_, err := r.db.ExecContext(
		ctx,
		query,
		order.ID.Value,
		order.Status.String(),
		order.CreatedAt,
		order.UpdatedAt.Time,
	)
	if err != nil {
		span.AddEvent("error insert order", observability.Any("error", err))
		return err
	}
	return nil
}

func (r *orderRepository) InsertItems(ctx context.Context, items []*entities.OrderItem) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "order_repository.insert_items")
	defer span.End()

	if len(items) == 0 {
		return nil
	}

	// Build batch insert query with placeholders
	// Single query: INSERT INTO order_items (...) VALUES ($1,$2,...),($8,$9,...),(...)
	query := `insert into order_items (id, order_id, product_name, quantity, price, created_at, updated_at) values `

	// Create placeholders and args for batch insert
	placeholders := make([]string, 0, len(items))
	args := make([]interface{}, 0, len(items)*7)

	for i, item := range items {
		offset := i * 7
		placeholders = append(placeholders, fmt.Sprintf("($%d, $%d, $%d, $%d, $%d, $%d, $%d)",
			offset+1, offset+2, offset+3, offset+4, offset+5, offset+6, offset+7))

		args = append(args,
			item.ID.Value,
			item.OrderID.Value,
			item.ProductName,
			item.Quantity,
			item.Price,
			item.CreatedAt,
			item.UpdatedAt.Time,
		)
	}

	// Execute batch insert
	query += strings.Join(placeholders, ", ")

	if _, err := r.db.ExecContext(ctx, query, args...); err != nil {
		span.AddEvent("error batch insert order items", observability.Any("error", err))
		return err
	}

	span.AddEvent("batch insert successful", observability.Any("items_count", len(items)))
	return nil
}

func (r *orderRepository) Update(ctx context.Context, order *entities.Order) error {
	ctx, span := r.o11y.Tracer().Start(ctx, "order_repository.update")
	defer span.End()

	query := `update
				orders
			  set
				status = $1,
				updated_at = $2
			  where
				id = $3`

	_, err := r.db.ExecContext(
		ctx,
		query,
		order.Status.String(),
		order.UpdatedAt.Time,
		order.ID.Value,
	)
	if err != nil {
		span.AddEvent("error update order", observability.Any("error", err))
		return err
	}
	return nil
}
