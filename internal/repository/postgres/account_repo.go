package postgres

import (
	"context"

	"github.com/google/uuid"
	db "github.com/howallet/howallet/internal/db"
	"github.com/howallet/howallet/internal/model"
	"github.com/howallet/howallet/internal/repository"
	"github.com/shopspring/decimal"
)

type accountRepo struct {
	queries *db.Queries
}

func (r *accountRepo) Create(ctx context.Context, params repository.CreateAccountParams) (model.Account, error) {
	a, err := r.queries.CreateAccount(ctx, db.CreateAccountParams{
		HouseholdID: params.HouseholdID,
		Name:        params.Name,
		Type:        db.AccountType(params.Type),
		Balance:     params.Balance,
		Currency:    params.Currency,
		CreatedBy:   params.CreatedBy,
	})
	if err != nil {
		return model.Account{}, err
	}
	return toAccountModel(a), nil
}

func (r *accountRepo) GetByID(ctx context.Context, id, householdID uuid.UUID) (model.Account, error) {
	a, err := r.queries.GetAccount(ctx, db.GetAccountParams{ID: id, HouseholdID: householdID})
	if err != nil {
		return model.Account{}, err
	}
	return toAccountModel(a), nil
}

func (r *accountRepo) ListByHousehold(ctx context.Context, householdID uuid.UUID) ([]model.Account, error) {
	rows, err := r.queries.ListAccountsByHousehold(ctx, householdID)
	if err != nil {
		return nil, err
	}
	out := make([]model.Account, 0, len(rows))
	for _, a := range rows {
		out = append(out, toAccountModel(a))
	}
	return out, nil
}

func (r *accountRepo) Update(ctx context.Context, params repository.UpdateAccountParams) (model.Account, error) {
	dbParams := db.UpdateAccountParams{
		ID:          params.ID,
		HouseholdID: params.HouseholdID,
	}
	if params.Name != nil {
		dbParams.Name = params.Name
	}
	if params.Type != nil {
		t := db.AccountType(*params.Type)
		dbParams.Type = &t
	}
	if params.Currency != nil {
		dbParams.Currency = params.Currency
	}
	a, err := r.queries.UpdateAccount(ctx, dbParams)
	if err != nil {
		return model.Account{}, err
	}
	return toAccountModel(a), nil
}

func (r *accountRepo) Delete(ctx context.Context, id, householdID uuid.UUID) error {
	return r.queries.DeleteAccount(ctx, db.DeleteAccountParams{ID: id, HouseholdID: householdID})
}

func (r *accountRepo) UpdateBalance(ctx context.Context, id uuid.UUID, delta decimal.Decimal) error {
	return r.queries.UpdateAccountBalance(ctx, db.UpdateAccountBalanceParams{ID: id, Balance: delta})
}

func (r *accountRepo) CountTransactions(ctx context.Context, accountID uuid.UUID) (int64, error) {
	return r.queries.CountTransactionsByAccount(ctx, accountID)
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
