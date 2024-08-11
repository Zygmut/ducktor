package main

import (
	"fmt"
	"log"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("Server received your connection!\n"))
	fmt.Println("Handled connection")
}

func startServer(port string) {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go handleConnection(conn)
	}
}

func main() {
	port := "8002"
	fmt.Printf("Starting server on port %s\n", port)
	startServer(port)
}
