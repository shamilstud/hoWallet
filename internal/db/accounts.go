package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

type CreateAccountParams struct {
	HouseholdID uuid.UUID
	Name        string
	Type        AccountType
	Balance     decimal.Decimal
	Currency    string
	CreatedBy   uuid.UUID
}

func (q *Queries) CreateAccount(ctx context.Context, arg CreateAccountParams) (Account, error) {
	row := q.queryRow(ctx,
		`INSERT INTO accounts (household_id, name, type, balance, currency, created_by)
		 VALUES ($1, $2, $3, $4, $5, $6)
		 RETURNING id, household_id, name, type, balance, currency, created_by, created_at, updated_at`,
		arg.HouseholdID, arg.Name, arg.Type, arg.Balance, arg.Currency, arg.CreatedBy,
	)
	var a Account
	err := row.Scan(&a.ID, &a.HouseholdID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

type GetAccountParams struct {
	ID          uuid.UUID
	HouseholdID uuid.UUID
}

func (q *Queries) GetAccount(ctx context.Context, arg GetAccountParams) (Account, error) {
	row := q.queryRow(ctx,
		`SELECT id, household_id, name, type, balance, currency, created_by, created_at, updated_at
		 FROM accounts WHERE id = $1 AND household_id = $2`,
		arg.ID, arg.HouseholdID,
	)
	var a Account
	err := row.Scan(&a.ID, &a.HouseholdID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

func (q *Queries) ListAccountsByHousehold(ctx context.Context, householdID uuid.UUID) ([]Account, error) {
	rows, err := q.query(ctx,
		`SELECT id, household_id, name, type, balance, currency, created_by, created_at, updated_at
		 FROM accounts WHERE household_id = $1 ORDER BY created_at`,
		householdID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Account
	for rows.Next() {
		var a Account
		if err := rows.Scan(&a.ID, &a.HouseholdID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt); err != nil {
			return nil, err
		}
		out = append(out, a)
	}
	return out, rows.Err()
}

type UpdateAccountParams struct {
	ID          uuid.UUID
	HouseholdID uuid.UUID
	Name        *string
	Type        *AccountType
	Currency    *string
}

func (q *Queries) UpdateAccount(ctx context.Context, arg UpdateAccountParams) (Account, error) {
	row := q.queryRow(ctx,
		`UPDATE accounts
		 SET name     = COALESCE($3, name),
		     type     = COALESCE($4, type),
		     currency = COALESCE($5, currency)
		 WHERE id = $1 AND household_id = $2
		 RETURNING id, household_id, name, type, balance, currency, created_by, created_at, updated_at`,
		arg.ID, arg.HouseholdID, arg.Name, arg.Type, arg.Currency,
	)
	var a Account
	err := row.Scan(&a.ID, &a.HouseholdID, &a.Name, &a.Type, &a.Balance, &a.Currency, &a.CreatedBy, &a.CreatedAt, &a.UpdatedAt)
	return a, err
}

type UpdateAccountBalanceParams struct {
	ID      uuid.UUID
	Balance decimal.Decimal
}

func (q *Queries) UpdateAccountBalance(ctx context.Context, arg UpdateAccountBalanceParams) error {
	return q.exec(ctx,
		`UPDATE accounts SET balance = balance + $2 WHERE id = $1`,
		arg.ID, arg.Balance,
	)
}

type SetAccountBalanceParams struct {
	ID      uuid.UUID
	Balance decimal.Decimal
}

func (q *Queries) SetAccountBalance(ctx context.Context, arg SetAccountBalanceParams) error {
	return q.exec(ctx,
		`UPDATE accounts SET balance = $2 WHERE id = $1`,
		arg.ID, arg.Balance,
	)
}

type DeleteAccountParams struct {
	ID          uuid.UUID
	HouseholdID uuid.UUID
}

func (q *Queries) DeleteAccount(ctx context.Context, arg DeleteAccountParams) error {
	return q.exec(ctx,
		`DELETE FROM accounts WHERE id = $1 AND household_id = $2`,
		arg.ID, arg.HouseholdID,
	)
}

func (q *Queries) CountTransactionsByAccount(ctx context.Context, accountID uuid.UUID) (int64, error) {
	var count int64
	err := q.queryRow(ctx,
		`SELECT COUNT(*) FROM transactions WHERE account_id = $1`,
		accountID,
	).Scan(&count)
	return count, err
}
