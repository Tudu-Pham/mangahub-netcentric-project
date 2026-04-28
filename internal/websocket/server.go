package websocket

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ChatMessage struct {
	Type      string `json:"type"`
	UserID    string `json:"user_id"`
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type Client struct {
	Conn     *websocket.Conn
	UserID   string
	Username string
}

type Server struct {
	clients    map[*websocket.Conn]*Client
	broadcast  chan ChatMessage
	register   chan *Client
	unregister chan *Client
}

func NewServer() *Server {
	return &Server{
		clients:    make(map[*websocket.Conn]*Client),
		broadcast:  make(chan ChatMessage, 100),
		register:   make(chan *Client),
		unregister: make(chan *Client),
	}
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

func (s *Server) RegisterRoutes(r *gin.Engine) {
	r.GET("/ws/chat", s.HandleConnection)
}

func (s *Server) Run() {
	for {
		select {
		case client := <-s.register:
			s.clients[client.Conn] = client
			log.Println("WebSocket client connected:", client.Username)

			s.broadcast <- ChatMessage{
				Type:      "join",
				UserID:    client.UserID,
				Username:  client.Username,
				Message:   client.Username + " joined the chat",
				Timestamp: time.Now().Unix(),
			}

		case client := <-s.unregister:
			if _, ok := s.clients[client.Conn]; ok {
				delete(s.clients, client.Conn)
				client.Conn.Close()
				log.Println("WebSocket client disconnected:", client.Username)

				s.broadcast <- ChatMessage{
					Type:      "left",
					UserID:    client.UserID,
					Username:  client.Username,
					Message:   client.Username + " left the chat",
					Timestamp: time.Now().Unix(),
				}
			}

		case msg := <-s.broadcast:
			for conn := range s.clients {
				err := conn.WriteJSON(msg)
				if err != nil {
					log.Println("WebSocket write error:", err)
					conn.Close()
					delete(s.clients, conn)
				}
			}
		}
	}
}

func (s *Server) HandleConnection(c *gin.Context) {
	conn, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		log.Println("WebSocket upgrade error:", err)
		return
	}

	userID := c.Query("user_id")
	username := c.Query("username")

	if userID == "" {
		userID = "guest"
	}

	if username == "" {
		username = "Anonymous"
	}

	client := &Client{
		Conn:     conn,
		UserID:   userID,
		Username: username,
	}

	s.register <- client

	defer func() {
		s.unregister <- client
	}()

	for {
		var input struct {
			Message string `json:"message"`
		}

		err := conn.ReadJSON(&input)
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		if input.Message == "" {
			continue
		}

		s.broadcast <- ChatMessage{
			Type:      "chat",
			UserID:    client.UserID,
			Username:  client.Username,
			Message:   input.Message,
			Timestamp: time.Now().Unix(),
		}
	}
}
