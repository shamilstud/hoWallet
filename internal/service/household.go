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

	"github.com/howallet/howallet/internal/model"
	"github.com/howallet/howallet/internal/repository/postgres"
)

var (
	ErrHouseholdNotFound = errors.New("household not found")
	ErrNotHouseholdOwner = errors.New("only household owner can perform this action")
	ErrNotMember         = errors.New("user is not a member of this household")
	ErrInvitationInvalid = errors.New("invitation is invalid or expired")
	ErrAlreadyMember     = errors.New("user is already a member")
)

type HouseholdService struct {
	repos       *postgres.Repos
	emailSvc    *EmailService
	frontendURL string
}

func NewHouseholdService(repos *postgres.Repos, emailSvc *EmailService, frontendURL string) *HouseholdService {
	return &HouseholdService{repos: repos, emailSvc: emailSvc, frontendURL: frontendURL}
}

func (s *HouseholdService) Create(ctx context.Context, userID uuid.UUID, req model.CreateHouseholdRequest) (*model.Household, error) {
	var hh model.Household
	err := s.repos.RunInTx(ctx, func(txCtx context.Context) error {
		txRepos := postgres.TxReposFromCtx(txCtx)

		var txErr error
		hh, txErr = txRepos.Households.Create(txCtx, req.Name, userID)
		if txErr != nil {
			return fmt.Errorf("create household: %w", txErr)
		}

		txErr = txRepos.Households.AddMember(txCtx, hh.ID, userID, model.HouseholdRoleOwner)
		if txErr != nil {
			return fmt.Errorf("add owner: %w", txErr)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return &hh, nil
}

func (s *HouseholdService) List(ctx context.Context, userID uuid.UUID) ([]model.Household, error) {
	list, err := s.repos.Households.ListByUser(ctx, userID)
	if err != nil {
		return nil, fmt.Errorf("list households: %w", err)
	}
	return list, nil
}

func (s *HouseholdService) Get(ctx context.Context, id uuid.UUID) (*model.Household, error) {
	hh, err := s.repos.Households.GetByID(ctx, id)
	if err != nil {
		return nil, ErrHouseholdNotFound
	}
	return &hh, nil
}

func (s *HouseholdService) ListMembers(ctx context.Context, householdID uuid.UUID) ([]model.HouseholdMember, error) {
	members, err := s.repos.Households.ListMembers(ctx, householdID)
	if err != nil {
		return nil, fmt.Errorf("list members: %w", err)
	}
	return members, nil
}

func (s *HouseholdService) RemoveMember(ctx context.Context, householdID, ownerID, targetUserID uuid.UUID) error {
	member, err := s.repos.Households.GetMember(ctx, householdID, ownerID)
	if err != nil {
		return ErrNotMember
	}
	if member.Role != model.HouseholdRoleOwner {
		return ErrNotHouseholdOwner
	}

	return s.repos.Households.RemoveMember(ctx, householdID, targetUserID)
}

// Invite creates an invitation token for the given email.
func (s *HouseholdService) Invite(ctx context.Context, householdID, inviterID uuid.UUID, email string) (*model.Invitation, error) {
	// Verify inviter is owner
	member, err := s.repos.Households.GetMember(ctx, householdID, inviterID)
	if err != nil {
		return nil, ErrNotMember
	}
	if member.Role != model.HouseholdRoleOwner {
		return nil, ErrNotHouseholdOwner
	}

	// Check if already a member
	existingUser, err := s.repos.Users.GetByEmail(ctx, email)
	if err == nil {
		isMember, _ := s.repos.Households.IsMember(ctx, householdID, existingUser.ID)
		if isMember {
			return nil, ErrAlreadyMember
		}
	}

	// Generate token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return nil, fmt.Errorf("generate token: %w", err)
	}
	token := hex.EncodeToString(tokenBytes)

	inv, err := s.repos.Invitations.Create(ctx, householdID, inviterID, email, token, time.Now().Add(7*24*time.Hour))
	if err != nil {
		return nil, fmt.Errorf("create invitation: %w", err)
	}

	// Send invitation email (best-effort: don't fail the invite if email fails)
	if s.emailSvc != nil && s.emailSvc.cfg.Host != "" {
		hh, _ := s.repos.Households.GetByID(ctx, householdID)
		inviter, _ := s.repos.Users.GetByID(ctx, inviterID)
		inviterName := "A hoWallet user"
		if inviter.Name != "" {
			inviterName = inviter.Name
		}
		_ = s.emailSvc.SendInvitation(email, hh.Name, inviterName, token, s.frontendURL)
	}

	return &inv, nil
}

// AcceptInvitation accepts an invitation token and adds the user to the household.
func (s *HouseholdService) AcceptInvitation(ctx context.Context, token string, userID uuid.UUID) error {
	inv, err := s.repos.Invitations.GetByToken(ctx, token)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrInvitationInvalid
		}
		return fmt.Errorf("get invitation: %w", err)
	}

	if inv.Status != model.InvitationStatusPending {
		return ErrInvitationInvalid
	}
	if inv.ExpiresAt.Before(time.Now()) {
		return ErrInvitationInvalid
	}

	return s.repos.RunInTx(ctx, func(txCtx context.Context) error {
		txRepos := postgres.TxReposFromCtx(txCtx)

		if err := txRepos.Households.AddMember(txCtx, inv.HouseholdID, userID, model.HouseholdRoleMember); err != nil {
			return fmt.Errorf("add member: %w", err)
		}

		if err := txRepos.Invitations.Accept(txCtx, inv.ID); err != nil {
			return fmt.Errorf("accept invitation: %w", err)
		}

		return nil
	})
}

// CheckMembership verifies the user is a member of the household.
func (s *HouseholdService) CheckMembership(ctx context.Context, householdID, userID uuid.UUID) error {
	isMember, err := s.repos.Households.IsMember(ctx, householdID, userID)
	if err != nil {
		return fmt.Errorf("check membership: %w", err)
	}
	if !isMember {
		return ErrNotMember
	}
	return nil
}

// ListPendingInvitations returns pending invitations for a household.
func (s *HouseholdService) ListPendingInvitations(ctx context.Context, householdID uuid.UUID) ([]model.Invitation, error) {
	return s.repos.Invitations.ListPendingByHousehold(ctx, householdID)
}
