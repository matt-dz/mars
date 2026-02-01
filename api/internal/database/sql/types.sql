-- Type definitions for sqlc to parse
-- The actual database uses the idempotent definitions in schema.sql

CREATE TYPE role AS ENUM ('admin', 'user');

CREATE TYPE playlist_type AS ENUM ('weekly', 'monthly');
