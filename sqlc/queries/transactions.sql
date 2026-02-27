-- name: CreateTransaction :one
INSERT INTO transactions (
    household_id, type, description, amount,
    account_id, destination_account_id, tags, note,
    transacted_at, created_by
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
RETURNING *;

-- name: GetTransaction :one
SELECT * FROM transactions
WHERE id = $1 AND household_id = $2;

-- name: ListTransactions :many
SELECT * FROM transactions
WHERE household_id = $1
  AND ($2::timestamptz IS NULL OR transacted_at >= $2)
  AND ($3::timestamptz IS NULL OR transacted_at <= $3)
  AND ($4::transaction_type IS NULL OR type = $4)
  AND ($5::uuid IS NULL OR account_id = $5 OR destination_account_id = $5)
ORDER BY transacted_at DESC
LIMIT $6 OFFSET $7;

-- name: CountTransactions :one
SELECT COUNT(*) FROM transactions
WHERE household_id = $1
  AND ($2::timestamptz IS NULL OR transacted_at >= $2)
  AND ($3::timestamptz IS NULL OR transacted_at <= $3)
  AND ($4::transaction_type IS NULL OR type = $4)
  AND ($5::uuid IS NULL OR account_id = $5 OR destination_account_id = $5);

-- name: UpdateTransaction :one
UPDATE transactions
SET description            = $3,
    amount                 = $4,
    account_id             = $5,
    destination_account_id = $6,
    tags                   = $7,
    note                   = $8,
    transacted_at          = $9,
    type                   = $10
WHERE id = $1 AND household_id = $2
RETURNING *;

-- name: DeleteTransaction :one
DELETE FROM transactions
WHERE id = $1 AND household_id = $2
RETURNING *;

-- name: ListTransactionsForExport :many
SELECT
    t.transacted_at,
    t.description,
    t.amount,
    t.type,
    t.tags,
    t.note,
    a.name  AS account_name,
    a.currency AS account_currency,
    da.name AS destination_account_name
FROM transactions t
JOIN accounts a ON a.id = t.account_id
LEFT JOIN accounts da ON da.id = t.destination_account_id
WHERE t.household_id = $1
  AND ($2::timestamptz IS NULL OR t.transacted_at >= $2)
  AND ($3::timestamptz IS NULL OR t.transacted_at <= $3)
ORDER BY t.transacted_at DESC;
