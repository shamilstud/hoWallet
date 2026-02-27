-- name: CreateAccount :one
INSERT INTO accounts (household_id, name, type, balance, currency, created_by)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: GetAccount :one
SELECT * FROM accounts WHERE id = $1 AND household_id = $2;

-- name: ListAccountsByHousehold :many
SELECT * FROM accounts
WHERE household_id = $1
ORDER BY created_at;

-- name: UpdateAccount :one
UPDATE accounts
SET name     = COALESCE(sqlc.narg('name'), name),
    type     = COALESCE(sqlc.narg('type'), type),
    currency = COALESCE(sqlc.narg('currency'), currency)
WHERE id = $1 AND household_id = $2
RETURNING *;

-- name: UpdateAccountBalance :exec
UPDATE accounts
SET balance = balance + $2
WHERE id = $1;

-- name: SetAccountBalance :exec
UPDATE accounts
SET balance = $2
WHERE id = $1;

-- name: DeleteAccount :exec
DELETE FROM accounts WHERE id = $1 AND household_id = $2;

-- name: CountTransactionsByAccount :one
SELECT COUNT(*) FROM transactions WHERE account_id = $1;
