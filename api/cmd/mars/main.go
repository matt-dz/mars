package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"mars/internal/api"
	"mars/internal/env"
	marshttp "mars/internal/http"
	marslog "mars/internal/log"
	"mars/internal/mars"
	"mars/internal/setup"
)

const (
	defaultPort            uint16 = 8080
	spotifyRefreshInterval        = 30 * time.Minute
)

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := marslog.New(nil)

	errCh := make(chan error, 1)
	go func() {
		errCh <- run(ctx, logger)
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Fatal(err)
		}
	case <-ctx.Done():
		// Wait for run to finish after context cancellation
		if err := <-errCh; err != nil {
			log.Fatal(err)
		}
	}
}

func run(ctx context.Context, logger *slog.Logger) error {
	db, err := setup.Database(ctx)
	if err != nil {
		return fmt.Errorf("setting up database: %w", err)
	}

	err = setup.Admin(ctx, db, logger)
	if err != nil {
		return fmt.Errorf("setting up admin: %w", err)
	}

	e := env.New()
	e.Logger = logger
	e.Database = db
	e.HTTP = marshttp.New()
	e.HTTP.Logger = logger

	err = setup.AppSecret(e)
	if err != nil {
		return fmt.Errorf("setting up app secret: %w", err)
	}

	port := defaultPort
	if portStr := os.Getenv("PORT"); portStr != "" {
		p, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			log.Fatalf("invalid PORT value: %v", err)
		}
		port = uint16(p)
	}

	// Start Spotify token refresh goroutine
	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")
	go runSpotifyTokenRefresh(ctx, logger, *e.HTTP, adminEmail, adminPassword)

	return api.Start(ctx, port, e)
}

// runSpotifyTokenRefresh waits a specified interval before refreshing all user spotify tokens.
func runSpotifyTokenRefresh(ctx context.Context, logger *slog.Logger, client marshttp.Client, email, password string) {
	ticker := time.NewTicker(spotifyRefreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("stopping spotify token refresh goroutine")
			return
		case <-ticker.C:
			if err := mars.RefreshSpotifyTokens(ctx, client, email, password); err != nil {
				logger.Error("failed to refresh spotify tokens", "error", err)
			} else {
				logger.Info("refreshed spotify tokens")
			}
		}
	}
}
