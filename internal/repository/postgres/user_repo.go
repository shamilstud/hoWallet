package postgres

import (
	"context"

	"github.com/google/uuid"
	db "github.com/howallet/howallet/internal/db"
	"github.com/howallet/howallet/internal/model"
)

type userRepo struct {
	queries *db.Queries
}

func (r *userRepo) Create(ctx context.Context, email, passwordHash, name string) (model.User, error) {
	u, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        email,
		PasswordHash: passwordHash,
		Name:         name,
	})
	if err != nil {
		return model.User{}, err
	}
	return toUserModel(u), nil
}

func (r *userRepo) GetByID(ctx context.Context, id uuid.UUID) (model.User, error) {
	u, err := r.queries.GetUserByID(ctx, id)
	if err != nil {
		return model.User{}, err
	}
	return toUserModel(u), nil
}

func (r *userRepo) GetByEmail(ctx context.Context, email string) (model.User, error) {
	u, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		return model.User{}, err
	}
	return toUserModel(u), nil
}

func toUserModel(u db.User) model.User {
	return model.User{
		ID:           u.ID,
		Email:        u.Email,
		PasswordHash: u.PasswordHash,
		Name:         u.Name,
		CreatedAt:    u.CreatedAt.Time,
		UpdatedAt:    u.UpdatedAt.Time,
	}
}
