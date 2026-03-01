package postgres

import (
	"context"

	"github.com/google/uuid"
	db "github.com/howallet/howallet/internal/db"
	"github.com/howallet/howallet/internal/model"
)

type householdRepo struct {
	queries *db.Queries
}

func (r *householdRepo) Create(ctx context.Context, name string, ownerID uuid.UUID) (model.Household, error) {
	h, err := r.queries.CreateHousehold(ctx, db.CreateHouseholdParams{Name: name, OwnerID: ownerID})
	if err != nil {
		return model.Household{}, err
	}
	return toHouseholdModel(h), nil
}

func (r *householdRepo) GetByID(ctx context.Context, id uuid.UUID) (model.Household, error) {
	h, err := r.queries.GetHousehold(ctx, id)
	if err != nil {
		return model.Household{}, err
	}
	return toHouseholdModel(h), nil
}

func (r *householdRepo) ListByUser(ctx context.Context, userID uuid.UUID) ([]model.Household, error) {
	rows, err := r.queries.ListUserHouseholds(ctx, userID)
	if err != nil {
		return nil, err
	}
	out := make([]model.Household, 0, len(rows))
	for _, h := range rows {
		out = append(out, toHouseholdModel(h))
	}
	return out, nil
}

func (r *householdRepo) AddMember(ctx context.Context, householdID, userID uuid.UUID, role model.HouseholdRole) error {
	return r.queries.AddHouseholdMember(ctx, db.AddHouseholdMemberParams{
		HouseholdID: householdID,
		UserID:      userID,
		Role:        db.HouseholdRole(role),
	})
}

func (r *householdRepo) RemoveMember(ctx context.Context, householdID, userID uuid.UUID) error {
	return r.queries.RemoveHouseholdMember(ctx, db.RemoveHouseholdMemberParams{
		HouseholdID: householdID,
		UserID:      userID,
	})
}

func (r *householdRepo) GetMember(ctx context.Context, householdID, userID uuid.UUID) (model.HouseholdMember, error) {
	m, err := r.queries.GetHouseholdMember(ctx, db.GetHouseholdMemberParams{
		HouseholdID: householdID,
		UserID:      userID,
	})
	if err != nil {
		return model.HouseholdMember{}, err
	}
	return model.HouseholdMember{
		HouseholdID: m.HouseholdID,
		UserID:      m.UserID,
		Role:        model.HouseholdRole(m.Role),
		JoinedAt:    m.JoinedAt.Time,
	}, nil
}

func (r *householdRepo) ListMembers(ctx context.Context, householdID uuid.UUID) ([]model.HouseholdMember, error) {
	rows, err := r.queries.ListHouseholdMembers(ctx, householdID)
	if err != nil {
		return nil, err
	}
	out := make([]model.HouseholdMember, 0, len(rows))
	for _, m := range rows {
		out = append(out, model.HouseholdMember{
			HouseholdID: m.HouseholdID,
			UserID:      m.UserID,
			Role:        model.HouseholdRole(m.Role),
			JoinedAt:    m.JoinedAt,
			Email:       m.Email,
			UserName:    m.UserName,
		})
	}
	return out, nil
}

func (r *householdRepo) IsMember(ctx context.Context, householdID, userID uuid.UUID) (bool, error) {
	return r.queries.IsHouseholdMember(ctx, db.IsHouseholdMemberParams{
		HouseholdID: householdID,
		UserID:      userID,
	})
}

func toHouseholdModel(h db.Household) model.Household {
	return model.Household{
		ID:        h.ID,
		Name:      h.Name,
		OwnerID:   h.OwnerID,
		CreatedAt: h.CreatedAt.Time,
	}
}
