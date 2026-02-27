-- name: CreateHousehold :one
INSERT INTO households (name, owner_id)
VALUES ($1, $2)
RETURNING *;

-- name: GetHousehold :one
SELECT * FROM households WHERE id = $1;

-- name: ListUserHouseholds :many
SELECT h.*
FROM households h
JOIN household_members hm ON hm.household_id = h.id
WHERE hm.user_id = $1
ORDER BY h.created_at;

-- name: UpdateHousehold :one
UPDATE households
SET name = $2
WHERE id = $1
RETURNING *;

-- name: DeleteHousehold :exec
DELETE FROM households WHERE id = $1;

-- name: AddHouseholdMember :exec
INSERT INTO household_members (household_id, user_id, role)
VALUES ($1, $2, $3)
ON CONFLICT (household_id, user_id) DO NOTHING;

-- name: RemoveHouseholdMember :exec
DELETE FROM household_members
WHERE household_id = $1 AND user_id = $2;

-- name: GetHouseholdMember :one
SELECT * FROM household_members
WHERE household_id = $1 AND user_id = $2;

-- name: ListHouseholdMembers :many
SELECT hm.*, u.email, u.name AS user_name
FROM household_members hm
JOIN users u ON u.id = hm.user_id
WHERE hm.household_id = $1
ORDER BY hm.joined_at;

-- name: IsHouseholdMember :one
SELECT EXISTS (
    SELECT 1 FROM household_members
    WHERE household_id = $1 AND user_id = $2
) AS is_member;
