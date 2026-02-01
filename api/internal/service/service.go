// Package service provides service account management functionality.
package service

import (
	"context"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"mars/internal/argon2id"
	"mars/internal/database"
)

const (
	defaultServiceEmail    = "service@mars.com"
	serviceSecretPath      = "/data/service_credentials"
	serviceSecretFilePerms = 0o600
	servicePasswordBytes   = 32
)

var ErrPartialCredentials = errors.New(
	"service account credentials incomplete: both SERVICE_EMAIL and SERVICE_PASSWORD must be set, or neither")

type serviceCredentials struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// LoadOrCreateCredentials loads service account credentials from environment or secret file,
// or generates new ones if none exist. Returns the email and password.
func LoadOrCreateCredentials(logger *slog.Logger) (email, password string, err error) {
	envEmail := os.Getenv("SERVICE_EMAIL")
	envPassword := os.Getenv("SERVICE_PASSWORD")

	// If both env vars are set, use them
	if envEmail != "" && envPassword != "" {
		logger.Info("using service account credentials from environment")
		return envEmail, envPassword, nil
	}

	// If only one is set, error
	if envEmail != "" || envPassword != "" {
		return "", "", ErrPartialCredentials
	}

	// Try to load from secret file
	if data, err := os.ReadFile(serviceSecretPath); err == nil {
		var creds serviceCredentials
		if err := json.Unmarshal(data, &creds); err != nil {
			return "", "", fmt.Errorf("parsing service credentials file: %w", err)
		}
		logger.Info("using service account credentials from secret file")
		// Set as environment variables for use by other components
		if err = os.Setenv("SERVICE_EMAIL", creds.Email); err != nil {
			return "", "", fmt.Errorf("setting SERVICE_EMAIL: %w", err)
		}
		if err = os.Setenv("SERVICE_PASSWORD", creds.Password); err != nil {
			return "", "", fmt.Errorf("setting SERVICE_PASSWORD: %w", err)
		}
		return creds.Email, creds.Password, nil
	} else if !errors.Is(err, os.ErrNotExist) {
		return "", "", fmt.Errorf("reading service credentials file: %w", err)
	}

	// Generate new credentials
	logger.Info("generating new service account credentials")
	email = defaultServiceEmail

	passwordBytes := make([]byte, servicePasswordBytes)
	if _, err := rand.Read(passwordBytes); err != nil {
		return "", "", fmt.Errorf("generating service account password: %w", err)
	}
	password = base64.URLEncoding.EncodeToString(passwordBytes)

	// Save to secret file
	creds := serviceCredentials{
		Email:    email,
		Password: password,
	}
	data, err := json.Marshal(creds)
	if err != nil {
		return "", "", fmt.Errorf("marshaling service credentials: %w", err)
	}
	if err := os.WriteFile(serviceSecretPath, data, serviceSecretFilePerms); err != nil {
		return "", "", fmt.Errorf("writing service credentials file: %w", err)
	}

	// Set as environment variables for use by other components
	if err = os.Setenv("SERVICE_EMAIL", email); err != nil {
		return "", "", fmt.Errorf("setting SERVICE_EMAIL: %w", err)
	}
	if err = os.Setenv("SERVICE_PASSWORD", password); err != nil {
		return "", "", fmt.Errorf("setting SERVICE_PASSWORD: %w", err)
	}

	logger.Info("service account credentials saved to secret file")
	return email, password, nil
}

// SeedServiceAccount creates a service account if none exists.
func SeedServiceAccount(ctx context.Context, db database.Querier, logger *slog.Logger, email, password string) error {
	exists, err := db.ServiceAccountExists(ctx)
	if err != nil {
		return fmt.Errorf("checking if service account exists: %w", err)
	}

	if exists {
		logger.InfoContext(ctx, "service account already exists, skipping seed")
		return nil
	}

	passwordHash, err := argon2id.HashAndEncode(password, argon2id.DefaultParams)
	if err != nil {
		return fmt.Errorf("hashing service account password: %w", err)
	}

	_, err = db.CreateServiceAccount(ctx, database.CreateServiceAccountParams{
		Email:        email,
		PasswordHash: passwordHash,
	})
	if err != nil {
		return fmt.Errorf("creating service account: %w", err)
	}

	logger.InfoContext(ctx, "service account created successfully", slog.String("email", email))
	return nil
}
