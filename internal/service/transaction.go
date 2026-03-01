package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/howallet/howallet/internal/model"
	"github.com/howallet/howallet/internal/repository"
	"github.com/howallet/howallet/internal/repository/postgres"
)

var (
	ErrTransactionNotFound = errors.New("transaction not found")
	ErrTransferMissingDest = errors.New("transfer requires destination_account_id")
)

type TransactionService struct {
	repos *postgres.Repos
}

func NewTransactionService(repos *postgres.Repos) *TransactionService {
	return &TransactionService{repos: repos}
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

	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	var txn model.Transaction
	err = s.repos.RunInTx(ctx, func(txCtx context.Context) error {
		txRepos := postgres.TxReposFromCtx(txCtx)

		var txErr error
		txn, txErr = txRepos.Transactions.Create(txCtx, repository.CreateTransactionParams{
			HouseholdID:          householdID,
			Type:                 req.Type,
			Description:          req.Description,
			Amount:               amount,
			AccountID:            req.AccountID,
			DestinationAccountID: req.DestinationAccountID,
			Tags:                 tags,
			Note:                 req.Note,
			TransactedAt:         req.TransactedAt,
			CreatedBy:            userID,
		})
		if txErr != nil {
			return fmt.Errorf("create transaction: %w", txErr)
		}

		return applyBalanceChange(txCtx, txRepos.Accounts, req.Type, amount, req.AccountID, req.DestinationAccountID)
	})
	if err != nil {
		return nil, err
	}

	return &txn, nil
}

// List returns paginated transactions with filters.
func (s *TransactionService) List(ctx context.Context, householdID uuid.UUID, q model.ListTransactionsQuery) (*model.PaginatedResponse, error) {
	if q.Limit <= 0 {
		q.Limit = 50
	}

	params := repository.ListTransactionsParams{
		HouseholdID: householdID,
		From:        q.From,
		To:          q.To,
		Type:        q.Type,
		AccountID:   q.AccountID,
		Limit:       q.Limit,
		Offset:      q.Offset,
	}

	txns, err := s.repos.Transactions.List(ctx, params)
	if err != nil {
		return nil, fmt.Errorf("list transactions: %w", err)
	}

	total, err := s.repos.Transactions.Count(ctx, repository.CountTransactionsParams{
		HouseholdID: householdID,
		From:        q.From,
		To:          q.To,
		Type:        q.Type,
		AccountID:   q.AccountID,
	})
	if err != nil {
		return nil, fmt.Errorf("count transactions: %w", err)
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
	txn, err := s.repos.Transactions.GetByID(ctx, id, householdID)
	if err != nil {
		return nil, ErrTransactionNotFound
	}
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

	tags := req.Tags
	if tags == nil {
		tags = []string{}
	}

	var txn model.Transaction
	err = s.repos.RunInTx(ctx, func(txCtx context.Context) error {
		txRepos := postgres.TxReposFromCtx(txCtx)

		// Get old transaction to reverse balance
		old, txErr := txRepos.Transactions.GetByID(txCtx, id, householdID)
		if txErr != nil {
			return ErrTransactionNotFound
		}

		// Reverse old balance
		if txErr = reverseBalanceChange(txCtx, txRepos.Accounts, old.Type, old.Amount, old.AccountID, old.DestinationAccountID); txErr != nil {
			return txErr
		}

		// Update transaction
		txn, txErr = txRepos.Transactions.Update(txCtx, repository.UpdateTransactionParams{
			ID:                   id,
			HouseholdID:          householdID,
			Type:                 req.Type,
			Description:          req.Description,
			Amount:               newAmount,
			AccountID:            req.AccountID,
			DestinationAccountID: req.DestinationAccountID,
			Tags:                 tags,
			Note:                 req.Note,
			TransactedAt:         req.TransactedAt,
		})
		if txErr != nil {
			return fmt.Errorf("update transaction: %w", txErr)
		}

		// Apply new balance
		return applyBalanceChange(txCtx, txRepos.Accounts, req.Type, newAmount, req.AccountID, req.DestinationAccountID)
	})
	if err != nil {
		return nil, err
	}

	return &txn, nil
}

// Delete removes a transaction and reverses its balance effect.
func (s *TransactionService) Delete(ctx context.Context, id, householdID uuid.UUID) error {
	return s.repos.RunInTx(ctx, func(txCtx context.Context) error {
		txRepos := postgres.TxReposFromCtx(txCtx)

		deleted, err := txRepos.Transactions.Delete(txCtx, id, householdID)
		if err != nil {
			return ErrTransactionNotFound
		}

		return reverseBalanceChange(txCtx, txRepos.Accounts, deleted.Type, deleted.Amount, deleted.AccountID, deleted.DestinationAccountID)
	})
}

// --- balance helpers ---

func applyBalanceChange(ctx context.Context, accounts repository.AccountRepository, txnType model.TransactionType, amount decimal.Decimal, accountID uuid.UUID, destID *uuid.UUID) error {
	switch txnType {
	case model.TransactionTypeIncome:
		return accounts.UpdateBalance(ctx, accountID, amount)
	case model.TransactionTypeExpense:
		return accounts.UpdateBalance(ctx, accountID, amount.Neg())
	case model.TransactionTypeTransfer:
		if destID == nil {
			return ErrTransferMissingDest
		}
		if err := accounts.UpdateBalance(ctx, accountID, amount.Neg()); err != nil {
			return err
		}
		return accounts.UpdateBalance(ctx, *destID, amount)
	}
	return nil
}

func reverseBalanceChange(ctx context.Context, accounts repository.AccountRepository, txnType model.TransactionType, amount decimal.Decimal, accountID uuid.UUID, destID *uuid.UUID) error {
	switch txnType {
	case model.TransactionTypeIncome:
		return accounts.UpdateBalance(ctx, accountID, amount.Neg())
	case model.TransactionTypeExpense:
		return accounts.UpdateBalance(ctx, accountID, amount)
	case model.TransactionTypeTransfer:
		if destID == nil {
			return nil
		}
		if err := accounts.UpdateBalance(ctx, accountID, amount); err != nil {
			return err
		}
		return accounts.UpdateBalance(ctx, *destID, amount.Neg())
	}
	return nil
}
