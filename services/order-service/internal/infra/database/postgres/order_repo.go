package postgres

import (
	"context"
	"errors"

	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/order-service/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresOrderRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresOrderRepository(pool *pgxpool.Pool) *PostgresOrderRepository {
	return &PostgresOrderRepository{pool: pool}
}

func (r *PostgresOrderRepository) Create(ctx context.Context, o *entity.Order) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	orderQuery := `
		INSERT INTO orders.orders (id, user_id, total_amount, status, idempotency_key, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`
	_, err = tx.Exec(ctx, orderQuery, o.ID, o.UserID, o.TotalAmount, string(o.Status), o.IdempotencyKey, o.CreatedAt, o.UpdatedAt)
	if err != nil {
		return err
	}

	itemQuery := `
		INSERT INTO orders.order_items (id, order_id, product_id, name, price, quantity, subtotal, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8);
	`
	for _, item := range o.Items {
		itemID := uuid.NewString()
		_, err := tx.Exec(ctx, itemQuery, itemID, o.ID, item.ProductID, item.Name, item.Price, item.Quantity, item.Subtotal, o.CreatedAt)
		if err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func (r *PostgresOrderRepository) FindByIdempotencyKey(ctx context.Context, key string) (*entity.Order, error) {
	query := `
		SELECT id, user_id, total_amount, status, idempotency_key, created_at, updated_at
		FROM orders.orders
		WHERE idempotency_key = $1
		LIMIT 1;
	`
	var o entity.Order
	var statusStr string
	err := r.pool.QueryRow(ctx, query, key).Scan(
		&o.ID, &o.UserID, &o.TotalAmount, &statusStr, &o.IdempotencyKey, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrOrderNotFound
		}
		return nil, err
	}
	o.Status = entity.OrderStatus(statusStr)
	return &o, nil
}

func (r *PostgresOrderRepository) FindByID(ctx context.Context, id string) (*entity.Order, error) {
	query := `
		SELECT id, user_id, total_amount, status, idempotency_key, created_at, updated_at
		FROM orders.orders
		WHERE id = $1
		LIMIT 1;
	`
	var o entity.Order
	var statusStr string
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&o.ID, &o.UserID, &o.TotalAmount, &statusStr, &o.IdempotencyKey, &o.CreatedAt, &o.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrOrderNotFound
		}
		return nil, err
	}
	o.Status = entity.OrderStatus(statusStr)

	// Busca os itens do pedido
	itemsQuery := `
		SELECT id, product_id, name, price, quantity, subtotal
		FROM orders.order_items
		WHERE order_id = $1;
	`
	rows, err := r.pool.Query(ctx, itemsQuery, id)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var item entity.OrderItem
		if err := rows.Scan(&item.ID, &item.ProductID, &item.Name, &item.Price, &item.Quantity, &item.Subtotal); err != nil {
			return nil, err
		}
		o.Items = append(o.Items, &item)
	}

	return &o, nil
}

func (r *PostgresOrderRepository) ListByUserID(ctx context.Context, userID string) ([]*entity.Order, error) {
	query := `
		SELECT id, user_id, total_amount, status, idempotency_key, created_at, updated_at
		FROM orders.orders
		WHERE user_id = $1
		ORDER BY created_at DESC;
	`
	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Order
	for rows.Next() {
		var o entity.Order
		var statusStr string
		if err := rows.Scan(&o.ID, &o.UserID, &o.TotalAmount, &statusStr, &o.IdempotencyKey, &o.CreatedAt, &o.UpdatedAt); err != nil {
			return nil, err
		}
		o.Status = entity.OrderStatus(statusStr)
		list = append(list, &o)
	}
	return list, nil
}

func (r *PostgresOrderRepository) UpdateStatus(ctx context.Context, id string, status entity.OrderStatus) error {
	query := `
		UPDATE orders.orders
		SET status = $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1;
	`
	tag, err := r.pool.Exec(ctx, query, id, string(status))
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return repository.ErrOrderNotFound
	}
	return nil
}
