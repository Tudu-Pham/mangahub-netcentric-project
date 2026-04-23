package main

import (
	"log"

	"mangahub/pkg/database"
)

func main() {
	log.Println("Starting MangaHub server...")

	database.Connect()
	database.Migrate()

	log.Println("Project setup completed successfully")
}