package postgres

import (
	"context"

	"github.com/Ronan-Rodrigues/commerce-ms/services/payment-service/internal/domain/entity"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresTransactionRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresTransactionRepository(pool *pgxpool.Pool) *PostgresTransactionRepository {
	return &PostgresTransactionRepository{pool: pool}
}

func (r *PostgresTransactionRepository) Create(ctx context.Context, tx *entity.Transaction) error {
	query := `
		INSERT INTO payments.transactions (id, order_id, amount, status, payment_method, gateway_transaction_id, created_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`
	_, err := r.pool.Exec(
		ctx, query,
		tx.ID, tx.OrderID, tx.Amount, string(tx.Status), tx.PaymentMethod, tx.GatewayTransactionID, tx.CreatedAt,
	)
	return err
}

func (r *PostgresTransactionRepository) FindByOrderID(ctx context.Context, orderID string) ([]*entity.Transaction, error) {
	query := `
		SELECT id, order_id, amount, status, payment_method, gateway_transaction_id, created_at
		FROM payments.transactions
		WHERE order_id = $1
		ORDER BY created_at DESC;
	`
	rows, err := r.pool.Query(ctx, query, orderID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Transaction
	for rows.Next() {
		var tx entity.Transaction
		var statusStr string
		if err := rows.Scan(
			&tx.ID, &tx.OrderID, &tx.Amount, &statusStr, &tx.PaymentMethod, &tx.GatewayTransactionID, &tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		tx.Status = entity.TransactionStatus(statusStr)
		list = append(list, &tx)
	}
	return list, nil
}

func (r *PostgresTransactionRepository) GetFinancialSummary(ctx context.Context) (*entity.FinancialSummary, error) {
	query := `
		SELECT
			COALESCE(SUM(CASE WHEN status = 'succeeded' THEN amount ELSE 0 END), 0) AS total_revenue,
			COUNT(*) AS total_count,
			COUNT(CASE WHEN status = 'succeeded' THEN 1 END) AS success_count,
			COUNT(CASE WHEN status = 'failed' THEN 1 END) AS failed_count,
			COALESCE(ROUND(AVG(CASE WHEN status = 'succeeded' THEN amount ELSE NULL END)), 0) AS avg_ticket
		FROM payments.transactions;
	`
	var s entity.FinancialSummary
	var avgTicket float64
	err := r.pool.QueryRow(ctx, query).Scan(
		&s.TotalRevenue,
		&s.TotalCount,
		&s.SuccessCount,
		&s.FailedCount,
		&avgTicket,
	)
	if err != nil {
		return nil, err
	}
	s.AverageTicket = int64(avgTicket)
	return &s, nil
}

func (r *PostgresTransactionRepository) ListTransactions(ctx context.Context, limit, offset int) ([]*entity.Transaction, error) {
	query := `
		SELECT id, order_id, amount, status, payment_method, gateway_transaction_id, created_at
		FROM payments.transactions
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2;
	`
	rows, err := r.pool.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []*entity.Transaction
	for rows.Next() {
		var tx entity.Transaction
		var statusStr string
		if err := rows.Scan(
			&tx.ID, &tx.OrderID, &tx.Amount, &statusStr, &tx.PaymentMethod, &tx.GatewayTransactionID, &tx.CreatedAt,
		); err != nil {
			return nil, err
		}
		tx.Status = entity.TransactionStatus(statusStr)
		list = append(list, &tx)
	}
	return list, nil
}
