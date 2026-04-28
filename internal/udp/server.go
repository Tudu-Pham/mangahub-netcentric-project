package udp

import (
	"encoding/json"
	"log"
	"net"
	"sync"
	"time"
)

type Notification struct {
	Type      string `json:"type"`
	MangaID   string `json:"manga_id"`
	Message   string `json:"message"`
	Timestamp int64  `json:"timestamp"`
}

type ClientMessage struct {
	Type string `json:"type"`
}

type Server struct {
	Address string
	conn    *net.UDPConn
	clients map[string]*net.UDPAddr
	mu      sync.Mutex
}

func NewServer(address string) *Server {
	return &Server{
		Address: address,
		clients: make(map[string]*net.UDPAddr),
	}
}

func (s *Server) Start() {
	addr, err := net.ResolveUDPAddr("udp", s.Address)
	if err != nil {
		log.Println("UDP resolve error:", err)
		return
	}

	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		log.Println("UDP listen error:", err)
		return
	}

	s.conn = conn
	log.Println("UDP Notification Server running at", s.Address)

	buffer := make([]byte, 1024)

	for {
		n, clientAddr, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("UDP read error:", err)
			continue
		}

		var msg ClientMessage
		if err := json.Unmarshal(buffer[:n], &msg); err != nil {
			log.Println("Invalid UDP message from", clientAddr.String())
			continue
		}

		switch msg.Type {
		case "register":
			s.RegisterClient(clientAddr)

		case "unregister":
			s.UnregisterClient(clientAddr)

		default:
			log.Println("Unknown UDP message type:", msg.Type)
		}
	}
}

func (s *Server) RegisterClient(addr *net.UDPAddr) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.clients[addr.String()] = addr
	log.Println("UDP client registered:", addr.String())

	s.sendToClient(addr, Notification{
		Type:      "registered",
		MangaID:   "",
		Message:   "Registered for manga notifications",
		Timestamp: time.Now().Unix(),
	})
}

func (s *Server) UnregisterClient(addr *net.UDPAddr) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.clients, addr.String())
	log.Println("UDP client unregistered:", addr.String())
}

func (s *Server) Broadcast(notification Notification) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.conn == nil {
		log.Println("UDP connection is not ready")
		return
	}

	data, err := json.Marshal(notification)
	if err != nil {
		log.Println("UDP marshal error:", err)
		return
	}

	for key, client := range s.clients {
		_, err := s.conn.WriteToUDP(data, client)
		if err != nil {
			log.Println("UDP send error to", key, ":", err)
			delete(s.clients, key)
			continue
		}

		log.Println("UDP notification sent to", key)
	}
}

func (s *Server) sendToClient(addr *net.UDPAddr, notification Notification) {
	if s.conn == nil {
		return
	}

	data, err := json.Marshal(notification)
	if err != nil {
		log.Println("UDP marshal error:", err)
		return
	}

	_, err = s.conn.WriteToUDP(data, addr)
	if err != nil {
		log.Println("UDP send error:", err)
	}
}
