package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

var (
	upgrader = websocket.Upgrader{
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
)

func main() {
	e := echo.New()
	e.GET("/ws", handleWebSocket)

	fmt.Println("Server is running on :8000")
	e.Logger.Fatal(e.Start(":8000"))
}

func handleWebSocket(c echo.Context) error {
	ws, err := upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return err
	}
	defer ws.Close()

	for {
		// Read message from the client
		_, msg, err := ws.ReadMessage()
		if err != nil {
			log.Printf("Error reading message: %v", err)
			break
		}
		log.Printf("Received: %s", msg)

		// Send a response back to the client
		response := fmt.Sprintf("Server received: %s", msg)
		err = ws.WriteMessage(websocket.TextMessage, []byte(response))
		if err != nil {
			log.Printf("Error writing message: %v", err)
			break
		}
	}

	return nil
}
