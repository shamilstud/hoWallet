package db

import (
	"context"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

type CreateTransactionParams struct {
	HouseholdID          uuid.UUID
	Type                 TransactionType
	Description          string
	Amount               decimal.Decimal
	AccountID            uuid.UUID
	DestinationAccountID pgtype.UUID
	Tags                 []string
	Note                 pgtype.Text
	TransactedAt         pgtype.Timestamptz
	CreatedBy            uuid.UUID
}

func (q *Queries) CreateTransaction(ctx context.Context, arg CreateTransactionParams) (Transaction, error) {
	row := q.queryRow(ctx,
		`INSERT INTO transactions (
			household_id, type, description, amount,
			account_id, destination_account_id, tags, note,
			transacted_at, created_by
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, household_id, type, description, amount,
			account_id, destination_account_id, tags, note,
			transacted_at, created_by, created_at, updated_at`,
		arg.HouseholdID, arg.Type, arg.Description, arg.Amount,
		arg.AccountID, arg.DestinationAccountID, arg.Tags, arg.Note,
		arg.TransactedAt, arg.CreatedBy,
	)
	var t Transaction
	err := row.Scan(
		&t.ID, &t.HouseholdID, &t.Type, &t.Description, &t.Amount,
		&t.AccountID, &t.DestinationAccountID, &t.Tags, &t.Note,
		&t.TransactedAt, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
	)
	return t, err
}

type GetTransactionParams struct {
	ID          uuid.UUID
	HouseholdID uuid.UUID
}

func (q *Queries) GetTransaction(ctx context.Context, arg GetTransactionParams) (Transaction, error) {
	row := q.queryRow(ctx,
		`SELECT id, household_id, type, description, amount,
			account_id, destination_account_id, tags, note,
			transacted_at, created_by, created_at, updated_at
		 FROM transactions WHERE id = $1 AND household_id = $2`,
		arg.ID, arg.HouseholdID,
	)
	var t Transaction
	err := row.Scan(
		&t.ID, &t.HouseholdID, &t.Type, &t.Description, &t.Amount,
		&t.AccountID, &t.DestinationAccountID, &t.Tags, &t.Note,
		&t.TransactedAt, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
	)
	return t, err
}

type ListTransactionsParams struct {
	HouseholdID uuid.UUID
	Column2     pgtype.Timestamptz // from
	Column3     pgtype.Timestamptz // to
	Column4     pgtype.Text        // type filter
	Column5     pgtype.UUID        // account filter
	Limit       int32
	Offset      int32
}

func (q *Queries) ListTransactions(ctx context.Context, arg ListTransactionsParams) ([]Transaction, error) {
	rows, err := q.query(ctx,
		`SELECT id, household_id, type, description, amount,
			account_id, destination_account_id, tags, note,
			transacted_at, created_by, created_at, updated_at
		 FROM transactions
		 WHERE household_id = $1
		   AND ($2::timestamptz IS NULL OR transacted_at >= $2)
		   AND ($3::timestamptz IS NULL OR transacted_at <= $3)
		   AND ($4::transaction_type IS NULL OR type = $4)
		   AND ($5::uuid IS NULL OR account_id = $5 OR destination_account_id = $5)
		 ORDER BY transacted_at DESC
		 LIMIT $6 OFFSET $7`,
		arg.HouseholdID, arg.Column2, arg.Column3, arg.Column4, arg.Column5,
		arg.Limit, arg.Offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Transaction
	for rows.Next() {
		var t Transaction
		if err := rows.Scan(
			&t.ID, &t.HouseholdID, &t.Type, &t.Description, &t.Amount,
			&t.AccountID, &t.DestinationAccountID, &t.Tags, &t.Note,
			&t.TransactedAt, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
		); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

type CountTransactionsParams struct {
	HouseholdID uuid.UUID
	Column2     pgtype.Timestamptz
	Column3     pgtype.Timestamptz
	Column4     pgtype.Text
	Column5     pgtype.UUID
}

func (q *Queries) CountTransactions(ctx context.Context, arg CountTransactionsParams) (int64, error) {
	var count int64
	err := q.queryRow(ctx,
		`SELECT COUNT(*) FROM transactions
		 WHERE household_id = $1
		   AND ($2::timestamptz IS NULL OR transacted_at >= $2)
		   AND ($3::timestamptz IS NULL OR transacted_at <= $3)
		   AND ($4::transaction_type IS NULL OR type = $4)
		   AND ($5::uuid IS NULL OR account_id = $5 OR destination_account_id = $5)`,
		arg.HouseholdID, arg.Column2, arg.Column3, arg.Column4, arg.Column5,
	).Scan(&count)
	return count, err
}

type UpdateTransactionParams struct {
	ID                   uuid.UUID
	HouseholdID          uuid.UUID
	Description          string
	Amount               decimal.Decimal
	AccountID            uuid.UUID
	DestinationAccountID pgtype.UUID
	Tags                 []string
	Note                 pgtype.Text
	TransactedAt         pgtype.Timestamptz
	Type                 TransactionType
}

func (q *Queries) UpdateTransaction(ctx context.Context, arg UpdateTransactionParams) (Transaction, error) {
	row := q.queryRow(ctx,
		`UPDATE transactions
		 SET description            = $3,
		     amount                 = $4,
		     account_id             = $5,
		     destination_account_id = $6,
		     tags                   = $7,
		     note                   = $8,
		     transacted_at          = $9,
		     type                   = $10
		 WHERE id = $1 AND household_id = $2
		 RETURNING id, household_id, type, description, amount,
			account_id, destination_account_id, tags, note,
			transacted_at, created_by, created_at, updated_at`,
		arg.ID, arg.HouseholdID, arg.Description, arg.Amount,
		arg.AccountID, arg.DestinationAccountID, arg.Tags, arg.Note,
		arg.TransactedAt, arg.Type,
	)
	var t Transaction
	err := row.Scan(
		&t.ID, &t.HouseholdID, &t.Type, &t.Description, &t.Amount,
		&t.AccountID, &t.DestinationAccountID, &t.Tags, &t.Note,
		&t.TransactedAt, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
	)
	return t, err
}

type DeleteTransactionParams struct {
	ID          uuid.UUID
	HouseholdID uuid.UUID
}

func (q *Queries) DeleteTransaction(ctx context.Context, arg DeleteTransactionParams) (Transaction, error) {
	row := q.queryRow(ctx,
		`DELETE FROM transactions WHERE id = $1 AND household_id = $2
		 RETURNING id, household_id, type, description, amount,
			account_id, destination_account_id, tags, note,
			transacted_at, created_by, created_at, updated_at`,
		arg.ID, arg.HouseholdID,
	)
	var t Transaction
	err := row.Scan(
		&t.ID, &t.HouseholdID, &t.Type, &t.Description, &t.Amount,
		&t.AccountID, &t.DestinationAccountID, &t.Tags, &t.Note,
		&t.TransactedAt, &t.CreatedBy, &t.CreatedAt, &t.UpdatedAt,
	)
	return t, err
}

// --- Export query ---

type ListTransactionsForExportParams struct {
	HouseholdID uuid.UUID
	Column2     pgtype.Timestamptz // from
	Column3     pgtype.Timestamptz // to
}

type ListTransactionsForExportRow struct {
	TransactedAt           pgtype.Timestamptz
	Description            string
	Amount                 decimal.Decimal
	Type                   TransactionType
	Tags                   []string
	Note                   pgtype.Text
	AccountName            string
	AccountCurrency        string
	DestinationAccountName *string
}

func (q *Queries) ListTransactionsForExport(ctx context.Context, arg ListTransactionsForExportParams) ([]ListTransactionsForExportRow, error) {
	rows, err := q.query(ctx,
		`SELECT
			t.transacted_at,
			t.description,
			t.amount,
			t.type,
			t.tags,
			t.note,
			a.name  AS account_name,
			a.currency AS account_currency,
			da.name AS destination_account_name
		 FROM transactions t
		 JOIN accounts a ON a.id = t.account_id
		 LEFT JOIN accounts da ON da.id = t.destination_account_id
		 WHERE t.household_id = $1
		   AND ($2::timestamptz IS NULL OR t.transacted_at >= $2)
		   AND ($3::timestamptz IS NULL OR t.transacted_at <= $3)
		 ORDER BY t.transacted_at DESC`,
		arg.HouseholdID, arg.Column2, arg.Column3,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ListTransactionsForExportRow
	for rows.Next() {
		var r ListTransactionsForExportRow
		if err := rows.Scan(
			&r.TransactedAt, &r.Description, &r.Amount, &r.Type,
			&r.Tags, &r.Note, &r.AccountName, &r.AccountCurrency,
			&r.DestinationAccountName,
		); err != nil {
			return nil, err
		}
		out = append(out, r)
	}
	return out, rows.Err()
}
