package main

import (
	"log"

	"github.com/zealot/managing-up/apps/api/internal/config"
	"github.com/zealot/managing-up/apps/api/internal/repository/postgres"
)

func main() {
	cfg := config.Load()
	if !cfg.Database.Enabled() {
		log.Fatal("database config is required: set DB_DRIVER and DATABASE_URL")
	}

	if err := postgres.Seed(cfg.Database.DSN); err != nil {
		log.Fatalf("seed failed: %v", err)
	}

	log.Print("seed data applied")
}
