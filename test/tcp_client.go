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

	token := "DAN_TOKEN_LOGIN_VAO_DAY"

	authMsg := fmt.Sprintf(`{"type":"auth","token":"%s"}`+"\n", token)
	conn.Write([]byte(authMsg))

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		fmt.Println("FROM TCP SERVER:", scanner.Text())
	}
}
