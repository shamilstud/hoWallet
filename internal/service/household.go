package service

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/howallet/howallet/internal/db"
	"github.com/howallet/howallet/internal/model"
)

var (
	ErrHouseholdNotFound = errors.New("household not found")
	ErrNotHouseholdOwner = errors.New("only household owner can perform this action")
	ErrNotMember         = errors.New("user is not a member of this household")
	ErrInvitationInvalid = errors.New("invitation is invalid or expired")
	ErrAlreadyMember     = errors.New("user is already a member")
)

type HouseholdService struct {
	queries *db.Queries
	pool    *pgxpool.Pool
}

func NewHouseholdService(pool *pgxpool.Pool, queries *db.Queries) *HouseholdService {
	return &HouseholdService{queries: queries, pool: pool}
}

func (s *HouseholdService) Create(ctx context.Context, userID uuid.UUID, req model.CreateHouseholdRequest) (*model.Household, error) {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	hh, err := qtx.CreateHousehold(ctx, db.CreateHouseholdParams{
		Name:    req.Name,
		OwnerID: userID,
	})
	if err != nil {
		return nil, fmt.Errorf("create household: %w", err)
	}

	err = qtx.AddHouseholdMember(ctx, db.AddHouseholdMemberParams{
		HouseholdID: hh.ID,
		UserID:      userID,
		Role:        db.HouseholdRoleOwner,
	})
	if err != nil {
		return nil, fmt.Errorf("add owner: %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("commit: %w", err)
	}

	result := toHouseholdModel(hh)
	return &result, nil
}

func (s *HouseholdService) List(ctx context.Context, userID uuid.UUID) ([]model.Household, error) {
	rows, err := s.queries.ListUserHouseholds(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list households: %w", err)
	}

	out := make([]model.Household, 0, len(rows))
	for _, r := range rows {
		out = append(out, toHouseholdModel(r))
	}
	return out, nil
}

func (s *HouseholdService) Get(ctx context.Context, id uuid.UUID) (*model.Household, error) {
	hh, err := s.queries.GetHousehold(ctx, id)
	if err != nil {
		return nil, ErrHouseholdNotFound
	}
	result := toHouseholdModel(hh)
	return &result, nil
}

func (s *HouseholdService) ListMembers(ctx context.Context, householdID uuid.UUID) ([]model.HouseholdMember, error) {
	rows, err := s.queries.ListHouseholdMembers(ctx, householdID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}

	members := make([]model.HouseholdMember, 0, len(rows))
	for _, r := range rows {
		members = append(members, model.HouseholdMember{
			HouseholdID: r.HouseholdID,
			UserID:      r.UserID,
			Role:        model.HouseholdRole(r.Role),
			JoinedAt:    r.JoinedAt,
			Email:       r.Email,
			UserName:    r.UserName,
		})
	}
	return members, nil
}

func (s *HouseholdService) RemoveMember(ctx context.Context, householdID, ownerID, targetUserID uuid.UUID) error {
	member, err := s.queries.GetHouseholdMember(ctx, db.GetHouseholdMemberParams{
		HouseholdID: householdID,
		UserID:      ownerID,
	})
	if err != nil {
		return ErrNotMember
	}
	if member.Role != db.HouseholdRoleOwner {
		return ErrNotHouseholdOwner
	}

	return s.queries.RemoveHouseholdMember(ctx, db.RemoveHouseholdMemberParams{
		HouseholdID: householdID,
		UserID:      targetUserID,
	})
}

// Invite creates an invitation token for the given email.
func (s *HouseholdService) Invite(ctx context.Context, householdID, inviterID uuid.UUID, email string) (*model.Invitation, error) {
	// Verify inviter is owner
	member, err := s.queries.GetHouseholdMember(ctx, db.GetHouseholdMemberParams{
		HouseholdID: householdID,
		UserID:      inviterID,
	})
	if err != nil {
		return nil, ErrNotMember
	}
	if member.Role != db.HouseholdRoleOwner {
		return nil, ErrNotHouseholdOwner
	}

	// Check if already a member
	existingUser, err := s.queries.GetUserByEmail(ctx, email)
	if err == nil {
		isMember, _ := s.queries.IsHouseholdMember(ctx, db.IsHouseholdMemberParams{
			HouseholdID: householdID,
			UserID:      existingUser.ID,
		})
		if isMember {
			return nil, ErrAlreadyMember
		}
	}

	// Generate token
	tokenBytes := make([]byte, 32)
	_, _ = rand.Read(tokenBytes)
	token := hex.EncodeToString(tokenBytes)

	inv, err := s.queries.CreateInvitation(ctx, db.CreateInvitationParams{
		HouseholdID: householdID,
		Email:       email,
		InvitedBy:   inviterID,
		Token:       token,
		ExpiresAt:   time.Now().Add(7 * 24 * time.Hour),
	})
	if err != nil {
		return nil, fmt.Errorf("create invitation: %w", err)
	}

	result := toInvitationModel(inv)
	return &result, nil
}

// AcceptInvitation accepts an invitation token and adds the user to the household.
func (s *HouseholdService) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) error {
	inv, err := s.queries.GetInvitationByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvitationInvalid
		}
		return fmt.Errorf("get invitation: %w", err)
	}

	if inv.Status != db.InvitationStatusPending {
		return ErrInvitationInvalid
	}
	if inv.ExpiresAt.Time.Before(time.Now()) {
		return ErrInvitationInvalid
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	qtx := s.queries.WithTx(tx)

	err = qtx.AddHouseholdMember(ctx, db.AddHouseholdMemberParams{
		HouseholdID: inv.HouseholdID,
		UserID:      userID,
		Role:        db.HouseholdRoleMember,
	})
	if err != nil {
		return fmt.Errorf("add member: %w", err)
	}

	err = qtx.AcceptInvitation(ctx, inv.ID)
	if err != nil {
		return fmt.Errorf("accept invitation: %w", err)
	}

	return tx.Commit(ctx)
}

// CheckMembership verifies the user is a member of the household.
func (s *HouseholdService) CheckMembership(ctx context.Context, householdID, userID uuid.UUID) error {
	isMember, err := s.queries.IsHouseholdMember(ctx, db.IsHouseholdMemberParams{
		HouseholdID: householdID,
		UserID:      userID,
	})
	if err != nil {
		return fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return ErrNotMember
	}
	return nil
}

func toHouseholdModel(h db.Household) model.Household {
	return model.Household{
		ID:        h.ID,
		Name:      h.Name,
		OwnerID:   h.OwnerID,
		CreatedAt: h.CreatedAt.Time,
	}
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
