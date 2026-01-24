-- name: Ping :exec
SELECT
  1;

-- name: AdminExists :one
SELECT
  EXISTS (
    SELECT
      1
    FROM
      users
    WHERE
      ROLE = 'admin');

-- name: CreateAdminUser :one
INSERT INTO users (email, role, password_hash)
  VALUES (trim(lower(@email::text)), 'admin', $1)
RETURNING
  id;

-- name: GetUserByEmail :one
SELECT
  id,
  email,
  password_hash
FROM
  users
WHERE
  email = trim(lower(@email::text));

-- name: UpdateUserRefreshToken :exec
UPDATE
  users
SET
  refresh_token_hash = $1,
  refresh_token_expires_at = $2
WHERE
  id = $1;
