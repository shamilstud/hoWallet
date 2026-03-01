package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	db "github.com/howallet/howallet/internal/db"
	"github.com/howallet/howallet/internal/repository"
)

type refreshTokenRepo struct {
	queries *db.Queries
}

func (r *refreshTokenRepo) Create(ctx context.Context, userID uuid.UUID, tokenHash string, expiresAt time.Time) error {
	return r.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: tokenHash,
		ExpiresAt: expiresAt,
	})
}

func (r *refreshTokenRepo) GetByHash(ctx context.Context, tokenHash string) (repository.RefreshTokenRow, error) {
	rt, err := r.queries.GetRefreshToken(ctx, tokenHash)
	if err != nil {
		return repository.RefreshTokenRow{}, err
	}
	return repository.RefreshTokenRow{
		ID:        rt.ID,
		UserID:    rt.UserID,
		TokenHash: rt.TokenHash,
		ExpiresAt: rt.ExpiresAt.Time,
		CreatedAt: rt.CreatedAt.Time,
	}, nil
}

func (r *refreshTokenRepo) Delete(ctx context.Context, tokenHash string) error {
	return r.queries.DeleteRefreshToken(ctx, tokenHash)
}

func (r *refreshTokenRepo) DeleteByUser(ctx context.Context, userID uuid.UUID) error {
	return r.queries.DeleteUserRefreshTokens(ctx, userID)
}

func (r *refreshTokenRepo) DeleteExpired(ctx context.Context) error {
	return r.queries.DeleteExpiredRefreshTokens(ctx)
}
