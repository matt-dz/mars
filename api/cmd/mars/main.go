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

	"mars/internal/api"
	"mars/internal/env"
	marslog "mars/internal/log"
	"mars/internal/setup"
)

const defaultPort uint16 = 8080

func main() {
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()

	logger := marslog.New(nil)

	if err := run(ctx, logger); err != nil {
		log.Fatal(err)
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

	return api.Start(ctx, port, e)
}
