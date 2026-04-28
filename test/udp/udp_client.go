package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
)

type ClientMessage struct {
	Type string `json:"type"`
}

func main() {
	serverAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1:7070")
	if err != nil {
		log.Fatal(err)
	}

	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	msg := ClientMessage{
		Type: "register",
	}

	data, _ := json.Marshal(msg)

	_, err = conn.Write(data)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Registered to UDP server")
	fmt.Println("Waiting for notifications...")

	buffer := make([]byte, 2048)

	for {
		n, _, err := conn.ReadFromUDP(buffer)
		if err != nil {
			log.Println("Read error:", err)
			continue
		}

		fmt.Println("Received:", string(buffer[:n]))
	}
}
