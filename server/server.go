package main

import (
	"fmt"
	"log"
	"net/http"
	"sync"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Client struct {
	conn *websocket.Conn
	ip   string
}

type Server struct {
	clients    map[*Client]bool
	clientsMux sync.RWMutex
	upgrader   websocket.Upgrader
}

func NewServer() *Server {
	return &Server{
		clients:  make(map[*Client]bool),
		upgrader: websocket.Upgrader{CheckOrigin: func(r *http.Request) bool { return true }},
	}
}

func (s *Server) addClient(client *Client) {
	s.clientsMux.Lock()
	defer s.clientsMux.Unlock()
	s.clients[client] = true
}

func (s *Server) removeClient(client *Client) {
	s.clientsMux.Lock()
	defer s.clientsMux.Unlock()
	delete(s.clients, client)
}

func (s *Server) handleWebSocket(c echo.Context) error {
	ws, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return fmt.Errorf("websocket upgrade error: %w", err)
	}

	client := &Client{conn: ws, ip: ws.RemoteAddr().String()}
	s.addClient(client)
	log.Printf("New client connected: %s", client.ip)

	defer func() {
		s.removeClient(client)
		ws.Close()
		log.Printf("Client disconnected: %s", client.ip)
	}()

	err = s.handleClientMessages(client)
	if err != nil {
		if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
			log.Printf("error: %v", err)
		}
	}

	return nil
}

func (s *Server) handleClientMessages(client *Client) error {
	for {
		messageType, msg, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("Client %s closed connection normally", client.ip)
			} else {
				log.Printf("Unexpected close error from %s: %v", client.ip, err)
			}
			return nil // Exit the loop for any error
		}

		if len(msg) == 0 {
			log.Printf("Received empty message from %s, ignoring", client.ip)
			continue
		}

		log.Printf("Received from %s: %s", client.ip, msg)

		if err := s.sendResponse(client, messageType, msg); err != nil {
			log.Printf("Error sending response to %s: %v", client.ip, err)
			return err
		}
	}
}

func (s *Server) sendResponse(client *Client, messageType int, msg []byte) error {
	response := fmt.Sprintf("Hi dear client with %s IP; your message was: %s", client.ip, msg)
	return client.conn.WriteMessage(messageType, []byte(response))
}

func main() {
	server := NewServer()
	e := echo.New()
	e.GET("/ws", server.handleWebSocket)

	fmt.Println("Server is running on :8000")
	if err := e.Start(":8000"); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
