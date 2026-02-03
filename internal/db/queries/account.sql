-- name: CreateAccount :one
INSERT INTO accounts (
  user_id, balance, currency
) VALUES (
  $1, $2, $3
)
RETURNING id, user_id, balance, currency, created_at;

-- name: GetAccount :one
SELECT id, user_id, balance, currency, created_at FROM accounts
WHERE id = $1 LIMIT 1;

-- name: GetAccountForUpdate :one
SELECT id, user_id, balance, currency, created_at FROM accounts
WHERE id = $1 LIMIT 1
FOR NO KEY UPDATE;

-- name: ListAccounts :many
SELECT id, user_id, balance, currency, created_at FROM accounts
WHERE user_id = $1
ORDER BY id
LIMIT $2 OFFSET $3;

-- name: UpdateAccount :one
UPDATE accounts
SET balance = $2
WHERE id = $1
RETURNING id, user_id, balance, currency, created_at;

-- name: AddAccountBalance :one
UPDATE accounts
SET balance = balance + sqlc.arg(amount)
WHERE id = sqlc.arg(id)
RETURNING id, user_id, balance, currency, created_at;

-- name: DeleteAccount :exec
DELETE FROM accounts
WHERE id = $1;
