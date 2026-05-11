package main

import (
	"bufio"
	"fmt"
	"net"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:9090")
	if err != nil {
		panic(err)
	}
	defer conn.Close()

	token := "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3Nzg2MjI2OTAsInVzZXJfaWQiOiJkMWM2M2Q3OS03MjQyLTRiOTAtYmZkYi0xOTdiZWY4NDI5ZGMifQ.-Xgv_qMkfDIl8A0GHL0qbH1kayVb9d4RDX-wJBJw2fE"

	authMsg := fmt.Sprintf(`{"type":"auth","token":"%s"}`+"\n", token)
	conn.Write([]byte(authMsg))

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Println("FROM TCP SERVER:", scanner.Text())
	}
}
