package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"net/url"
	"os"
	"os/signal"
	"strings"
	"sync"

	"github.com/gorilla/websocket"
)

func main() {
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	u := url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/ws"}
	log.Printf("Connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("dial:", err)
	}
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	// Channel to send messages to the server
	sendCh := make(chan string)

	// Read messages from server
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			default:
				_, message, err := c.ReadMessage()
				if err != nil {
					log.Println("read:", err)
					cancel()
					return
				}
				fmt.Printf("\nReceived from server: %s\n", message)
				fmt.Print("Enter message: ")
			}
		}
	}()

	// Send messages to server
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return
			case message := <-sendCh:
				err := c.WriteMessage(websocket.TextMessage, []byte(message))
				if err != nil {
					log.Println("write:", err)
					cancel()
					return
				}
			}
		}
	}()

	// Read user input
	go func() {
		reader := bufio.NewReader(os.Stdin)
		for {
			fmt.Print("Enter message: ")
			message, _ := reader.ReadString('\n')
			message = strings.TrimSpace(message)
			if message != "" {
				sendCh <- message
			}
			if ctx.Err() != nil {
				return
			}
		}
	}()

	select {
	case <-interrupt:
		log.Println("Interrupt received, closing connection...")
		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("write close:", err)
		}
		cancel()
	case <-ctx.Done():
	}

	wg.Wait()
}
