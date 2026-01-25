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
  id = $3;

-- name: GetUserRefreshToken :one
SELECT
  refresh_token_hash,
  refresh_token_expires_at
FROM
  users
WHERE
  id = $1;

-- name: UpdateUserSpotifyID :exec
UPDATE
  users
SET
  spotify_id = $1
WHERE
  id = $2;

-- name: UpsertUserSpotifyTokens :exec
INSERT INTO spotify_tokens (spotify_user_id, access_token, token_type, scope, refresh_token, expires_at)
  VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (spotify_user_id)
  DO UPDATE SET
    access_token = EXCLUDED.access_token,
    token_type = EXCLUDED.token_type,
    scope = EXCLUDED.scope,
    refresh_token = EXCLUDED.refresh_token,
    expires_at = EXCLUDED.expires_at;
