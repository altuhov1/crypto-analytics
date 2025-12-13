package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/joho/godotenv"
	"github.com/pressly/goose/v3"

	_ "crypto-analytics/migrations"
)

var (
	dir    = flag.String("dir", "migrations", "directory with migration files")
	logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
		Level: slog.LevelInfo,
	}))
)

func main() {
	slog.SetDefault(logger)
	flag.Parse()

	if _, err := os.Stat(".env"); err == nil {
		if err := godotenv.Load(); err != nil {
			logger.Error("Failed to load .env", "error", err)
			os.Exit(1)
		}
	}

	dsn := fmt.Sprintf(
		"host=%s port=5432 user=%s password=%s dbname=%s sslmode=disable",
		getEnv("DB_PG_HOST", "localhost"),
		getEnv("USER_DB_PG", "user"),
		getEnv("PASS_DB_PG", "1111"),
		getEnv("NAME_DB_PG", "pullrequestdb"),
	)

	pool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		logger.Error("Failed to create pgx pool", "error", err)
		os.Exit(1)
	}
	defer pool.Close()

	db := stdlib.OpenDBFromPool(pool)
	defer db.Close()

	args := flag.Args()
	if len(args) == 0 {
		logger.Error("Usage: migrate [up|down|reset|status|version]")
		os.Exit(1)
	}

	command := args[0]
	switch command {
	case "up":
		if err := goose.Up(db, *dir); err != nil {
			logger.Error("goose up failed", "error", err)
			os.Exit(1)
		}
		logger.Info("Migrations applied")

	case "down":
		if len(args) > 1 {
			version, err := parseVersion(args[1])
			if err != nil {
				logger.Error("Invalid version number", "version", args[1], "error", err)
				os.Exit(1)
			}

			if err := goose.DownTo(db, *dir, version); err != nil {
				logger.Error("goose down to version failed",
					"version", version,
					"error", err)
				os.Exit(1)
			}
			logger.Info("Migration rolled back to version", "version", version)
		} else {
			if err := goose.Down(db, *dir); err != nil {
				logger.Error("goose down failed", "error", err)
				os.Exit(1)
			}
			logger.Info("Migration rolled back")
		}

	case "reset":
		if err := goose.Reset(db, *dir); err != nil {
			logger.Error("goose reset failed", "error", err)
			os.Exit(1)
		}
		logger.Info("All migrations reset")

	case "status":
		if err := goose.Status(db, *dir); err != nil {
			logger.Error("goose status failed", "error", err)
			os.Exit(1)
		}

	case "version":
		version, err := goose.GetDBVersion(db)
		if err != nil {
			logger.Error("Failed to get DB version", "error", err)
			os.Exit(1)
		}
		logger.Info("Current DB version", "version", version)

	default:
		logger.Error("Unknown command",
			"command", command,
			"available", []string{"up", "down", "reset", "status", "version"})
		os.Exit(1)
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func parseVersion(s string) (int64, error) {
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0, fmt.Errorf("invalid version number: %w", err)
	}
	return n, nil
}
