package postgres

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/howallet/howallet/internal/db"
	"github.com/howallet/howallet/internal/repository"
)

// Repos groups all postgres repository implementations.
type Repos struct {
	pool    *pgxpool.Pool
	queries *db.Queries

	Users         repository.UserRepository
	Accounts      repository.AccountRepository
	Transactions  repository.TransactionRepository
	Households    repository.HouseholdRepository
	Invitations   repository.InvitationRepository
	RefreshTokens repository.RefreshTokenRepository
}

// New creates all postgres repositories from a connection pool.
func New(pool *pgxpool.Pool) *Repos {
	queries := db.New(pool)
	r := &Repos{pool: pool, queries: queries}

	r.Users = &userRepo{queries: queries}
	r.Accounts = &accountRepo{queries: queries}
	r.Transactions = &transactionRepo{queries: queries}
	r.Households = &householdRepo{queries: queries}
	r.Invitations = &invitationRepo{queries: queries}
	r.RefreshTokens = &refreshTokenRepo{queries: queries}

	return r
}

// RunInTx executes fn inside a database transaction.
func (r *Repos) RunInTx(ctx context.Context, fn repository.TxFunc) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Create a child Repos with transactional queries
	qtx := r.queries.WithTx(tx)
	txRepos := &Repos{pool: r.pool, queries: qtx}
	txRepos.Users = &userRepo{queries: qtx}
	txRepos.Accounts = &accountRepo{queries: qtx}
	txRepos.Transactions = &transactionRepo{queries: qtx}
	txRepos.Households = &householdRepo{queries: qtx}
	txRepos.Invitations = &invitationRepo{queries: qtx}
	txRepos.RefreshTokens = &refreshTokenRepo{queries: qtx}

	// Store transactional repos in context so services can access them
	ctx = WithTxRepos(ctx, txRepos)
	if err := fn(ctx); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

// --- context helpers for transactional repos ---

type txReposKey struct{}

// WithTxRepos stores transactional repos in context.
func WithTxRepos(ctx context.Context, repos *Repos) context.Context {
	return context.WithValue(ctx, txReposKey{}, repos)
}

// TxReposFromCtx returns the transactional repos from context, or nil.
func TxReposFromCtx(ctx context.Context) *Repos {
	if v, ok := ctx.Value(txReposKey{}).(*Repos); ok {
		return v
	}
	return nil
}
