package service

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/google/uuid"

	"github.com/howallet/howallet/internal/model"
	"github.com/howallet/howallet/internal/repository"
)

// ExportService handles CSV export in Buxfer-compatible format.
type ExportService struct {
	transactions repository.TransactionRepository
}

func NewExportService(transactions repository.TransactionRepository) *ExportService {
	return &ExportService{transactions: transactions}
}

// ExportCSV writes Buxfer-format CSV to the given writer.
// Columns: Date,Description,Amount,Account,Tags,Type,Status,Currency
func (s *ExportService) ExportCSV(ctx context.Context, w io.Writer, householdID uuid.UUID, from, to *time.Time) error {
	rows, err := s.transactions.ListForExport(ctx, householdID, from, to)
	if err != nil {
		return fmt.Errorf("list transactions for export: %w", err)
	}

	cw := csv.NewWriter(w)
	defer cw.Flush()

	// Header
	if err := cw.Write([]string{"Date", "Description", "Amount", "Account", "Tags", "Type", "Status", "Currency"}); err != nil {
		return err
	}

	for _, r := range rows {
		txnType := string(r.Type)

		if r.Type == model.TransactionTypeTransfer {
			// Two rows for transfers (Buxfer convention)
			// 1) Outgoing from source
			if err := cw.Write([]string{
				r.TransactedAt.Format("2006-01-02"),
				r.Description,
				r.Amount.Neg().StringFixed(2),
				r.AccountName,
				strings.Join(r.Tags, ", "),
				txnType,
				"cleared",
				r.AccountCurrency,
			}); err != nil {
				return err
			}
			// 2) Incoming to destination
			destName := ""
			if r.DestinationAccountName != nil {
				destName = *r.DestinationAccountName
			}
			if err := cw.Write([]string{
				r.TransactedAt.Format("2006-01-02"),
				r.Description,
				r.Amount.StringFixed(2),
				destName,
				strings.Join(r.Tags, ", "),
				txnType,
				"cleared",
				r.AccountCurrency,
			}); err != nil {
				return err
			}
		} else {
			// Income or expense — single row
			amt := r.Amount
			if r.Type == model.TransactionTypeExpense {
				amt = amt.Neg()
			}
			if err := cw.Write([]string{
				r.TransactedAt.Format("2006-01-02"),
				r.Description,
				amt.StringFixed(2),
				r.AccountName,
				strings.Join(r.Tags, ", "),
				txnType,
				"cleared",
				r.AccountCurrency,
			}); err != nil {
				return err
			}
		}
	}

	return nil
}
