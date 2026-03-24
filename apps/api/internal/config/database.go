package config

import "os"

// Database holds persistence configuration for the API service.
type Database struct {
	Driver string
	DSN    string
}

// Enabled reports whether a database driver and DSN are configured.
func (d Database) Enabled() bool {
	return d.Driver != "" && d.DSN != ""
}

func loadDatabase() Database {
	return Database{
		Driver: os.Getenv("DB_DRIVER"),
		DSN:    os.Getenv("DATABASE_URL"),
	}
}

