package main

import (
	"log"

	"mangahub/internal/auth"
	"mangahub/internal/manga"
	"mangahub/internal/tcp"
	"mangahub/internal/user"
	ws "mangahub/internal/websocket"
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

	tcpServer := tcp.NewServer(":9090")
	go tcpServer.Start()

	wsServer := ws.NewServer()
	wsServer.RegisterRoutes(r)
	go wsServer.Run()

	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "pong"})
	})

	mangaHandler := manga.NewHandler(db)
	mangaHandler.RegisterRoutes(r)

	authHandler := auth.NewHandler(db)
	authHandler.RegisterRoutes(r)

	userHandler := user.NewHandler(db, tcpServer)
	userHandler.RegisterRoutes(r)

	r.GET("/me", auth.AuthMiddleware(), func(c *gin.Context) {
		userID := c.GetString("user_id")
		c.JSON(200, gin.H{
			"user_id": userID,
		})
	})

	r.GET("/seed", func(c *gin.Context) {
		_, err := db.Exec(`
	INSERT OR IGNORE INTO manga (id, title, author, genres, status, total_chapters, description)
	VALUES 
	('one-piece', 'One Piece', 'Oda Eiichiro', 'Action,Adventure,Shounen', 'ongoing', 1100, 'Pirate adventure'),
	('naruto', 'Naruto', 'Masashi Kishimoto', 'Action,Shounen', 'completed', 700, 'Ninja story'),
	('death-note', 'Death Note', 'Tsugumi Ohba', 'Mystery,Psychological', 'completed', 108, 'Notebook kills people');
	`)
		if err != nil {
			c.JSON(500, gin.H{"error": err.Error()})
			return
		}

		c.JSON(200, gin.H{"message": "seed data inserted"})
	})

	log.Println("Server running at http://localhost:8080")
	r.Run(":8080")
}
