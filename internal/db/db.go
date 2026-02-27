package db

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Queries struct {
	pool *pgxpool.Pool
	tx   pgx.Tx
}

func New(pool *pgxpool.Pool) *Queries {
	return &Queries{pool: pool}
}

func (q *Queries) WithTx(tx pgx.Tx) *Queries {
	return &Queries{pool: q.pool, tx: tx}
}

func (q *Queries) queryRow(ctx context.Context, sql string, args ...interface{}) pgx.Row {
	if q.tx != nil {
		return q.tx.QueryRow(ctx, sql, args...)
	}
	return q.pool.QueryRow(ctx, sql, args...)
}

func (q *Queries) query(ctx context.Context, sql string, args ...interface{}) (pgx.Rows, error) {
	if q.tx != nil {
		return q.tx.Query(ctx, sql, args...)
	}
	return q.pool.Query(ctx, sql, args...)
}

func (q *Queries) exec(ctx context.Context, sql string, args ...interface{}) error {
	if q.tx != nil {
		_, err := q.tx.Exec(ctx, sql, args...)
		return err
	}
	_, err := q.pool.Exec(ctx, sql, args...)
	return err
}
