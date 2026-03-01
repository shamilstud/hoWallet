package db

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// --- Households ---

type CreateHouseholdParams struct {
	Name    string
	OwnerID uuid.UUID
}

func (q *Queries) CreateHousehold(ctx context.Context, arg CreateHouseholdParams) (Household, error) {
	row := q.queryRow(ctx,
		`INSERT INTO households (name, owner_id) VALUES ($1, $2) RETURNING id, name, owner_id, created_at`,
		arg.Name, arg.OwnerID,
	)
	var h Household
	err := row.Scan(&h.ID, &h.Name, &h.OwnerID, &h.CreatedAt)
	return h, err
}

func (q *Queries) GetHousehold(ctx context.Context, id uuid.UUID) (Household, error) {
	row := q.queryRow(ctx,
		`SELECT id, name, owner_id, created_at FROM households WHERE id = $1`,
		id,
	)
	var h Household
	err := row.Scan(&h.ID, &h.Name, &h.OwnerID, &h.CreatedAt)
	return h, err
}

func (q *Queries) ListUserHouseholds(ctx context.Context, userID uuid.UUID) ([]Household, error) {
	rows, err := q.query(ctx,
		`SELECT h.id, h.name, h.owner_id, h.created_at
		 FROM households h
		 JOIN household_members hm ON hm.household_id = h.id
		 WHERE hm.user_id = $1
		 ORDER BY h.created_at`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Household
	for rows.Next() {
		var h Household
		if err := rows.Scan(&h.ID, &h.Name, &h.OwnerID, &h.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, h)
	}
	return out, rows.Err()
}

// --- Household Members ---

type AddHouseholdMemberParams struct {
	HouseholdID uuid.UUID
	UserID      uuid.UUID
	Role        HouseholdRole
}

func (q *Queries) AddHouseholdMember(ctx context.Context, arg AddHouseholdMemberParams) error {
	return q.exec(ctx,
		`INSERT INTO household_members (household_id, user_id, role) VALUES ($1, $2, $3) ON CONFLICT (household_id, user_id) DO NOTHING`,
		arg.HouseholdID, arg.UserID, arg.Role,
	)
}

type RemoveHouseholdMemberParams struct {
	HouseholdID uuid.UUID
	UserID      uuid.UUID
}

func (q *Queries) RemoveHouseholdMember(ctx context.Context, arg RemoveHouseholdMemberParams) error {
	return q.exec(ctx,
		`DELETE FROM household_members WHERE household_id = $1 AND user_id = $2`,
		arg.HouseholdID, arg.UserID,
	)
}

type GetHouseholdMemberParams struct {
	HouseholdID uuid.UUID
	UserID      uuid.UUID
}

func (q *Queries) GetHouseholdMember(ctx context.Context, arg GetHouseholdMemberParams) (HouseholdMember, error) {
	row := q.queryRow(ctx,
		`SELECT household_id, user_id, role, joined_at FROM household_members WHERE household_id = $1 AND user_id = $2`,
		arg.HouseholdID, arg.UserID,
	)
	var hm HouseholdMember
	err := row.Scan(&hm.HouseholdID, &hm.UserID, &hm.Role, &hm.JoinedAt)
	return hm, err
}

// ListHouseholdMembersRow includes joined user info.
type ListHouseholdMembersRow struct {
	HouseholdID uuid.UUID
	UserID      uuid.UUID
	Role        HouseholdRole
	JoinedAt    time.Time
	Email       string
	UserName    string
}

func (q *Queries) ListHouseholdMembers(ctx context.Context, householdID uuid.UUID) ([]ListHouseholdMembersRow, error) {
	rows, err := q.query(ctx,
		`SELECT hm.household_id, hm.user_id, hm.role, hm.joined_at, u.email, u.name
		 FROM household_members hm
		 JOIN users u ON u.id = hm.user_id
		 WHERE hm.household_id = $1
		 ORDER BY hm.joined_at`,
		householdID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []ListHouseholdMembersRow
	for rows.Next() {
		var m ListHouseholdMembersRow
		if err := rows.Scan(&m.HouseholdID, &m.UserID, &m.Role, &m.JoinedAt, &m.Email, &m.UserName); err != nil {
			return nil, err
		}
		out = append(out, m)
	}
	return out, rows.Err()
}

type IsHouseholdMemberParams struct {
	HouseholdID uuid.UUID
	UserID      uuid.UUID
}

func (q *Queries) IsHouseholdMember(ctx context.Context, arg IsHouseholdMemberParams) (bool, error) {
	var exists bool
	err := q.queryRow(ctx,
		`SELECT EXISTS (SELECT 1 FROM household_members WHERE household_id = $1 AND user_id = $2)`,
		arg.HouseholdID, arg.UserID,
	).Scan(&exists)
	return exists, err
}

// --- Invitations ---

type CreateInvitationParams struct {
	HouseholdID uuid.UUID
	Email       string
	InvitedBy   uuid.UUID
	Token       string
	ExpiresAt   time.Time
}

func (q *Queries) CreateInvitation(ctx context.Context, arg CreateInvitationParams) (Invitation, error) {
	row := q.queryRow(ctx,
		`INSERT INTO invitations (household_id, email, invited_by, token, expires_at)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id, household_id, email, invited_by, token, status, expires_at, created_at`,
		arg.HouseholdID, arg.Email, arg.InvitedBy, arg.Token, arg.ExpiresAt,
	)
	var inv Invitation
	err := row.Scan(&inv.ID, &inv.HouseholdID, &inv.Email, &inv.InvitedBy, &inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt)
	return inv, err
}

func (q *Queries) GetInvitationByToken(ctx context.Context, token string) (Invitation, error) {
	row := q.queryRow(ctx,
		`SELECT id, household_id, email, invited_by, token, status, expires_at, created_at FROM invitations WHERE token = $1`,
		token,
	)
	var inv Invitation
	err := row.Scan(&inv.ID, &inv.HouseholdID, &inv.Email, &inv.InvitedBy, &inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt)
	if err == pgx.ErrNoRows {
		return inv, err
	}
	return inv, err
}

func (q *Queries) AcceptInvitation(ctx context.Context, id uuid.UUID) error {
	return q.exec(ctx, `UPDATE invitations SET status = 'accepted' WHERE id = $1`, id)
}

func (q *Queries) ListPendingInvitations(ctx context.Context, householdID uuid.UUID) ([]Invitation, error) {
	rows, err := q.query(ctx,
		`SELECT id, household_id, email, invited_by, token, status, expires_at, created_at
		 FROM invitations
		 WHERE household_id = $1 AND status = 'pending' AND expires_at > now()
		 ORDER BY created_at DESC`,
		householdID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var out []Invitation
	for rows.Next() {
		var inv Invitation
		if err := rows.Scan(&inv.ID, &inv.HouseholdID, &inv.Email, &inv.InvitedBy, &inv.Token, &inv.Status, &inv.ExpiresAt, &inv.CreatedAt); err != nil {
			return nil, err
		}
		out = append(out, inv)
	}
	return out, rows.Err()
}
