package main

import (
	"bufio"
	"log"
	"net/url"
	"os"
	"os/signal"
	"time"

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

	done := make(chan struct{})

	go receiveMessages(c, done)
	go sendMessages(c, done)

	for {
		select {
		case <-done:
			return
		case <-interrupt:
			log.Println("Interrupt received, closing connection...")
			err := c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
			if err != nil {
				log.Println("write close:", err)
			}
			select {
			case <-done:
			case <-time.After(time.Millisecond * 500):
			}
			return
		}
	}
}

func receiveMessages(c *websocket.Conn, done chan struct{}) {
	defer close(done)
	for {
		_, message, err := c.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Println("Connection closed normally")
			} else {
				handleCloseError(err)
			}
			return
		}
		log.Printf("Received: %s", message)
	}
}

func sendMessages(c *websocket.Conn, done chan struct{}) {
	scanner := bufio.NewScanner(os.Stdin)
	for {
		select {
		case <-done:
			return
		default:
			if scanner.Scan() {
				message := scanner.Text()
				err := c.WriteMessage(websocket.TextMessage, []byte(message))
				if err != nil {
					log.Println("Error while sending message:")
					handleCloseError(err)
					return
				}
			} else if err := scanner.Err(); err != nil {
				log.Printf("Error reading input: %v", err)
				return
			}
		}
	}
}

func handleCloseError(err error) {
	if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
		log.Println("Connection closed normally")
	} else if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
		log.Printf("Unexpected close error: %v", err)
	} else {
		log.Printf("WebSocket error: %v", err)
	}
}
