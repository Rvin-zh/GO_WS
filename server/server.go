package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
)

func main() {
	fmt.Println("Starting server...")
	listener, err := net.Listen("tcp", ":8080")
	if err != nil {
		fmt.Println("Error starting server:", err)
		os.Exit(1)
	}
	defer listener.Close()

	fmt.Println("Server is listening on port 8080")

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}
		go handleConnection(conn)
	}
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	fmt.Println("New client connected:", conn.RemoteAddr())

	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Client disconnected:", conn.RemoteAddr())
			return
		}

		fmt.Print("Message from client:", message)
		conn.Write([]byte("Hello Client\n"))
	}
}
