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

-- name: ServiceAccountExists :one
SELECT
  EXISTS (
    SELECT
      1
    FROM
      users
    WHERE
      ROLE = 'service');

-- name: CreateServiceAccount :one
INSERT INTO users (email, role, password_hash)
  VALUES (trim(lower(@email::text)), 'service', $1)
RETURNING
  id;

-- name: GetUserByEmail :one
SELECT
  id,
  email,
  ROLE,
  password_hash
FROM
  users
WHERE
  email = trim(lower(@email::text));

-- name: GetUserRole :one
SELECT
  ROLE
FROM
  users
WHERE
  id = $1;

-- name: GetUser :one
SELECT
  email,
  ROLE
FROM
  users
WHERE
  id = $1;

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

-- name: GetUserSpotifyTokenExpiration :one
SELECT
  st.expires_at
FROM
  users u
  JOIN spotify_tokens st ON st.spotify_user_id = u.spotify_id
WHERE
  u.spotify_id IS NOT NULL
  AND u.id = $1;

-- name: GetUserSpotifyRefreshToken :one
SELECT
  st.refresh_token
FROM
  users u
  JOIN spotify_tokens st ON st.spotify_user_id = u.spotify_id
WHERE
  u.spotify_id IS NOT NULL
  AND u.id = $1;

-- name: GetUserSpotifyAccessToken :one
SELECT
  st.access_token
FROM
  users u
  JOIN spotify_tokens st ON st.spotify_user_id = u.spotify_id
WHERE
  u.spotify_id IS NOT NULL
  AND u.id = $1;

-- name: GetUserSpotifyId :one
SELECT
  u.spotify_id
FROM
  users u
  JOIN spotify_tokens st ON st.spotify_user_id = u.spotify_id
WHERE
  u.spotify_id IS NOT NULL
  AND u.id = $1;

-- name: UpdateUserSpotifyTokens :exec
UPDATE
  spotify_tokens
SET
  access_token = $1,
  refresh_token = $2,
  token_type = $3,
  scope = $4,
  expires_at = $5
WHERE
  spotify_user_id = $6;

-- name: GetUserIDs :many
SELECT
  id
FROM
  users
ORDER BY
  created_at ASC
LIMIT $1;

-- name: UpsertTrack :exec
INSERT INTO tracks (image_url, id, name, artists, href, uri)
  VALUES (sqlc.narg ('image_url'), $1, $2, $3, $4, $5)
ON CONFLICT (id)
  DO UPDATE SET
    updated_at = NOW(),
    image_url = EXCLUDED.image_url,
    name = EXCLUDED.name,
    artists = EXCLUDED.artists,
    href = EXCLUDED.href;

-- name: UpsertTrackListen :exec
INSERT INTO track_listens (user_id, track_id, played_at)
  VALUES ($1, $2, $3)
ON CONFLICT (user_id, track_id, played_at)
  DO NOTHING;

-- name: ListensByTrackInRange :many
SELECT
  track_id,
  COUNT(*)::bigint AS listen_count
FROM
  track_listens
WHERE
  user_id = $1
  AND played_at >= @start_date::timestamptz
  AND played_at < @end_date::timestamptz
GROUP BY
  track_id
ORDER BY
  listen_count DESC,
  track_id ASC
LIMIT 50;

-- name: CreatePlaylist :one
INSERT INTO playlists (user_id, playlist_type, name)
  VALUES ($1, $2, $3)
RETURNING
  id;

-- name: AddPlaylistTrack :exec
INSERT INTO playlist_tracks (playlist_id, track_id, plays)
  VALUES ($1, $2, $3);

-- name: GetUserPlaylists :many
SELECT
  id,
  playlist_type,
  name,
  created_at
FROM
  playlists
WHERE
  user_id = $1
ORDER BY
  created_at DESC;

-- name: GetUserPlaylist :one
SELECT
  id,
  playlist_type,
  name,
  created_at
FROM
  playlists
WHERE
  user_id = $1
  AND id = $2;

-- name: GetPlaylistTracks :many
SELECT
  t.id,
  t.name,
  t.artists,
  t.href,
  t.image_url,
  t.uri,
  pt.plays
FROM
  playlist_tracks pt
  JOIN tracks t ON pt.track_id = t.id
WHERE
  pt.playlist_id = $1
ORDER BY
  pt.plays DESC,
  pt.track_id ASC;
