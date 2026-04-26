package tcp

import (
	"encoding/json"
	"fmt"
	"log"
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

type Server struct {
	Address     string
	clients     map[net.Conn]bool
	mu          sync.Mutex
	BroadcastCh chan ProgressUpdate
}

func NewServer(address string) *Server {
	return &Server{
		Address:     address,
		clients:     make(map[net.Conn]bool),
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

		s.mu.Lock()
		s.clients[conn] = true
		s.mu.Unlock()

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

	buffer := make([]byte, 1024)

	for {
		_, err := conn.Read(buffer)
		if err != nil {
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
		for conn := range s.clients {
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
