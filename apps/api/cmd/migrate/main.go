package main

import (
	"log"
	"os"
	"path/filepath"

	"github.com/zealot/managing-up/apps/api/internal/config"
	"github.com/zealot/managing-up/apps/api/internal/repository/postgres"
)

func main() {
	cfg := config.Load()
	if !cfg.Database.Enabled() {
		log.Fatal("database config is required: set DB_DRIVER and DATABASE_URL")
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("resolve working directory: %v", err)
	}

	moduleRoot := wd
	if filepath.Base(wd) == "cmd" {
		moduleRoot = filepath.Dir(filepath.Dir(wd))
	}

	migrationsDir := postgres.MigrationsDir(moduleRoot)
	if err := postgres.Migrate(cfg.Database.DSN, migrationsDir); err != nil {
		log.Fatalf("migration failed: %v", err)
	}

	log.Printf("migrations applied from %s", migrationsDir)
}
