package main

import (
	"log"

	"mangahub/internal/auth"
	"mangahub/internal/manga"
	"mangahub/pkg/database"

	"github.com/gin-gonic/gin"
)

func main() {
	db, err := database.Connect()
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	if err := database.Migrate(db); err != nil {
		log.Fatal(err)
	}

	r := gin.Default()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	mangaHandler := manga.NewHandler(db)
	mangaHandler.RegisterRoutes(r)

	authHandler := auth.NewHandler(db)
	authHandler.RegisterRoutes(r)

	log.Println("Server running at http://localhost:8080")
	r.Run(":8080")
}
