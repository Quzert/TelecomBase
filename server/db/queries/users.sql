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

SELECT password_hash, role, approved
FROM users
WHERE username = $1;

SELECT role, approved
FROM users
WHERE username = $1;

SELECT id, username, role, created_at
FROM users
WHERE approved = FALSE
ORDER BY created_at;

UPDATE users
SET approved = TRUE
WHERE id = $1;

SELECT id, username, role, approved, created_at
FROM users
ORDER BY created_at;

SELECT role
FROM users
WHERE id = $1;

UPDATE users
SET approved = $2
WHERE id = $1;

SELECT username, role
FROM users
WHERE id = $1;

DELETE FROM users
WHERE id = $1;
