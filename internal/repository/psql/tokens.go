package psql

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"hippo/internal/domain"
	"hippo/internal/repository"
)

type Token struct {
	db *sql.DB
}

func NewToken(db *sql.DB) *Token {
	return &Token{
		db: db,
	}
}

func (t *Token) Create(ctx context.Context, token domain.RefreshSession) error {
	const op = "psql.tokens.Create"

	query := `
		INSERT INTO refresh_tokens (user_id, token, expires_at)
		VALUES ($1, $2, $3)
	`

	_, err := t.db.ExecContext(ctx, query, token.UserID, token.Token, token.ExpiresAt)
	if err != nil {
		return fmt.Errorf("%s: failed to insert token: %w", op, err)
	}

	return nil
}

func (t *Token) Get(ctx context.Context, token string) (domain.RefreshSession, error) {
	const op = "psql.tokens.Get"

	tx, err := t.db.BeginTx(ctx, nil)
	if err != nil {
		return domain.RefreshSession{}, fmt.Errorf("%s: begin tx: %w", op, err)
	}
	defer func() {
		_ = tx.Rollback()
	}()

	var session domain.RefreshSession
	query := `SELECT id, user_id, token, expires_at FROM refresh_tokens WHERE token = $1`

	err = tx.QueryRowContext(ctx, query, token).Scan(
		&session.ID, &session.UserID, &session.Token, &session.ExpiresAt,
	)

	if errors.Is(err, sql.ErrNoRows) {
		return domain.RefreshSession{}, repository.NewErrTokenNotFound()
	}

	if err != nil {
		return domain.RefreshSession{}, fmt.Errorf("%s: select failed: %w", op, err)
	}

	delQuery := `DELETE FROM refresh_tokens WHERE token = $1`
	if _, err := tx.ExecContext(ctx, delQuery, token); err != nil {
		return domain.RefreshSession{}, fmt.Errorf("%s: delete failed: %w", op, err)
	}

	if err := tx.Commit(); err != nil {
		return domain.RefreshSession{}, fmt.Errorf("%s: commit failed: %w", op, err)
	}

	return session, nil
}
