package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"learning/internal/infrastructure/config"
	"learning/internal/infrastructure/csvloader"
	"learning/internal/infrastructure/db"
)

func main() {
	if err := run(); err != nil {
		slog.Error("seed failed", "error", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	ctx := context.Background()
	repository, err := db.Open(ctx, cfg.DatabaseURL)
	if err != nil {
		return err
	}
	defer repository.Close()
	if err := repository.Migrate(ctx, cfg.MigrationsSQL); err != nil {
		return err
	}

	dataset, err := csvloader.Load("../data/readings.csv", "../data/events.csv")
	if err != nil {
		return err
	}
	if err := repository.Seed(ctx, dataset); err != nil {
		return err
	}
	fmt.Printf("Loaded %d readings and %d events.\n", len(dataset.Readings), len(dataset.Events))
	return nil
}
