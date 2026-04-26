package websocket

import (
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
)

type ChatMessage struct {
	Username  string `json:"username"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type Server struct {
	clients    map[*websocket.Conn]bool
	broadcast  chan ChatMessage
	register   chan *websocket.Conn
	unregister chan *websocket.Conn
}

func NewServer() *Server {
	return &Server{
		clients:    make(map[*websocket.Conn]bool),
		broadcast:  make(chan ChatMessage),
		register:   make(chan *websocket.Conn),
		unregister: make(chan *websocket.Conn),
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
		case conn := <-s.register:
			s.clients[conn] = true
			log.Println("WebSocket client connected")

		case conn := <-s.unregister:
			if _, ok := s.clients[conn]; ok {
				delete(s.clients, conn)
				conn.Close()
				log.Println("WebSocket client disconnected")
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

	s.register <- conn

	defer func() {
		s.unregister <- conn
	}()

	for {
		var msg ChatMessage

		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("WebSocket read error:", err)
			break
		}

		msg.Timestamp = time.Now().Unix()
		s.broadcast <- msg
	}
}
