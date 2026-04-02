package main

import (
	"fmt"
	"log"

	"github.com/zealot/managing-up/apps/api/internal/service"
)

func main() {
	hash, err := service.HashPassword("admin")
	if err != nil {
		log.Fatalf("Failed to hash: %v", err)
	}
	fmt.Println("Hash for 'admin':", hash)
}
