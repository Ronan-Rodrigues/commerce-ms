package postgres

import (
	"context"
	"errors"

	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/catalog-service/internal/domain/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

type PostgresCategoryRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresCategoryRepository(pool *pgxpool.Pool) *PostgresCategoryRepository {
	return &PostgresCategoryRepository{pool: pool}
}

func (r *PostgresCategoryRepository) Create(ctx context.Context, c *entity.Category) error {
	query := `
		INSERT INTO catalog.categories (id, name, slug, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6);
	`
	_, err := r.pool.Exec(ctx, query, c.ID, c.Name, c.Slug, c.Description, c.CreatedAt, c.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return repository.ErrCategoryAlreadyExists
		}
		return err
	}
	return nil
}

func (r *PostgresCategoryRepository) FindByID(ctx context.Context, id string) (*entity.Category, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM catalog.categories
		WHERE id = $1
		LIMIT 1;
	`
	var c entity.Category
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrCategoryNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *PostgresCategoryRepository) FindBySlug(ctx context.Context, slug string) (*entity.Category, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM catalog.categories
		WHERE slug = $1
		LIMIT 1;
	`
	var c entity.Category
	err := r.pool.QueryRow(ctx, query, slug).Scan(
		&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt, &c.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrCategoryNotFound
		}
		return nil, err
	}
	return &c, nil
}

func (r *PostgresCategoryRepository) ListAll(ctx context.Context) ([]*entity.Category, error) {
	query := `
		SELECT id, name, slug, description, created_at, updated_at
		FROM catalog.categories
		ORDER BY name ASC;
	`
	rows, err := r.pool.Query(ctx, query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var categories []*entity.Category
	for rows.Next() {
		var c entity.Category
		if err := rows.Scan(&c.ID, &c.Name, &c.Slug, &c.Description, &c.CreatedAt, &c.UpdatedAt); err != nil {
			return nil, err
		}
		categories = append(categories, &c)
	}
	return categories, nil
}
