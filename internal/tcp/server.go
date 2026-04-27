package tcp

import (
	"encoding/json"
	"fmt"
	"log"
	"mangahub/internal/auth"
	"net"
	"sync"
	"time"
)

type ProgressUpdate struct {
	UserID    string `json:"user_id"`
	MangaID   string `json:"manga_id"`
	Chapter   int    `json:"chapter"`
	Timestamp int64  `json:"timestamp"`
}

type AuthMessage struct {
	Type  string `json:"type"`
	Token string `json:"token"`
}

type Server struct {
	Address     string
	clients     map[net.Conn]string
	mu          sync.Mutex
	BroadcastCh chan ProgressUpdate
}

func NewServer(address string) *Server {
	return &Server{
		Address:     address,
		clients:     make(map[net.Conn]string),
		BroadcastCh: make(chan ProgressUpdate),
	}
}

func (s *Server) Start() {
	listener, err := net.Listen("tcp", s.Address)
	if err != nil {
		log.Fatal("TCP listen error:", err)
	}
	defer listener.Close()

	log.Println("TCP Sync Server running at", s.Address)

	go s.handleBroadcast()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("TCP accept error:", err)
			continue
		}

		log.Println("TCP client connected:", conn.RemoteAddr())

		go s.handleClient(conn)
	}
}

func (s *Server) handleClient(conn net.Conn) {
	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()

		conn.Close()
		log.Println("TCP client disconnected:", conn.RemoteAddr())
	}()

	decoder := json.NewDecoder(conn)

	var authMsg AuthMessage
	if err := decoder.Decode(&authMsg); err != nil {
		conn.Write([]byte(`{"type":"error","message":"invalid auth message"}` + "\n"))
		return
	}

	if authMsg.Type != "auth" || authMsg.Token == "" {
		conn.Write([]byte(`{"type":"error","message":"authentication required"}` + "\n"))
		return
	}

	userID, err := auth.ValidateToken(authMsg.Token)
	if err != nil {
		conn.Write([]byte(`{"type":"error","message":"invalid token"}` + "\n"))
		return
	}

	s.mu.Lock()
	s.clients[conn] = userID
	s.mu.Unlock()

	conn.Write([]byte(`{"type":"auth_success","message":"TCP authentication successful"}` + "\n"))

	log.Println("TCP client authenticated:", conn.RemoteAddr(), "user:", userID)

	for {
		var msg map[string]interface{}
		if err := decoder.Decode(&msg); err != nil {
			return
		}
	}
}

func (s *Server) handleBroadcast() {
	for update := range s.BroadcastCh {
		data, err := json.Marshal(update)
		if err != nil {
			log.Println("TCP marshal error:", err)
			continue
		}

		message := fmt.Sprintf("%s\n", data)

		s.mu.Lock()
		for conn, userID := range s.clients {
			if userID != update.UserID {
				continue
			}

			_, err := conn.Write([]byte(message))
			if err != nil {
				log.Println("TCP write error:", err)
				conn.Close()
				delete(s.clients, conn)
			}
		}
		s.mu.Unlock()
	}
}

func NewProgressUpdate(userID, mangaID string, chapter int) ProgressUpdate {
	return ProgressUpdate{
		UserID:    userID,
		MangaID:   mangaID,
		Chapter:   chapter,
		Timestamp: time.Now().Unix(),
	}
}
