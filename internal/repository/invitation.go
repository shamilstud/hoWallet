package repository

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/howallet/howallet/internal/model"
)

// InvitationRepository defines data access for invitations.
type InvitationRepository interface {
	Create(ctx context.Context, householdID, invitedBy uuid.UUID, email, token string, expiresAt time.Time) (model.Invitation, error)
	GetByToken(ctx context.Context, token string) (model.Invitation, error)
	Accept(ctx context.Context, id uuid.UUID) error
	ListPendingByHousehold(ctx context.Context, householdID uuid.UUID) ([]model.Invitation, error)
}
