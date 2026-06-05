package main

import (
	"context"
	"encoding/json"
	"log/slog"
	"os"
	"time"

	"luminous/internal/config"
	"luminous/internal/model"
	"luminous/internal/repository"
)

func main() {
	if err := config.LoadConfig(); err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	schoolsFile := "./data/schools.json"
	if len(os.Args) > 1 {
		schoolsFile = os.Args[1]
	}

	data, err := os.ReadFile(schoolsFile)
	if err != nil {
		slog.Error("Failed to read schools file", "path", schoolsFile, "error", err)
		os.Exit(1)
	}

	var schools map[string]*model.School
	if err := json.Unmarshal(data, &schools); err != nil {
		slog.Error("Failed to parse schools file", "error", err)
		os.Exit(1)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pgRepo, err := repository.NewPGSchoolRepository(ctx, config.Cfg.Database)
	if err != nil {
		slog.Error("Failed to connect to PostgreSQL", "error", err)
		os.Exit(1)
	}
	defer pgRepo.Close()

	count, failed := 0, 0
	for _, s := range schools {
		if err := pgRepo.Create(ctx, s); err != nil {
			slog.Warn("Migration skipped school", "code", s.Code, "error", err)
			failed++
			continue
		}
		count++
	}

	slog.Info("Migration complete", "migrated", count, "failed", failed)
}
