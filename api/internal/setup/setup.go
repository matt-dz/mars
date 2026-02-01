// Package setup contains setup functions for the api
package setup

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"mars/internal/admin"
	"mars/internal/database"
	"mars/internal/env"
	"mars/internal/service"

	"github.com/jackc/pgx/v5/pgxpool"
)

func AppSecret(env *env.Env) error {
	const secretPath = "/data/secret"
	const appSecretFilePerms = 0o600
	const dataDirectoryPerms = 0o755
	const appSecretBytes = 64
	var secret string

	if f1, err := os.Lstat(secretPath); err != nil {
		if !errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("checking secret path: %w", err)
		}

		// Create secret
		bytes := make([]byte, appSecretBytes)
		if _, err := rand.Reader.Read(bytes); err != nil {
			return fmt.Errorf("creating app secret: %w", err)
		}
		secret = base64.URLEncoding.EncodeToString(bytes)

		// Write file
		err = os.Mkdir("/data", dataDirectoryPerms)
		if err != nil {
			return fmt.Errorf("creating data directory: %w", err)
		}
		err = os.WriteFile(secretPath, []byte(secret), appSecretFilePerms)
		if err != nil {
			return fmt.Errorf("writing app secret: %w", err)
		}
	} else {
		if f1.IsDir() {
			return fmt.Errorf("expected file, got directory at %q", secretPath)
		}
		data, err := os.ReadFile(secretPath)
		if err != nil {
			return fmt.Errorf("reading file: %w", err)
		}
		secret = string(data)
	}

	env.Set("APP_SECRET", secret)
	return nil
}

func Database(ctx context.Context) (*database.Queries, *pgxpool.Pool, error) {
	databaseHost := os.Getenv("DATABASE_HOST")
	if databaseHost == "" {
		return nil, nil, errors.New("DATABASE_HOST environment variable is required")
	}
	databasePort := os.Getenv("DATABASE_PORT")
	if databasePort == "" {
		return nil, nil, errors.New("DATABASE_PORT environment variable is required")
	}
	databaseUser := os.Getenv("DATABASE_USER")
	if databaseUser == "" {
		return nil, nil, errors.New("DATABASE_USER environment variable is required")
	}
	databasePassword := os.Getenv("DATABASE_PASSWORD")
	if databasePassword == "" {
		return nil, nil, errors.New("DATABASE_PASSWORD environment variable is required")
	}
	databaseName := os.Getenv("DATABASE_NAME")
	if databaseName == "" {
		return nil, nil, errors.New("DATABASE_NAME environment variable is required")
	}

	poolConfig, err := pgxpool.ParseConfig("")
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create database config: %w", err)
	}

	poolConfig.ConnConfig.Host = databaseHost
	poolConfig.ConnConfig.Port, err = func() (uint16, error) {
		p, err := strconv.ParseUint(databasePort, 10, 16)
		if err != nil {
			return 0, fmt.Errorf("failed to parse database port: %w", err)
		}
		return uint16(p), nil
	}()
	if err != nil {
		return nil, nil, err
	}
	poolConfig.ConnConfig.User = databaseUser
	poolConfig.ConnConfig.Password = databasePassword
	poolConfig.ConnConfig.Database = databaseName

	// Creating DB connection
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create database connection: %w", err)
	}

	db := database.New(pool)
	if err := db.Ping(ctx); err != nil {
		return nil, nil, fmt.Errorf("failed to ping database: %w", err)
	}
	if err := database.ApplySchema(ctx, pool); err != nil {
		return nil, nil, fmt.Errorf("failed to apply database schema: %w", err)
	}

	return db, pool, nil
}

func Admin(ctx context.Context, db database.Querier, logger *slog.Logger) error {
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if err := admin.SeedAdmin(ctx, db, logger, adminEmail, adminPassword); err != nil {
		return fmt.Errorf("seeding admin user: %w", err)
	}
	return nil
}

func ServiceAccount(ctx context.Context, db database.Querier, logger *slog.Logger) (email, password string, err error) {
	email, password, err = service.LoadOrCreateCredentials(logger)
	if err != nil {
		return "", "", fmt.Errorf("loading service account credentials: %w", err)
	}

	if err := service.SeedServiceAccount(ctx, db, logger, email, password); err != nil {
		return "", "", fmt.Errorf("seeding service account: %w", err)
	}

	return email, password, nil
}
