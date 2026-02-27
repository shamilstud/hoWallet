package db

import (
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/shopspring/decimal"
)

// Enum types matching PostgreSQL enums
type AccountType string

const (
	AccountTypeCard    AccountType = "card"
	AccountTypeDeposit AccountType = "deposit"
	AccountTypeCash    AccountType = "cash"
)

type TransactionType string

const (
	TransactionTypeIncome   TransactionType = "income"
	TransactionTypeExpense  TransactionType = "expense"
	TransactionTypeTransfer TransactionType = "transfer"
)

type HouseholdRole string

const (
	HouseholdRoleOwner  HouseholdRole = "owner"
	HouseholdRoleMember HouseholdRole = "member"
)

type InvitationStatus string

const (
	InvitationStatusPending  InvitationStatus = "pending"
	InvitationStatusAccepted InvitationStatus = "accepted"
	InvitationStatusExpired  InvitationStatus = "expired"
)

// Table models
type User struct {
	ID           uuid.UUID          `json:"id"`
	Email        string             `json:"email"`
	PasswordHash string             `json:"password_hash"`
	Name         string             `json:"name"`
	CreatedAt    pgtype.Timestamptz `json:"created_at"`
	UpdatedAt    pgtype.Timestamptz `json:"updated_at"`
}

type Household struct {
	ID        uuid.UUID          `json:"id"`
	Name      string             `json:"name"`
	OwnerID   uuid.UUID          `json:"owner_id"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}

type HouseholdMember struct {
	HouseholdID uuid.UUID          `json:"household_id"`
	UserID      uuid.UUID          `json:"user_id"`
	Role        HouseholdRole      `json:"role"`
	JoinedAt    pgtype.Timestamptz `json:"joined_at"`
}

type Invitation struct {
	ID          uuid.UUID          `json:"id"`
	HouseholdID uuid.UUID          `json:"household_id"`
	Email       string             `json:"email"`
	InvitedBy   uuid.UUID          `json:"invited_by"`
	Token       string             `json:"token"`
	Status      InvitationStatus   `json:"status"`
	ExpiresAt   pgtype.Timestamptz `json:"expires_at"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
}

type Account struct {
	ID          uuid.UUID          `json:"id"`
	HouseholdID uuid.UUID          `json:"household_id"`
	Name        string             `json:"name"`
	Type        AccountType        `json:"type"`
	Balance     decimal.Decimal    `json:"balance"`
	Currency    string             `json:"currency"`
	CreatedBy   uuid.UUID          `json:"created_by"`
	CreatedAt   pgtype.Timestamptz `json:"created_at"`
	UpdatedAt   pgtype.Timestamptz `json:"updated_at"`
}

type Transaction struct {
	ID                   uuid.UUID          `json:"id"`
	HouseholdID          uuid.UUID          `json:"household_id"`
	Type                 TransactionType    `json:"type"`
	Description          string             `json:"description"`
	Amount               decimal.Decimal    `json:"amount"`
	AccountID            uuid.UUID          `json:"account_id"`
	DestinationAccountID pgtype.UUID        `json:"destination_account_id"`
	Tags                 []string           `json:"tags"`
	Note                 pgtype.Text        `json:"note"`
	TransactedAt         pgtype.Timestamptz `json:"transacted_at"`
	CreatedBy            uuid.UUID          `json:"created_by"`
	CreatedAt            pgtype.Timestamptz `json:"created_at"`
	UpdatedAt            pgtype.Timestamptz `json:"updated_at"`
}

type RefreshToken struct {
	ID        uuid.UUID          `json:"id"`
	UserID    uuid.UUID          `json:"user_id"`
	TokenHash string             `json:"token_hash"`
	ExpiresAt pgtype.Timestamptz `json:"expires_at"`
	CreatedAt pgtype.Timestamptz `json:"created_at"`
}

// Helper: convert time.Time to pgtype.Timestamptz
func ToPgTimestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
