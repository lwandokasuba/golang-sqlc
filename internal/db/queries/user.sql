-- name: CreateUser :one
INSERT INTO users (
  username, email, hashed_password
) VALUES (
  $1, $2, $3
)
RETURNING id, username, email, hashed_password, created_at;

-- name: GetUser :one
SELECT id, username, email, hashed_password, created_at FROM users
WHERE id = $1 LIMIT 1;

-- name: GetUserWithAccounts :many
SELECT 
    u.id AS user_id, 
    u.username, 
    u.email, 
    u.hashed_password, 
    u.created_at AS user_created_at,
    a.id AS account_id, 
    a.balance, 
    a.currency, 
    a.created_at AS account_created_at
FROM users u
LEFT JOIN accounts a ON a.user_id = u.id
WHERE u.id = $1;

-- name: ListUsers :many
SELECT id, username, email, hashed_password, created_at FROM users
ORDER BY id
LIMIT $1 OFFSET $2;
