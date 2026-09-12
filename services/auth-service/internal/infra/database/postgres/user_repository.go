package postgres

import (
	"context"
	"errors"
	"strings"

	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/entity"
	"github.com/Ronan-Rodrigues/commerce-ms/services/auth-service/internal/domain/repository"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// PostgresUserRepository implementa a interface repository.UserRepository
type PostgresUserRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{pool: pool}
}

// Create insere um novo usuário na tabela auth.users
func (r *PostgresUserRepository) Create(ctx context.Context, user *entity.User) error {
	query := `
		INSERT INTO auth.users (id, name, email, password_hash, role, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7);
	`

	_, err := r.pool.Exec(
		ctx,
		query,
		user.ID,
		user.Name,
		strings.ToLower(user.Email),
		user.PasswordHash,
		user.Role,
		user.CreatedAt,
		user.UpdatedAt,
	)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" { // 23505 = unique_violation
			return repository.ErrUserAlreadyExists
		}
		return err
	}

	return nil
}

// FindByEmail busca o usuário pelo email
func (r *PostgresUserRepository) FindByEmail(ctx context.Context, email string) (*entity.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, created_at, updated_at
		FROM auth.users
		WHERE LOWER(email) = LOWER($1)
		LIMIT 1;
	`

	var u entity.User
	err := r.pool.QueryRow(ctx, query, strings.ToLower(email)).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}

// FindByID busca o usuário pelo ID
func (r *PostgresUserRepository) FindByID(ctx context.Context, id string) (*entity.User, error) {
	query := `
		SELECT id, name, email, password_hash, role, created_at, updated_at
		FROM auth.users
		WHERE id = $1
		LIMIT 1;
	`

	var u entity.User
	err := r.pool.QueryRow(ctx, query, id).Scan(
		&u.ID,
		&u.Name,
		&u.Email,
		&u.PasswordHash,
		&u.Role,
		&u.CreatedAt,
		&u.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, repository.ErrUserNotFound
		}
		return nil, err
	}

	return &u, nil
}
