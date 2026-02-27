-- name: CreateInvitation :one
INSERT INTO invitations (household_id, email, invited_by, token, expires_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: GetInvitationByToken :one
SELECT * FROM invitations WHERE token = $1;

-- name: AcceptInvitation :exec
UPDATE invitations
SET status = 'accepted'
WHERE id = $1;

-- name: ExpireOldInvitations :exec
UPDATE invitations
SET status = 'expired'
WHERE status = 'pending' AND expires_at < now();

-- name: ListPendingInvitations :many
SELECT * FROM invitations
WHERE household_id = $1 AND status = 'pending'
ORDER BY created_at DESC;
