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

	_ "time/tzdata"
)

const (
	defaultPort              uint16 = 8080
	spotifyRefreshInterval          = 30 * time.Minute
	spotifyTrackSyncInterval        = 10 * time.Minute
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
	db, pool, err := setup.Database(ctx)
	if err != nil {
		return fmt.Errorf("setting up database: %w", err)
	}

	err = setup.Admin(ctx, db, logger)
	if err != nil {
		return fmt.Errorf("setting up admin: %w", err)
	}

	serviceEmail, servicePassword, err := setup.ServiceAccount(ctx, db, logger)
	if err != nil {
		return fmt.Errorf("setting up service account: %w", err)
	}

	e := env.New()
	e.Logger = logger
	e.Database = db
	e.Pool = pool
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

	// Start Spotify token refresh goroutine using service account
	go runSpotifyTokenRefresh(ctx, logger, *e.HTTP, serviceEmail, servicePassword)

	// Start Spotify track sync goroutine using service account
	go runSpotifyTrackSync(ctx, logger, *e.HTTP, serviceEmail, servicePassword)

	// Start weekly playlist goroutine using service account
	go runCreateWeeklyPlaylist(ctx, logger, *e.HTTP, serviceEmail, servicePassword)

	// Start monthly playlist gorouting using service account
	go runCreateMonthlyPlaylist(ctx, logger, *e.HTTP, serviceEmail, servicePassword)

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
			logger.Info("refreshing spotify tokens")
			if err := mars.RefreshSpotifyTokens(ctx, client, email, password); err != nil {
				logger.Error("failed to refresh spotify tokens", "error", err)
			} else {
				logger.Info("refreshed spotify tokens")
			}
		}
	}
}

// runSpotifyTrackSync waits a specified interval before syncing spotify tracks for all users.
func runSpotifyTrackSync(ctx context.Context, logger *slog.Logger, client marshttp.Client, email, password string) {
	ticker := time.NewTicker(spotifyTrackSyncInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("stopping spotify track sync goroutine")
			return
		case <-ticker.C:
			logger.Info("syncing spotify tracks")
			if err := mars.SyncSpotifyTracks(ctx, client, email, password); err != nil {
				logger.Error("failed to sync spotify tracks", "error", err)
			} else {
				logger.Info("synced spotify tracks")
			}
		}
	}
}

// runCreateWeeklyPlaylist creates weekly playlists for all users every Friday at 5 PM America/New_York time.
func runCreateWeeklyPlaylist(ctx context.Context, logger *slog.Logger, client marshttp.Client, email, password string) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatalf("failed to load timezone: %v", err)
	}

	daysUntilTarget := func(now time.Time) time.Time {
		// today at 17:00
		target := time.Date(
			now.Year(), now.Month(), now.Day(),
			17, 0, 0, 0, loc,
		)

		// days til friday
		daysUntil := ((int(time.Sunday) - int(now.Weekday())) + 7) % 7

		// friday, but we're past target
		if daysUntil == 0 && now.After(target) {
			daysUntil = 7
		}

		return target.AddDate(0, 0, daysUntil)
	}

	for {
		now := time.Now().In(loc)
		next := daysUntilTarget(now)
		timer := time.NewTimer(time.Until(next))

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			now = time.Now().In(loc)
			logger.Info("creating playlists", slog.Time("date", now))
			err = mars.CreatePlaylist(ctx, client, email, password, "weekly", now.Year(), now.Month(), now.Day())
			if err != nil {
				logger.Error("failed to create weekly playlist", slog.Any("error", err))
			} else {
				logger.Info("created monthly playlists")
			}
		}
	}
}

// runCreateMonthlyPlaylist creates monthly playlists for all users every first of the month at 5 PM.
func runCreateMonthlyPlaylist(
	ctx context.Context, logger *slog.Logger, client marshttp.Client, email, password string,
) {
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		log.Fatalf("failed to load timezone: %v", err)
	}

	daysUntilTarget := func(now time.Time) time.Time {
		// today at 17:00
		target := time.Date(
			now.Year(), now.Month(), now.Day(),
			17, 0, 0, 0, loc,
		)

		// past target, lets wait for next month
		if now.After(target) {
			return target.AddDate(0, 1, 0)
		}

		return target
	}

	for {
		now := time.Now().In(loc)
		next := daysUntilTarget(now)
		timer := time.NewTimer(time.Until(next))

		select {
		case <-ctx.Done():
			timer.Stop()
			return
		case <-timer.C:
			now = time.Now().In(loc)
			logger.Info("creating playlists", slog.Time("date", now))
			err = mars.CreatePlaylist(ctx, client, email, password, "monthly", now.Year(), now.Month(), now.Day())
			if err != nil {
				logger.Error("failed to create monthly playlist", slog.Any("error", err))
			} else {
				logger.Info("created monthly playlists")
			}
		}
	}
}
