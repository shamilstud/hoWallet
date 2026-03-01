package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/howallet/howallet/internal/model"
	"github.com/shopspring/decimal"
)

// AccountRepository defines data access for accounts.
type AccountRepository interface {
	Create(ctx context.Context, params CreateAccountParams) (model.Account, error)
	GetByID(ctx context.Context, id, householdID uuid.UUID) (model.Account, error)
	ListByHousehold(ctx context.Context, householdID uuid.UUID) ([]model.Account, error)
	Update(ctx context.Context, params UpdateAccountParams) (model.Account, error)
	Delete(ctx context.Context, id, householdID uuid.UUID) error
	UpdateBalance(ctx context.Context, id uuid.UUID, delta decimal.Decimal) error
	CountTransactions(ctx context.Context, accountID uuid.UUID) (int64, error)
}

// CreateAccountParams holds parameters for creating an account.
type CreateAccountParams struct {
	HouseholdID uuid.UUID
	Name        string
	Type        model.AccountType
	Balance     decimal.Decimal
	Currency    string
	CreatedBy   uuid.UUID
}

// UpdateAccountParams holds parameters for updating an account.
type UpdateAccountParams struct {
	ID          uuid.UUID
	HouseholdID uuid.UUID
	Name        *string
	Type        *model.AccountType
	Currency    *string
}
