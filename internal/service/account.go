package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"

	"github.com/howallet/howallet/internal/model"
	"github.com/howallet/howallet/internal/repository"
)

var (
	ErrAccountNotFound        = errors.New("account not found")
	ErrAccountHasTransactions = errors.New("account has transactions, cannot delete")
)

type AccountService struct {
	accounts repository.AccountRepository
}

func NewAccountService(accounts repository.AccountRepository) *AccountService {
	return &AccountService{accounts: accounts}
}

func (s *AccountService) Create(ctx context.Context, householdID, userID uuid.UUID, req model.CreateAccountRequest) (*model.Account, error) {
	balance, err := decimal.NewFromString(req.Balance)
	if err != nil {
		return nil, fmt.Errorf("invalid balance: %w", err)
	}

	currency := req.Currency
	if currency == "" {
		currency = "USD"
	}

	acc, err := s.accounts.Create(ctx, repository.CreateAccountParams{
		HouseholdID: householdID,
		Name:        req.Name,
		Type:        req.Type,
		Balance:     balance,
		Currency:    currency,
		CreatedBy:   userID,
	})
	if err != nil {
		return nil, fmt.Errorf("create account: %w", err)
	}

	return &acc, nil
}

func (s *AccountService) List(ctx context.Context, householdID uuid.UUID) ([]model.Account, error) {
	accounts, err := s.accounts.ListByHousehold(ctx, householdID)
	if err != nil {
		return nil, fmt.Errorf("list accounts: %w", err)
	}
	return accounts, nil
}

func (s *AccountService) Get(ctx context.Context, id, householdID uuid.UUID) (*model.Account, error) {
	acc, err := s.accounts.GetByID(ctx, id, householdID)
	if err != nil {
		return nil, ErrAccountNotFound
	}
	return &acc, nil
}

func (s *AccountService) Update(ctx context.Context, id, householdID uuid.UUID, req model.UpdateAccountRequest) (*model.Account, error) {
	acc, err := s.accounts.Update(ctx, repository.UpdateAccountParams{
		ID:          id,
		HouseholdID: householdID,
		Name:        req.Name,
		Type:        req.Type,
		Currency:    req.Currency,
	})
	if err != nil {
		return nil, ErrAccountNotFound
	}
	return &acc, nil
}

func (s *AccountService) Delete(ctx context.Context, id, householdID uuid.UUID) error {
	count, err := s.accounts.CountTransactions(ctx, id)
	if err != nil {
		return fmt.Errorf("count transactions: %w", err)
	}
	if count > 0 {
		return ErrAccountHasTransactions
	}

	return s.accounts.Delete(ctx, id, householdID)
}
