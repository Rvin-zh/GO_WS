package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strings"
)

func main() {
	conn, err := net.Dial("tcp", "localhost:8080")
	if err != nil {
		fmt.Println("Error connecting to server:", err)
		os.Exit(1)
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.TCPAddr)
	fmt.Printf("Connected to server from %s\n", localAddr.IP.String())

	go receiveMessages(conn)

	for {
		reader := bufio.NewReader(os.Stdin)
		fmt.Print("Enter message: ")
		message, _ := reader.ReadString('\n')
		message = strings.TrimSpace(message)

		if message == "/exit" {
			fmt.Fprintf(conn, "%s\n", message)
			fmt.Println("Exiting chat...")
			return
		}

		fmt.Fprintf(conn, "%s\n", message)
	}
}

func receiveMessages(conn net.Conn) {
	for {
		message, err := bufio.NewReader(conn).ReadString('\n')
		if err != nil {
			fmt.Println("Lost connection to server.")
			os.Exit(1)
		}
		fmt.Print(message)
	}
}
