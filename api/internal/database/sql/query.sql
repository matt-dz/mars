-- name: CheckUsersTableExists :one
SELECT
  EXISTS (
    SELECT
      1
    FROM
      information_schema.tables
    WHERE
      table_schema = 'public'
      AND table_name = 'users');

-- name: Ping :exec
SELECT
  1;

-- name: CreateUser :one
INSERT INTO users (email, role, password_hash)
  VALUES (trim(lower(@email::text)), 'user', $1)
RETURNING
  id;

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
