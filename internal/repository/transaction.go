package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/howallet/howallet/internal/model"
	"github.com/shopspring/decimal"
)

// TransactionRepository defines data access for transactions.
type TransactionRepository interface {
	Create(ctx context.Context, params CreateTransactionParams) (model.Transaction, error)
	GetByID(ctx context.Context, id, householdID uuid.UUID) (model.Transaction, error)
	List(ctx context.Context, params ListTransactionsParams) ([]model.Transaction, error)
	Count(ctx context.Context, params CountTransactionsParams) (int64, error)
	Update(ctx context.Context, params UpdateTransactionParams) (model.Transaction, error)
	Delete(ctx context.Context, id, householdID uuid.UUID) (model.Transaction, error)
	ListForExport(ctx context.Context, householdID uuid.UUID, from, to *time.Time) ([]ExportRow, error)
}

// CreateTransactionParams holds parameters for creating a transaction.
type CreateTransactionParams struct {
	HouseholdID          uuid.UUID
	Type                 model.TransactionType
	Description          string
	Amount               decimal.Decimal
	AccountID            uuid.UUID
	DestinationAccountID *uuid.UUID
	Tags                 []string
	Note                 *string
	TransactedAt         time.Time
	CreatedBy            uuid.UUID
}

// ListTransactionsParams holds parameters for listing transactions.
type ListTransactionsParams struct {
	HouseholdID uuid.UUID
	From        *time.Time
	To          *time.Time
	Type        *model.TransactionType
	AccountID   *uuid.UUID
	Limit       int32
	Offset      int32
}

// CountTransactionsParams holds parameters for counting transactions.
type CountTransactionsParams struct {
	HouseholdID uuid.UUID
	From        *time.Time
	To          *time.Time
	Type        *model.TransactionType
	AccountID   *uuid.UUID
}

// UpdateTransactionParams holds parameters for updating a transaction.
type UpdateTransactionParams struct {
	ID                   uuid.UUID
	HouseholdID          uuid.UUID
	Type                 model.TransactionType
	Description          string
	Amount               decimal.Decimal
	AccountID            uuid.UUID
	DestinationAccountID *uuid.UUID
	Tags                 []string
	Note                 *string
	TransactedAt         time.Time
}

// ExportRow represents a transaction row for CSV export.
type ExportRow struct {
	TransactedAt           time.Time
	Description            string
	Amount                 decimal.Decimal
	Type                   model.TransactionType
	Tags                   []string
	Note                   *string
	AccountName            string
	AccountCurrency        string
	DestinationAccountName *string
}
