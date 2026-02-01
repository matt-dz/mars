DO $$
BEGIN
  CREATE TYPE ROLE AS enum (
    'admin',
    'user',
    'service'
);
EXCEPTION
  WHEN duplicate_object THEN
    NULL;
END
$$;

CREATE TABLE IF NOT EXISTS users (
  id uuid DEFAULT gen_random_uuid () PRIMARY KEY,
  email text NOT NULL,
  ROLE ROLE NOT NULL DEFAULT 'user',
  password_hash text NOT NULL,
  spotify_id text UNIQUE,
  refresh_token_hash text,
  refresh_token_expires_at timestamptz,
  created_at timestamptz NOT NULL DEFAULT now(),
  updated_at timestamptz NOT NULL DEFAULT now(),
  CHECK ((refresh_token_hash IS NULL AND refresh_token_expires_at IS NULL) OR (refresh_token_hash IS NOT NULL AND
    refresh_token_expires_at IS NOT NULL))
);

CREATE UNIQUE INDEX IF NOT EXISTS users_unique_email ON users (trim(lower(email)))
WHERE
  email IS NOT NULL;

CREATE TABLE IF NOT EXISTS tracks (
  id text PRIMARY KEY,
  name text NOT NULL,
  artists text[] NOT NULL,
  href text NOT NULL,
  image_url text,
  created_at timestamptz NOT NULL DEFAULT NOW(),
  updated_at timestamptz NOT NULL DEFAULT now()
);

DO $$
BEGIN
  CREATE TYPE playlist_type AS enum (
    'weekly',
    'monthly'
);
EXCEPTION
  WHEN duplicate_object THEN
    NULL;
END
$$;

CREATE TABLE IF NOT EXISTS playlists (
  id uuid DEFAULT gen_random_uuid (),
  user_id uuid,
  playlist_type playlist_type NOT NULL,
  name text NOT NULL,
  created_at timestamptz NOT NULL DEFAULT now(),
  PRIMARY KEY (id),
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS playlist_tracks (
  playlist_id uuid NOT NULL,
  track_id text NOT NULL,
  plays integer NOT NULL,
  CONSTRAINT positive_plays CHECK (plays > 0),
  PRIMARY KEY (playlist_id, track_id),
  FOREIGN KEY (playlist_id) REFERENCES playlists (id) ON DELETE CASCADE,
  FOREIGN KEY (track_id) REFERENCES tracks (id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS track_listens (
  user_id uuid NOT NULL,
  track_id text NOT NULL,
  played_at timestamptz NOT NULL,
  PRIMARY KEY (user_id, track_id, played_at),
  FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE CASCADE,
  FOREIGN KEY (track_id) REFERENCES tracks (id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_track_listens_user_played_at ON track_listens (user_id, played_at);

CREATE TABLE IF NOT EXISTS spotify_tokens (
  spotify_user_id text PRIMARY KEY,
  access_token text NOT NULL,
  token_type text NOT NULL,
  scope text NOT NULL,
  refresh_token text NOT NULL,
  expires_at timestamptz NOT NULL,
  FOREIGN KEY (spotify_user_id) REFERENCES users (spotify_id) ON DELETE CASCADE
);
