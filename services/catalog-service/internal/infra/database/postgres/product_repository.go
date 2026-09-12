package postgres

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresProductRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresProductRepository(pool *pgxpool.Pool) *PostgresProductRepository {
	return &PostgresProductRepository{pool: pool}
}

func (r *PostgresProductRepository) Create(ctx context.Context, p *entity.Product) error {
	query := `
		INSERT INTO catalog.products (id, name, slug, description, price, stock, sku, category_id, image_url, active, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12);
	`
	_, err := r.pool.Exec(
		ctx, query,
		p.ID, p.Name, p.Slug, p.Description, p.Price, p.Stock, p.SKU,
		p.CategoryID, p.ImageURL, p.Active, p.CreatedAt, p.UpdatedAt,
	)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrProductAlreadyExists
		}
		return err
	}
	return nil
}

func (r *PostgresProductRepository) FindByID(ctx context.Context, id string) (*entity.Product, error) {
	query := `
		SELECT id, name, slug, description, price, stock, sku, COALESCE(category_id, ''), image_url, active, created_at, updated_at
		FROM catalog.products
		WHERE id = $1
		LIMIT 1;
	`
	var p entity.Product
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.Stock, &p.SKU,
		&p.CategoryID, &p.ImageURL, &p.Active, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrProductNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *PostgresProductRepository) FindBySlug(ctx context.Context, slug string) (*entity.Product, error) {
	query := `
		SELECT id, name, slug, description, price, stock, sku, COALESCE(category_id, ''), image_url, active, created_at, updated_at
		FROM catalog.products
		WHERE slug = $1
		LIMIT 1;
	`
	var p entity.Product
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.Stock, &p.SKU,
		&p.CategoryID, &p.ImageURL, &p.Active, &p.CreatedAt, &p.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrProductNotFound
		}
		return nil, err
	}
	return &p, nil
}

func (r *PostgresProductRepository) List(ctx context.Context, filter repository.ProductFilter) ([]*entity.Product, int, error) {
	var conditions []string
	var args []interface{}
	argIdx := 1

	if filter.ActiveOnly {
		conditions = append(conditions, fmt.Sprintf("active = $%d", argIdx))
		args = append(args, true)
		argIdx++
	}

	if filter.CategoryID != "" {
		conditions = append(conditions, fmt.Sprintf("category_id = $%d", argIdx))
		args = append(args, filter.CategoryID)
		argIdx++
	}

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(LOWER(name) LIKE $%d OR LOWER(description) LIKE $%d)", argIdx, argIdx))
		args = append(args, "%"+strings.ToLower(filter.Search)+"%")
		argIdx++
	}

	if filter.MinPrice > 0 {
		conditions = append(conditions, fmt.Sprintf("price >= $%d", argIdx))
		args = append(args, filter.MinPrice)
		argIdx++
	}

	if filter.MaxPrice > 0 {
		conditions = append(conditions, fmt.Sprintf("price <= $%d", argIdx))
		args = append(args, filter.MaxPrice)
		argIdx++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Contagem total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM catalog.products %s;", whereClause)
	var total int
	if err := r.pool.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, err
	}

	// Consulta paginada
	query := fmt.Sprintf(`
		SELECT id, name, slug, description, price, stock, sku, COALESCE(category_id, ''), image_url, active, created_at, updated_at
		FROM catalog.products
		%s
		ORDER BY created_at DESC
		LIMIT $%d OFFSET $%d;
	`, whereClause, argIdx, argIdx+1)

	args = append(args, filter.Limit, filter.Offset)

	rows, err := r.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var products []*entity.Product
	for rows.Next() {
		var p entity.Product
		if err := rows.Scan(
			&p.ID, &p.Name, &p.Slug, &p.Description, &p.Price, &p.Stock, &p.SKU,
			&p.CategoryID, &p.ImageURL, &p.Active, &p.CreatedAt, &p.UpdatedAt,
		); err != nil {
			return nil, 0, err
		}
		products = append(products, &p)
	}

	return products, total, nil
}

func (r *PostgresProductRepository) Update(ctx context.Context, p *entity.Product) error {
	query := `
		UPDATE catalog.products
		SET name = $2, slug = $3, description = $4, price = $5, stock = $6, sku = $7,
		    category_id = $8, image_url = $9, active = $10, updated_at = $11
		WHERE id = $1;
	`
	_, err := r.pool.Exec(
		ctx, query,
		p.ID, p.Name, p.Slug, p.Description, p.Price, p.Stock, p.SKU,
		p.CategoryID, p.ImageURL, p.Active, p.UpdatedAt,
	)
	return err
}

// DeductStockAtomic deduz quantidade no estoque de forma atômica no banco (proteção contra concorrência)
func (r *PostgresProductRepository) DeductStockAtomic(ctx context.Context, id string, quantity int) error {
	query := `
		UPDATE catalog.products
		SET stock = stock - $2, updated_at = CURRENT_TIMESTAMP
		WHERE id = $1 AND stock >= $2;
	`
	tag, err := r.pool.Exec(ctx, query, id, quantity)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return errors.New("estoque insuficiente para dedução atômica")
	}
	return nil
}
