package udp

import (
	"encoding/json"
	"log"
	"net"
)

type Notification struct {
	Type    string `json:"type"`
	Message string `json:"message"`
}

type Server struct {
	Addr string
	conn *net.UDPConn
}

func NewServer(addr string) *Server {
	return &Server{Addr: addr}
}

func (s *Server) Start() {
	udpAddr, err := net.ResolveUDPAddr("udp", s.Addr)
	if err != nil {
		log.Fatal("UDP resolve error:", err)
	}

	conn, err := net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Fatal("UDP listen error:", err)
	}

	s.conn = conn

	log.Println("UDP Server running at", s.Addr)
}

func (s *Server) SendNotification(msg Notification, clientAddr string) {
	addr, err := net.ResolveUDPAddr("udp", clientAddr)
	if err != nil {
		log.Println("UDP resolve client error:", err)
		return
	}

	data, _ := json.Marshal(msg)

	_, err = s.conn.WriteToUDP(data, addr)
	if err != nil {
		log.Println("UDP send error:", err)
	}
}
