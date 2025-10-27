package main

import (
	"crypto-analytics/internal/app"
	"crypto-analytics/internal/config"
	"log/slog"
)

func main() {
	cfg := config.MustLoad()
	logger := config.NewLogger(cfg)
	slog.SetDefault(logger)
	app := app.NewApp(cfg)
	app.Run()
}
