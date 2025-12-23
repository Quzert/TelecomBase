-- name: CreateUserAutoAdmin :one
INSERT INTO users(username, password_hash, role, approved)
VALUES(
  $1,
  $2,
  CASE
    WHEN (SELECT COUNT(*) FROM users WHERE role = 'admin') = 0 THEN 'admin'
    ELSE 'user'
  END,
  CASE
    WHEN (SELECT COUNT(*) FROM users WHERE role = 'admin') = 0 THEN TRUE
    ELSE FALSE
  END
)
RETURNING role, approved;

-- name: GetUserAuthByUsername :one
SELECT password_hash, role, approved
FROM users
WHERE username = $1;

-- name: GetUserRoleApprovedByUsername :one
SELECT role, approved
FROM users
WHERE username = $1;

-- name: ListPendingUsers :many
SELECT id, username, role, created_at
FROM users
WHERE approved = FALSE
ORDER BY created_at;

-- name: ApproveUserByID :exec
UPDATE users
SET approved = TRUE
WHERE id = $1;

-- name: ListUsers :many
SELECT id, username, role, approved, created_at
FROM users
ORDER BY created_at;

-- name: GetUserRoleByID :one
SELECT role
FROM users
WHERE id = $1;

-- name: SetUserApprovedByID :exec
UPDATE users
SET approved = $2
WHERE id = $1;

-- name: GetUserUsernameRoleByID :one
SELECT username, role
FROM users
WHERE id = $1;

-- name: DeleteUserByID :exec
DELETE FROM users
WHERE id = $1;
