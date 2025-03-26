package main

import (
	"fmt"
	"log"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	upgrader  = websocket.Upgrader{}
	clients   = make(map[*websocket.Conn]bool)
	clientsMu sync.Mutex
)

func main() {
	e := echo.New()
	e.GET("/ws", handleWebSocket)

	log.Println("Server is listening on :8000")
	e.Logger.Fatal(e.Start(":8000"))
}

func handleWebSocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	clientIP := ws.RemoteAddr().String()
	log.Printf("New client connected: %s", clientIP)

	clientsMu.Lock()
	clients[ws] = true
	clientsMu.Unlock()

	defer func() {
		clientsMu.Lock()
		delete(clients, ws)
		clientsMu.Unlock()
		log.Printf("Client disconnected: %s", clientIP)
	}()

	for {
		messageType, p, err := ws.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("Client %s closed connection normally", clientIP)
			} else if websocket.IsUnexpectedCloseError(err, websocket.CloseAbnormalClosure) {
				log.Printf("Unexpected close error from %s: %v", clientIP, err)
			} else {
				log.Printf("Error reading message from %s: %v", clientIP, err)
			}
			return nil
		}

		log.Printf("Message from client %s: %s", clientIP, string(p))
		response := fmt.Sprintf("Hello dear client with %s IP!; your message was: %s", clientIP, p)

		if err := ws.WriteMessage(messageType, []byte(response)); err != nil {
			log.Printf("Error writing message to %s: %v", clientIP, err)
			return nil
		}
	}
}
