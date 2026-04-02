package config

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

type Config struct {
	Port     string
	Database Database
}

func Load() Config {
	loadEnvFiles()

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		Port:     port,
		Database: loadDatabase(),
	}
}

func loadEnvFiles() {
	candidates := []string{
		".env",
		filepath.Join("apps", "api", ".env"),
	}

	for _, path := range candidates {
		if _, err := os.Stat(path); err == nil {
			if err := godotenv.Load(path); err != nil {
				log.Printf("failed to load %s: %v", path, err)
			} else {
				log.Printf("loaded environment from %s", path)
			}
			return
		}
	}
}

func (c Config) Address() string {
	return fmt.Sprintf(":%s", c.Port)
}
