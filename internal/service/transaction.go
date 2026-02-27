package service

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	db "github.com/howallet/howallet/internal/db"
	"github.com/howallet/howallet/internal/model"
)

var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrTransferMissingDest = errors.New("transfer requires destination_account_id")
)

type TransactionService struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewTransactionService(pool *pgxpool.Pool, queries *db.Queries) *TransactionService {
	return &TransactionService{queries: queries, pool: pool}
}

// Create creates a transaction and updates account balances atomically.
func (s *TransactionService) Create(ctx context.Context, householdID, userID uuid.UUID, req model.CreateTransactionRequest) (*model.Transaction, error) {
	amount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	if req.Type == model.TransactionTypeTransfer && req.DestinationAccountID == nil {
		return nil, ErrTransferMissingDest
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	var destID *uuid.UUID
	if req.DestinationAccountID != nil {
		destID = req.DestinationAccountID
	}

	params := db.CreateTransactionParams{
		HouseholdID: householdID,
		Type:        db.TransactionType(req.Type),
		Description: req.Description,
		Amount:      amount,
		AccountID:   req.AccountID,
		Tags:        req.Tags,
		Note:        toPgText(req.Note),
		TransactedAt: pgtype.Timestamptz{
			Time:  req.TransactedAt,
			Valid: true,
		},
		CreatedBy: userID,
	}
	if destID != nil {
		params.DestinationAccountID = toNullUUID(destID)
	}

	dbTxn, err := qtx.CreateTransaction(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("create transaction: %w", err)
	}

	// Update balances
	if err := s.applyBalanceChange(ctx, qtx, req.Type, amount, req.AccountID, destID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	txn := toTransactionModel(dbTxn)
	return &txn, nil
}

// List returns paginated transactions with filters.
func (s *TransactionService) List(ctx context.Context, householdID uuid.UUID, q model.ListTransactionsQuery) (*model.PaginatedResponse, error) {
	if q.Limit <= 0 {
		q.Limit = 50
	}

	params := db.ListTransactionsParams{
		HouseholdID: householdID,
		Column2:     toPgTimestamptz(q.From),
		Column3:     toPgTimestamptz(q.To),
		Limit:       q.Limit,
		Offset:      q.Offset,
	}
	if q.Type != nil {
		params.Column4 = toNullTxnType(q.Type)
	}
	if q.AccountID != nil {
		params.Column5 = toNullUUID(q.AccountID)
	}

	rows, err := s.queries.ListTransactions(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}

	countParams := db.CountTransactionsParams{
		HouseholdID: householdID,
		Column2:     params.Column2,
		Column3:     params.Column3,
		Column4:     params.Column4,
		Column5:     params.Column5,
	}
	total, err := s.queries.CountTransactions(ctx, countParams)
	if err != nil {
		return nil, fmt.Errorf("count transactions: %w", err)
	}

	txns := make([]model.Transaction, 0, len(rows))
	for _, r := range rows {
		txns = append(txns, toTransactionModel(r))
	}

	return &model.PaginatedResponse{
		Data:   txns,
		Total:  total,
		Limit:  q.Limit,
		Offset: q.Offset,
	}, nil
}

// Get returns a single transaction.
func (s *TransactionService) Get(ctx context.Context, id, householdID uuid.UUID) (*model.Transaction, error) {
	dbTxn, err := s.queries.GetTransaction(ctx, db.GetTransactionParams{
		ID:          id,
		HouseholdID: householdID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrTransactionNotFound
		}
		return nil, fmt.Errorf("get transaction: %w", err)
	}
	txn := toTransactionModel(dbTxn)
	return &txn, nil
}

// Update modifies a transaction, rolling back old balances and applying new ones.
func (s *TransactionService) Update(ctx context.Context, id, householdID, userID uuid.UUID, req model.UpdateTransactionRequest) (*model.Transaction, error) {
	newAmount, err := decimal.NewFromString(req.Amount)
	if err != nil {
		return nil, fmt.Errorf("invalid amount: %w", err)
	}

	if req.Type == model.TransactionTypeTransfer && req.DestinationAccountID == nil {
		return nil, ErrTransferMissingDest
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	// Get old transaction to reverse balance
	old, err := qtx.GetTransaction(ctx, db.GetTransactionParams{ID: id, HouseholdID: householdID})
	if err != nil {
		return nil, ErrTransactionNotFound
	}

	// Reverse old balance
	oldDestID := nullUUIDToPtr(old.DestinationAccountID)
	if err := s.reverseBalanceChange(ctx, qtx, model.TransactionType(old.Type), old.Amount, old.AccountID, oldDestID); err != nil {
		return nil, err
	}

	// Update transaction
	params := db.UpdateTransactionParams{
		ID:          id,
		HouseholdID: householdID,
		Description: req.Description,
		Amount:      newAmount,
		AccountID:   req.AccountID,
		Tags:        req.Tags,
		Note:        toPgText(req.Note),
		TransactedAt: pgtype.Timestamptz{
			Time:  req.TransactedAt,
			Valid: true,
		},
		Type: db.TransactionType(req.Type),
	}
	if req.DestinationAccountID != nil {
		params.DestinationAccountID = toNullUUID(req.DestinationAccountID)
	}

	dbTxn, err := qtx.UpdateTransaction(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("update transaction: %w", err)
	}

	// Apply new balance
	if err := s.applyBalanceChange(ctx, qtx, req.Type, newAmount, req.AccountID, req.DestinationAccountID); err != nil {
		return nil, err
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	txn := toTransactionModel(dbTxn)
	return &txn, nil
}

// Delete removes a transaction and reverses its balance effect.
func (s *TransactionService) Delete(ctx context.Context, id, householdID uuid.UUID) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	deleted, err := qtx.DeleteTransaction(ctx, db.DeleteTransactionParams{ID: id, HouseholdID: householdID})
	if err != nil {
		return ErrTransactionNotFound
	}

	destID := nullUUIDToPtr(deleted.DestinationAccountID)
	if err := s.reverseBalanceChange(ctx, qtx, model.TransactionType(deleted.Type), deleted.Amount, deleted.AccountID, destID); err != nil {
		return err
	}

	return tx.Commit(ctx)
}

// --- balance helpers ---

func (s *TransactionService) applyBalanceChange(ctx context.Context, qtx *db.Queries, txnType model.TransactionType, amount decimal.Decimal, accountID uuid.UUID, destID *uuid.UUID) error {
	switch txnType {
	case model.TransactionTypeIncome:
		return qtx.UpdateAccountBalance(ctx, db.UpdateAccountBalanceParams{ID: accountID, Balance: amount})
	case model.TransactionTypeExpense:
		return qtx.UpdateAccountBalance(ctx, db.UpdateAccountBalanceParams{ID: accountID, Balance: amount.Neg()})
	case model.TransactionTypeTransfer:
		if destID == nil {
			return ErrTransferMissingDest
		}
		if err := qtx.UpdateAccountBalance(ctx, db.UpdateAccountBalanceParams{ID: accountID, Balance: amount.Neg()}); err != nil {
			return err
		}
		return qtx.UpdateAccountBalance(ctx, db.UpdateAccountBalanceParams{ID: *destID, Balance: amount})
	}
	return nil
}

func (s *TransactionService) reverseBalanceChange(ctx context.Context, qtx *db.Queries, txnType model.TransactionType, amount decimal.Decimal, accountID uuid.UUID, destID *uuid.UUID) error {
	switch txnType {
	case model.TransactionTypeIncome:
		return qtx.UpdateAccountBalance(ctx, db.UpdateAccountBalanceParams{ID: accountID, Balance: amount.Neg()})
	case model.TransactionTypeExpense:
		return qtx.UpdateAccountBalance(ctx, db.UpdateAccountBalanceParams{ID: accountID, Balance: amount})
	case model.TransactionTypeTransfer:
		if destID == nil {
			return nil
		}
		if err := qtx.UpdateAccountBalance(ctx, db.UpdateAccountBalanceParams{ID: accountID, Balance: amount}); err != nil {
			return err
		}
		return qtx.UpdateAccountBalance(ctx, db.UpdateAccountBalanceParams{ID: *destID, Balance: amount.Neg()})
	}
	return nil
}

// --- conversion helpers ---

func toTransactionModel(t db.Transaction) model.Transaction {
	txn := model.Transaction{
		ID:           t.ID,
		HouseholdID:  t.HouseholdID,
		Type:         model.TransactionType(t.Type),
		Description:  t.Description,
		Amount:       t.Amount,
		AccountID:    t.AccountID,
		Tags:         t.Tags,
		TransactedAt: t.TransactedAt.Time,
		CreatedBy:    t.CreatedBy,
		CreatedAt:    t.CreatedAt.Time,
		UpdatedAt:    t.UpdatedAt.Time,
	}
	if t.Note.Valid {
		txn.Note = &t.Note.String
	}
	txn.DestinationAccountID = nullUUIDToPtr(t.DestinationAccountID)
	return txn
}

func toNullUUID(id *uuid.UUID) pgtype.UUID {
	if id == nil {
		return pgtype.UUID{}
	}
	return pgtype.UUID{Bytes: *id, Valid: true}
}

func nullUUIDToPtr(nu pgtype.UUID) *uuid.UUID {
	if !nu.Valid {
		return nil
	}
	id := uuid.UUID(nu.Bytes)
	return &id
}

func toPgText(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func toPgTimestamptz(t *time.Time) pgtype.Timestamptz {
	if t == nil {
		return pgtype.Timestamptz{}
	}
	return pgtype.Timestamptz{Time: *t, Valid: true}
}

func toNullTxnType(t *model.TransactionType) pgtype.Text {
	if t == nil {
		return pgtype.Text{}
	}
	return pgtype.Text{String: string(*t), Valid: true}
}
