package main

import (
	"fmt"
	"log/slog"
	"net"
)

func handleConnection(conn net.Conn) {
	defer conn.Close()

	conn.Write([]byte("Server received your connection!\n"))
	fmt.Println("Handled connection")
}

func startServer(port string) error {
	listener, err := net.Listen("tcp", ":"+port)
	if err != nil {
		slog.Error(fmt.Sprintf("Error while creating the tcp listener on port %s: %s", port, err))
		return err
	}

	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			slog.Error(fmt.Sprintf("Error while accepting traffic: %s", err))
			return err
		}
		go handleConnection(conn)
	}
}

func main() {
	port := "8002"
	fmt.Printf("Starting server on port %s\n", port)
	startServer(port)
}
