package main

import (
	"context"
	"log"
	"log/slog"
	"os"
	"os/signal"
	"strconv"
	"syscall"

	"mars/internal/admin"
	"mars/internal/api"
	"mars/internal/database"
	"mars/internal/env"
	marslog "mars/internal/log"

	"github.com/jackc/pgx/v5/pgxpool"
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
	databaseHost := os.Getenv("DATABASE_HOST")
	if databaseHost == "" {
		log.Fatal("DATABASE_HOST environment variable is required")
	}
	databasePort := os.Getenv("DATABASE_PORT")
	if databasePort == "" {
		log.Fatal("DATABASE_PORT environment variable is required")
	}
	databaseUser := os.Getenv("DATABASE_USER")
	if databaseUser == "" {
		log.Fatal("DATABASE_USER environment variable is required")
	}
	databasePassword := os.Getenv("DATABASE_PASSWORD")
	if databasePassword == "" {
		log.Fatal("DATABASE_PASSWORD environment variable is required")
	}
	databaseName := os.Getenv("DATABASE_NAME")
	if databaseName == "" {
		log.Fatal("DATABASE_NAME environment variable is required")
	}

	poolConfig, err := pgxpool.ParseConfig("")
	if err != nil {
		log.Fatalf("failed to create database config: %v", err)
	}

	poolConfig.ConnConfig.Host = databaseHost
	poolConfig.ConnConfig.Port = func() uint16 {
		p, err := strconv.ParseUint(databasePort, 10, 16)
		if err != nil {
			log.Fatalf("failed to parse database port: %v", err)
		}
		return uint16(p)
	}()
	poolConfig.ConnConfig.User = databaseUser
	poolConfig.ConnConfig.Password = databasePassword
	poolConfig.ConnConfig.Database = databaseName

	// Creating DB connection
	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Fatalf("failed to create database connection: %v", err)
	}
	defer pool.Close()

	db := database.New(pool)
	if err := db.Ping(ctx); err != nil {
		log.Fatalf("failed to ping database: %v", err)
	}
	if err := database.ApplySchema(ctx, pool); err != nil {
		log.Fatalf("failed to apply database schema: %v", err)
	}

	adminEmail := os.Getenv("ADMIN_EMAIL")
	adminPassword := os.Getenv("ADMIN_PASSWORD")

	if err := admin.SeedAdmin(ctx, db, logger, adminEmail, adminPassword); err != nil {
		log.Fatalf("seeding admin user: %v", err)
	}

	port := defaultPort
	if portStr := os.Getenv("PORT"); portStr != "" {
		p, err := strconv.ParseUint(portStr, 10, 16)
		if err != nil {
			log.Fatalf("invalid PORT value: %v", err)
		}
		port = uint16(p)
	}

	e := &env.Env{
		Logger:   logger,
		Database: db,
	}

	return api.Start(ctx, port, e)
}
