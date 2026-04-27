package main

import (
	"log"

	"mangahub/internal/auth"
	"mangahub/internal/grpc"
	"mangahub/internal/manga"
	"mangahub/internal/tcp"
	"mangahub/internal/udp"
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

	udpServer := udp.NewServer(":7070")
	udpServer.Start()

	userHandler := user.NewHandler(db, tcpServer, udpServer)
	userHandler.RegisterRoutes(r)

	grpcServer := grpc.NewServer(db)
	go grpcServer.Start()

	r.GET("/me", auth.AuthMiddleware(), func(c *gin.Context) {
		userID := c.GetString("user_id")
		c.JSON(200, gin.H{
			"user_id": userID,
		})
	})

	log.Println("Server running at http://localhost:8080")
	r.Run(":8080")
}
