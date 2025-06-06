package repository

import (
	"context"
	"database/sql"
	"fmt"
)

// TokenRepository определяет интерфейс для работы с токенами
type TokenRepository interface {
	RevokeToken(ctx context.Context, token, userID string) error
	IsTokenRevoked(ctx context.Context, token string) (bool, error)
}

type tokenRepository struct {
	db *sql.DB
}

func NewTokenRepository(db *sql.DB) TokenRepository {
	return &tokenRepository{db: db}
}

func (r *tokenRepository) RevokeToken(ctx context.Context, token, userID string) error {
	//noinspection SqlNoDataSourceInspection
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO revoked_tokens (token, id_user) VALUES ($1, $2)`,
		token,
		userID,
	)
	if err != nil {
		return fmt.Errorf("failed to revoke token: %w", err)
	}
	return nil
}

func (r *tokenRepository) IsTokenRevoked(ctx context.Context, token string) (bool, error) {
	var exists bool
	//noinspection SqlNoDataSourceInspection
	err := r.db.QueryRowContext(ctx,
		`SELECT EXISTS(SELECT 1 FROM revoked_tokens WHERE token = $1)`,
		token,
	).Scan(&exists)
	if err != nil {
		return false, fmt.Errorf("failed to check token revocation: %w", err)
	}
	return exists, nil
}
