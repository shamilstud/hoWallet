package db

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CreateRefreshTokenParams struct {
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
}

func (q *Queries) CreateRefreshToken(ctx context.Context, arg CreateRefreshTokenParams) error {
	return q.exec(ctx,
		`INSERT INTO refresh_tokens (user_id, token_hash, expires_at) VALUES ($1, $2, $3)`,
		arg.UserID, arg.TokenHash, arg.ExpiresAt,
	)
}

func (q *Queries) GetRefreshToken(ctx context.Context, tokenHash string) (RefreshToken, error) {
	row := q.queryRow(ctx,
		`SELECT id, user_id, token_hash, expires_at, created_at FROM refresh_tokens WHERE token_hash = $1`,
		tokenHash,
	)
	var rt RefreshToken
	err := row.Scan(&rt.ID, &rt.UserID, &rt.TokenHash, &rt.ExpiresAt, &rt.CreatedAt)
	return rt, err
}

func (q *Queries) DeleteRefreshToken(ctx context.Context, tokenHash string) error {
	return q.exec(ctx, `DELETE FROM refresh_tokens WHERE token_hash = $1`, tokenHash)
}

func (q *Queries) DeleteUserRefreshTokens(ctx context.Context, userID uuid.UUID) error {
	return q.exec(ctx, `DELETE FROM refresh_tokens WHERE user_id = $1`, userID)
}

func (q *Queries) DeleteExpiredRefreshTokens(ctx context.Context) error {
	return q.exec(ctx, `DELETE FROM refresh_tokens WHERE expires_at < now()`)
}
