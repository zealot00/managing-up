package config

import (
	"fmt"
	"os"
)

// Config holds application configuration loaded from the environment.
type Config struct {
	Port     string
	Database Database
}

// Load reads application configuration from environment variables.
func Load() Config {
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	return Config{
		Port:     port,
		Database: loadDatabase(),
	}
}

// Address returns the listen address for the HTTP server.
func (c Config) Address() string {
	return fmt.Sprintf(":%s", c.Port)
}
