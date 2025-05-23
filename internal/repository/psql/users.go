package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"hippo/internal/domain"
	"hippo/internal/repository"
)

type Users struct {
	db *sql.DB
}

func NewUsers(db *sql.DB) *Users {
	return &Users{db: db}
}

func (r *Users) Create(ctx context.Context, user domain.User) error {
	const op = "repository.psql.users.Create"
	const query = `
		INSERT INTO users (name, email, password, registered_at) 
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (email) DO NOTHING
		RETURNING id
	`

	var id int64
	err := r.db.QueryRowContext(ctx, query,
		user.Name,
		user.Email,
		user.Password,
		user.CreatedAt,
	).Scan(&id)

	if errors.Is(err, sql.ErrNoRows) {
		return repository.NewErrDuplicateEmail(err)
	}

	if err != nil {
		return fmt.Errorf("%s: failed to create user: %w", op, err)
	}

	return nil
}

func (r *Users) GetByEmail(ctx context.Context, email string) (domain.User, error) {
	const op = "repository.psql.users.GetByEmail"
	const query = `
		SELECT id, name, email, password, registered_at 
		FROM users
		WHERE email = $1
	`

	var user domain.User
	err := r.db.QueryRowContext(ctx, query, email).Scan(
		&user.ID,
		&user.Name,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
	)

	switch {
	case errors.Is(err, sql.ErrNoRows):
		return domain.User{}, repository.NewErrInvalidCredential()
	case err != nil:
		return domain.User{}, fmt.Errorf("%s: failed to get user by credentials: %w", op, err)
	default:
		return user, nil
	}
}
