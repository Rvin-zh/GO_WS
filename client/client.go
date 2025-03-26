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
	"time"

	"github.com/gorilla/websocket"
)

func main() {
	u := url.URL{Scheme: "ws", Host: "localhost:8000", Path: "/ws"}
	log.Printf("Connecting to %s", u.String())

	c, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal("Error connecting to WebSocket server:", err)
	}
	defer c.Close()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	var wg sync.WaitGroup
	wg.Add(2)

	// Read messages from server
	go func() {
		defer wg.Done()
		readMessages(ctx, c)
	}()

	// Send messages to server
	go func() {
		defer wg.Done()
		sendMessages(ctx, c)
	}()

	// Handle interrupt signal
	interrupt := make(chan os.Signal, 1)
	signal.Notify(interrupt, os.Interrupt)

	select {
	case <-interrupt:
		log.Println("Interrupt received, closing connection...")
		err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
		if err != nil {
			log.Println("Error during closing WebSocket:", err)
		}
		cancel()
	case <-ctx.Done():
	}

	wg.Wait()
}

func readMessages(ctx context.Context, c *websocket.Conn) {
	for {
		select {
		case <-ctx.Done():
			return
		default:
			_, message, err := c.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
					log.Println("Server closed connection normally")
				} else {
					log.Printf("Error reading message: %v", err)
				}
				return
			}
			log.Printf("Received from server: %s", message)
		}
	}
}

func sendMessages(ctx context.Context, c *websocket.Conn) {
	reader := bufio.NewReader(os.Stdin)

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			fmt.Print("Enter message: ")
			message, err := reader.ReadString('\n')
			if err != nil {
				log.Println("Error reading input:", err)
				continue
			}
			message = strings.TrimSpace(message)

			err = c.WriteMessage(websocket.TextMessage, []byte(message))
			if err != nil {
				log.Println("Error writing message:", err)
				return
			}
		}
	}
}
