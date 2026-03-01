package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	db "github.com/howallet/howallet/internal/db"
	"github.com/howallet/howallet/internal/model"
	"github.com/howallet/howallet/internal/repository"
	"github.com/jackc/pgx/v5/pgtype"
)

type transactionRepo struct {
	queries *db.Queries
}

func (r *transactionRepo) Create(ctx context.Context, params repository.CreateTransactionParams) (model.Transaction, error) {
	dbParams := db.CreateTransactionParams{
		HouseholdID: params.HouseholdID,
		Type:        db.TransactionType(params.Type),
		Description: params.Description,
		Amount:      params.Amount,
		AccountID:   params.AccountID,
		Tags:        params.Tags,
		Note:        toPgText(params.Note),
		TransactedAt: pgtype.Timestamptz{
			Time:  params.TransactedAt,
			Valid: true,
		},
		CreatedBy: params.CreatedBy,
	}
	if params.DestinationAccountID != nil {
		dbParams.DestinationAccountID = toNullUUID(params.DestinationAccountID)
	}
	t, err := r.queries.CreateTransaction(ctx, dbParams)
	if err != nil {
		return model.Transaction{}, err
	}
	return toTransactionModel(t), nil
}

func (r *transactionRepo) GetByID(ctx context.Context, id, householdID uuid.UUID) (model.Transaction, error) {
	t, err := r.queries.GetTransaction(ctx, db.GetTransactionParams{ID: id, HouseholdID: householdID})
	if err != nil {
		return model.Transaction{}, err
	}
	return toTransactionModel(t), nil
}

func (r *transactionRepo) List(ctx context.Context, params repository.ListTransactionsParams) ([]model.Transaction, error) {
	dbParams := db.ListTransactionsParams{
		HouseholdID: params.HouseholdID,
		Column2:     toPgTimestamptz(params.From),
		Column3:     toPgTimestamptz(params.To),
		Limit:       params.Limit,
		Offset:      params.Offset,
	}
	if params.Type != nil {
		dbParams.Column4 = pgtype.Text{String: string(*params.Type), Valid: true}
	}
	if params.AccountID != nil {
		dbParams.Column5 = toNullUUID(params.AccountID)
	}
	rows, err := r.queries.ListTransactions(ctx, dbParams)
	if err != nil {
		return nil, err
	}
	out := make([]model.Transaction, 0, len(rows))
	for _, t := range rows {
		out = append(out, toTransactionModel(t))
	}
	return out, nil
}

func (r *transactionRepo) Count(ctx context.Context, params repository.CountTransactionsParams) (int64, error) {
	dbParams := db.CountTransactionsParams{
		HouseholdID: params.HouseholdID,
		Column2:     toPgTimestamptz(params.From),
		Column3:     toPgTimestamptz(params.To),
	}
	if params.Type != nil {
		dbParams.Column4 = pgtype.Text{String: string(*params.Type), Valid: true}
	}
	if params.AccountID != nil {
		dbParams.Column5 = toNullUUID(params.AccountID)
	}
	return r.queries.CountTransactions(ctx, dbParams)
}

func (r *transactionRepo) Update(ctx context.Context, params repository.UpdateTransactionParams) (model.Transaction, error) {
	dbParams := db.UpdateTransactionParams{
		ID:          params.ID,
		HouseholdID: params.HouseholdID,
		Description: params.Description,
		Amount:      params.Amount,
		AccountID:   params.AccountID,
		Tags:        params.Tags,
		Note:        toPgText(params.Note),
		TransactedAt: pgtype.Timestamptz{
			Time:  params.TransactedAt,
			Valid: true,
		},
		Type: db.TransactionType(params.Type),
	}
	if params.DestinationAccountID != nil {
		dbParams.DestinationAccountID = toNullUUID(params.DestinationAccountID)
	}
	t, err := r.queries.UpdateTransaction(ctx, dbParams)
	if err != nil {
		return model.Transaction{}, err
	}
	return toTransactionModel(t), nil
}

func (r *transactionRepo) Delete(ctx context.Context, id, householdID uuid.UUID) (model.Transaction, error) {
	t, err := r.queries.DeleteTransaction(ctx, db.DeleteTransactionParams{ID: id, HouseholdID: householdID})
	if err != nil {
		return model.Transaction{}, err
	}
	return toTransactionModel(t), nil
}

func (r *transactionRepo) ListForExport(ctx context.Context, householdID uuid.UUID, from, to *time.Time) ([]repository.ExportRow, error) {
	params := db.ListTransactionsForExportParams{
		HouseholdID: householdID,
	}
	if from != nil {
		params.Column2 = pgtype.Timestamptz{Time: *from, Valid: true}
	}
	if to != nil {
		params.Column3 = pgtype.Timestamptz{Time: *to, Valid: true}
	}
	rows, err := r.queries.ListTransactionsForExport(ctx, params)
	if err != nil {
		return nil, err
	}
	out := make([]repository.ExportRow, 0, len(rows))
	for _, row := range rows {
		er := repository.ExportRow{
			TransactedAt:           row.TransactedAt.Time,
			Description:            row.Description,
			Amount:                 row.Amount,
			Type:                   model.TransactionType(row.Type),
			Tags:                   row.Tags,
			AccountName:            row.AccountName,
			AccountCurrency:        row.AccountCurrency,
			DestinationAccountName: row.DestinationAccountName,
		}
		if row.Note.Valid {
			er.Note = &row.Note.String
		}
		out = append(out, er)
	}
	return out, nil
}

// --- pgtype conversion helpers ---

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
