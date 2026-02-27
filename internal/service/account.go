package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/shopspring/decimal"

	db "github.com/howallet/howallet/internal/db"
	"github.com/howallet/howallet/internal/model"
)

var (
	ErrAccountNotFound        = errors.New("account not found")
	ErrAccountHasTransactions = errors.New("account has transactions, cannot delete")
)

type AccountService struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewAccountService(pool *pgxpool.Pool, queries *db.Queries) *AccountService {
	return &AccountService{queries: queries, pool: pool}
}

func (s *AccountService) Create(ctx context.Context, householdID, userID uuid.UUID, req model.CreateAccountRequest) (*model.Account, error) {
	balance, err := decimal.NewFromString(req.Balance)
	if err != nil {
		return nil, fmt.Errorf("invalid balance: %w", err)
	}

	currency := req.Currency
	if currency == "" {
		currency = "UAH"
	}

	dbAcc, err := s.queries.CreateAccount(ctx, db.CreateAccountParams{
		HouseholdID: householdID,
		Name:        req.Name,
		Type:        db.AccountType(req.Type),
		Balance:     balance,
		Currency:    currency,
		CreatedBy:   userID,
	})
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}

	acc := toAccountModel(dbAcc)
	return &acc, nil
}

func (s *AccountService) List(ctx context.Context, householdID uuid.UUID) ([]model.Account, error) {
	rows, err := s.queries.ListAccountsByHousehold(ctx, householdID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}

	accounts := make([]model.Account, 0, len(rows))
	for _, r := range rows {
		accounts = append(accounts, toAccountModel(r))
	}
	return accounts, nil
}

func (s *AccountService) Get(ctx context.Context, id, householdID uuid.UUID) (*model.Account, error) {
	dbAcc, err := s.queries.GetAccount(ctx, db.GetAccountParams{
		ID:          id,
		HouseholdID: householdID,
	})
	if err != nil {
		return nil, ErrAccountNotFound
	}
	acc := toAccountModel(dbAcc)
	return &acc, nil
}

func (s *AccountService) Update(ctx context.Context, id, householdID uuid.UUID, req model.UpdateAccountRequest) (*model.Account, error) {
	params := db.UpdateAccountParams{
		ID:          id,
		HouseholdID: householdID,
	}

	if req.Name != nil {
		params.Name = req.Name
	}
	if req.Type != nil {
		t := db.AccountType(*req.Type)
		params.Type = &t
	}
	if req.Currency != nil {
		params.Currency = req.Currency
	}

	dbAcc, err := s.queries.UpdateAccount(ctx, params)
	if err != nil {
		return nil, ErrAccountNotFound
	}

	acc := toAccountModel(dbAcc)
	return &acc, nil
}

func (s *AccountService) Delete(ctx context.Context, id, householdID uuid.UUID) error {
	count, err := s.queries.CountTransactionsByAccount(ctx, id)
	if err != nil {
		return fmt.Errorf("count transactions: %w", err)
	}
	if count > 0 {
		return ErrAccountHasTransactions
	}

	return s.queries.DeleteAccount(ctx, db.DeleteAccountParams{
		ID:          id,
		HouseholdID: householdID,
	})
}

func toAccountModel(a db.Account) model.Account {
	return model.Account{
		ID:          a.ID,
		HouseholdID: a.HouseholdID,
		Name:        a.Name,
		Type:        model.AccountType(a.Type),
		Balance:     a.Balance,
		Currency:    a.Currency,
		CreatedBy:   a.CreatedBy,
		CreatedAt:   a.CreatedAt.Time,
		UpdatedAt:   a.UpdatedAt.Time,
	}
}
