// Package admin provides admin user management functionality.
package admin

import (
	"context"
	"errors"
	"fmt"
	"log/slog"

	"mars/internal/argon2id"
	"mars/internal/database"
)

var (
	ErrMissingEmail    = errors.New("admin credentials required: ADMIN_EMAIL must be set when no admin exists")
	ErrMissingPassword = errors.New("admin credentials required: ADMIN_PASSWORD must be set when no admin exists")
)

// SeedAdmin creates an admin user if none exists and credentials are provided.
// Returns an error if no admin exists and credentials are missing.
func SeedAdmin(ctx context.Context, db database.Querier, logger *slog.Logger, email, password string) error {
	exists, err := db.AdminExists(ctx)
	if err != nil {
		return fmt.Errorf("checking if admin exists: %w", err)
	}

	if exists {
		logger.InfoContext(ctx, "admin user already exists, skipping seed")
		return nil
	}

	if email == "" {
		return ErrMissingEmail
	}

	if password == "" {
		return ErrMissingPassword
	}

	passwordHash, err := argon2id.HashAndEncode(password, argon2id.DefaultParams)
	if err != nil {
		return fmt.Errorf("hashing admin password: %w", err)
	}

	_, err = db.CreateAdminUser(ctx, database.CreateAdminUserParams{
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return fmt.Errorf("creating admin user: %w", err)
	}

	logger.InfoContext(ctx, "admin user created successfully", slog.String("email", email))
	return nil
}
