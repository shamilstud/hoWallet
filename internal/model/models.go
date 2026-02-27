package model

import (
	"time"

	"github.com/google/uuid"
	"github.com/shopspring/decimal"
)

// ------------------------------------------------------------------
// Enums
// ------------------------------------------------------------------

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

// ------------------------------------------------------------------
// Domain entities
// ------------------------------------------------------------------

type User struct {
	ID           uuid.UUID `json:"id"`
	Email        string    `json:"email"`
	PasswordHash string    `json:"-"`
	Name         string    `json:"name"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
}

type Household struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	OwnerID   uuid.UUID `json:"owner_id"`
	CreatedAt time.Time `json:"created_at"`
}

type HouseholdMember struct {
	HouseholdID uuid.UUID     `json:"household_id"`
	UserID      uuid.UUID     `json:"user_id"`
	Role        HouseholdRole `json:"role"`
	JoinedAt    time.Time     `json:"joined_at"`
	Email       string        `json:"email,omitempty"`
	UserName    string        `json:"user_name,omitempty"`
}

type Invitation struct {
	ID          uuid.UUID        `json:"id"`
	HouseholdID uuid.UUID        `json:"household_id"`
	Email       string           `json:"email"`
	InvitedBy   uuid.UUID        `json:"invited_by"`
	Token       string           `json:"token"`
	Status      InvitationStatus `json:"status"`
	ExpiresAt   time.Time        `json:"expires_at"`
	CreatedAt   time.Time        `json:"created_at"`
}

type Account struct {
	ID          uuid.UUID       `json:"id"`
	HouseholdID uuid.UUID       `json:"household_id"`
	Name        string          `json:"name"`
	Type        AccountType     `json:"type"`
	Balance     decimal.Decimal `json:"balance"`
	Currency    string          `json:"currency"`
	CreatedBy   uuid.UUID       `json:"created_by"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type Transaction struct {
	ID                   uuid.UUID       `json:"id"`
	HouseholdID          uuid.UUID       `json:"household_id"`
	Type                 TransactionType `json:"type"`
	Description          string          `json:"description"`
	Amount               decimal.Decimal `json:"amount"`
	AccountID            uuid.UUID       `json:"account_id"`
	DestinationAccountID *uuid.UUID      `json:"destination_account_id,omitempty"`
	Tags                 []string        `json:"tags"`
	Note                 *string         `json:"note,omitempty"`
	TransactedAt         time.Time       `json:"transacted_at"`
	CreatedBy            uuid.UUID       `json:"created_by"`
	CreatedAt            time.Time       `json:"created_at"`
	UpdatedAt            time.Time       `json:"updated_at"`
}

// ------------------------------------------------------------------
// API request / response DTOs
// ------------------------------------------------------------------

// Auth
type RegisterRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type AuthResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token,omitempty"`
	User         User   `json:"user"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refresh_token"`
}

// Household
type CreateHouseholdRequest struct {
	Name string `json:"name"`
}

type InviteRequest struct {
	Email string `json:"email"`
}

// Account
type CreateAccountRequest struct {
	Name     string      `json:"name"`
	Type     AccountType `json:"type"`
	Balance  string      `json:"balance"`
	Currency string      `json:"currency"`
}

type UpdateAccountRequest struct {
	Name     *string      `json:"name,omitempty"`
	Type     *AccountType `json:"type,omitempty"`
	Currency *string      `json:"currency,omitempty"`
}

// Transaction
type CreateTransactionRequest struct {
	Type                 TransactionType `json:"type"`
	Description          string          `json:"description"`
	Amount               string          `json:"amount"`
	AccountID            uuid.UUID       `json:"account_id"`
	DestinationAccountID *uuid.UUID      `json:"destination_account_id,omitempty"`
	Tags                 []string        `json:"tags"`
	Note                 *string         `json:"note,omitempty"`
	TransactedAt         time.Time       `json:"transacted_at"`
}

type UpdateTransactionRequest struct {
	Type                 TransactionType `json:"type"`
	Description          string          `json:"description"`
	Amount               string          `json:"amount"`
	AccountID            uuid.UUID       `json:"account_id"`
	DestinationAccountID *uuid.UUID      `json:"destination_account_id,omitempty"`
	Tags                 []string        `json:"tags"`
	Note                 *string         `json:"note,omitempty"`
	TransactedAt         time.Time       `json:"transacted_at"`
}

// Pagination
type ListTransactionsQuery struct {
	From      *time.Time       `json:"from,omitempty"`
	To        *time.Time       `json:"to,omitempty"`
	Type      *TransactionType `json:"type,omitempty"`
	AccountID *uuid.UUID       `json:"account_id,omitempty"`
	Limit     int32            `json:"limit"`
	Offset    int32            `json:"offset"`
}

type PaginatedResponse struct {
	Data   interface{} `json:"data"`
	Total  int64       `json:"total"`
	Limit  int32       `json:"limit"`
	Offset int32       `json:"offset"`
}
