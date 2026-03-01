package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
)

// RefreshTokenRepository defines data access for refresh tokens.
type RefreshTokenRepository interface {
	Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error
	GetByHash(ctx context.Context, tokenHash string) (RefreshTokenRow, error)
	Delete(ctx context.Context, tokenHash string) error
	DeleteByUser(ctx context.Context, userID uuid.UUID) error
	DeleteExpired(ctx context.Context) error
}

// RefreshTokenRow holds the data returned when querying a refresh token.
type RefreshTokenRow struct {
	ID        uuid.UUID
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
	CreatedAt time.Time
}
