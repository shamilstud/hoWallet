package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/howallet/howallet/internal/model"
)

// HouseholdRepository defines data access for households and members.
type HouseholdRepository interface {
	Create(ctx context.Context, name string, ownerID uuid.UUID) (model.Household, error)
	GetByID(ctx context.Context, id uuid.UUID) (model.Household, error)
	ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Household, error)
	AddMember(ctx context.Context, householdID, userID uuid.UUID, role model.HouseholdRole) error
	RemoveMember(ctx context.Context, householdID, userID uuid.UUID) error
	GetMember(ctx context.Context, householdID, userID uuid.UUID) (model.HouseholdMember, error)
	ListMembers(ctx context.Context, householdID uuid.UUID) ([]model.HouseholdMember, error)
	IsMember(ctx context.Context, householdID, userID uuid.UUID) (bool, error)
}
