package db

import (
	"context"

	"github.com/google/uuid"
)

// --- Users ---

type CreateUserParams struct {
	Email        string
	PasswordHash string
	Name         string
}

func (q *Queries) CreateUser(ctx context.Context, arg CreateUserParams) (User, error) {
	row := q.queryRow(ctx,
		`INSERT INTO users (email, password_hash, name) VALUES ($1, $2, $3) RETURNING id, email, password_hash, name, created_at, updated_at`,
		arg.Email, arg.PasswordHash, arg.Name,
	)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (q *Queries) GetUserByEmail(ctx context.Context, email string) (User, error) {
	row := q.queryRow(ctx,
		`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE email = $1`,
		email,
	)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}

func (q *Queries) GetUserByID(ctx context.Context, id uuid.UUID) (User, error) {
	row := q.queryRow(ctx,
		`SELECT id, email, password_hash, name, created_at, updated_at FROM users WHERE id = $1`,
		id,
	)
	var u User
	err := row.Scan(&u.ID, &u.Email, &u.PasswordHash, &u.Name, &u.CreatedAt, &u.UpdatedAt)
	return u, err
}
