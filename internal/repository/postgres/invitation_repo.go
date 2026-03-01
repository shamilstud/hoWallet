package postgres

import (
	"context"
	"time"

	"github.com/google/uuid"
	db "github.com/howallet/howallet/internal/db"
	"github.com/howallet/howallet/internal/model"
)

type invitationRepo struct {
	queries *db.Queries
}

func (r *invitationRepo) Create(ctx context.Context, householdID, invitedBy uuid.UUID, email, token string, expiresAt time.Time) (model.Invitation, error) {
	inv, err := r.queries.CreateInvitation(ctx, db.CreateInvitationParams{
		HouseholdID: householdID,
		Email:       email,
		InvitedBy:   invitedBy,
		Token:       token,
		ExpiresAt:   expiresAt,
	})
	if err != nil {
		return model.Invitation{}, err
	}
	return toInvitationModel(inv), nil
}

func (r *invitationRepo) GetByToken(ctx context.Context, token string) (model.Invitation, error) {
	inv, err := r.queries.GetInvitationByToken(ctx, token)
	if err != nil {
		return model.Invitation{}, err
	}
	return toInvitationModel(inv), nil
}

func (r *invitationRepo) Accept(ctx context.Context, id uuid.UUID) error {
	return r.queries.AcceptInvitation(ctx, id)
}

func (r *invitationRepo) ListPendingByHousehold(ctx context.Context, householdID uuid.UUID) ([]model.Invitation, error) {
	rows, err := r.queries.ListPendingInvitations(ctx, householdID)
	if err != nil {
		return nil, err
	}
	out := make([]model.Invitation, 0, len(rows))
	for _, inv := range rows {
		out = append(out, toInvitationModel(inv))
	}
	return out, nil
}

func toInvitationModel(i db.Invitation) model.Invitation {
	return model.Invitation{
		ID:          i.ID,
		HouseholdID: i.HouseholdID,
		Email:       i.Email,
		InvitedBy:   i.InvitedBy,
		Token:       i.Token,
		Status:      model.InvitationStatus(i.Status),
		ExpiresAt:   i.ExpiresAt.Time,
		CreatedAt:   i.CreatedAt.Time,
	}
}
