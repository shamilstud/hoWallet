package repository

import "context"

// TxFunc is a function executed within a transaction.
type TxFunc func(ctx context.Context) error

// UnitOfWork provides transactional execution support.
type UnitOfWork interface {
	// RunInTx executes fn inside a database transaction.
	// If fn returns an error the transaction is rolled back, otherwise committed.
	RunInTx(ctx context.Context, fn TxFunc) error
}
