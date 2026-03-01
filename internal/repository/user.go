package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/howallet/howallet/internal/model"
)

// UserRepository defines data access for users.
type UserRepository interface {
	Create(ctx context.Context, email, passwordHash, name string) (model.User, error)
	GetByID(ctx context.Context, id uuid.UUID) (model.User, error)
	GetByEmail(ctx context.Context, email string) (model.User, error)
}
