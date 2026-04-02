package main

import (
	"database/sql"
	"fmt"
	"log"
	"os"

	_ "github.com/lib/pq"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "postgres://postgres:pass@172.20.0.16:5432/managing-up"
	}

	db, err := sql.Open("postgres", dsn)
	if err != nil {
		log.Fatalf("Failed to connect: %v", err)
	}
	defer db.Close()

	hash := "$2a$10$KP0KaTa.nYASJAExYb9KCutPCDnR.b2R3cYTOR8BrzeS9hkLXpfPW"

	_, err = db.Exec(`UPDATE users SET password_hash = $1 WHERE username = 'admin'`, hash)
	if err != nil {
		log.Fatalf("Failed to update: %v", err)
	}

	fmt.Println("Updated admin password hash")
}
