package main

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/labstack/echo/v4"
)

type Client struct {
	conn     *websocket.Conn
	ip       string
	username string
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
	username := client.username
	delete(s.clients, client)

	if username != "" {
		leaveMsg := fmt.Sprintf("[Server] %s has left the chat.", username)
		log.Println(leaveMsg)
		// We need to use a goroutine to broadcast since we're holding the lock
		go s.broadcastMessage(websocket.TextMessage, []byte(leaveMsg), nil)
	}
}

func (s *Server) handleWebSocket(c echo.Context) error {
	ws, err := s.upgrader.Upgrade(c.Response(), c.Request(), nil)
	if err != nil {
		return fmt.Errorf("websocket upgrade error: %w", err)
	}

	client := &Client{conn: ws, ip: ws.RemoteAddr().String(), username: ""}
	s.addClient(client)
	log.Printf("New client connected: %s", client.ip)

	defer func() {
		s.removeClient(client)
		ws.Close()
		log.Printf("Client disconnected: %s (Username: %s)", client.ip, client.username)
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
		messageType, p, err := client.conn.ReadMessage()
		if err != nil {
			if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) {
				log.Printf("Client %s closed connection normally", client.ip)
			} else {
				log.Printf("Unexpected close error from %s: %v", client.ip, err)
			}
			return nil
		}

		message := string(p)
		log.Printf("Received from %s: %s", client.ip, message)

		if len(message) == 0 {
			log.Printf("Received empty message from %s, ignoring", client.ip)
			continue
		}

		if strings.HasPrefix(message, "/nick ") {
			parts := strings.SplitN(message, " ", 2)
			if len(parts) == 2 {
				newUsername := strings.TrimSpace(parts[1])
				if newUsername != "" && len(newUsername) < 20 {
					if s.isUsernameTaken(newUsername, client) {
						client.conn.WriteMessage(websocket.TextMessage, []byte(fmt.Sprintf("[Server] Username '%s' is already taken.", newUsername)))
					} else {
						oldUsername := client.username
						client.username = newUsername

						if oldUsername == "" {
							joinMsg := fmt.Sprintf("[Server] %s has joined the chat.", newUsername)
							log.Println(joinMsg)
							s.broadcastMessage(websocket.TextMessage, []byte(joinMsg), nil)
						} else if oldUsername != newUsername {
							changeMsg := fmt.Sprintf("[Server] %s changed nickname to %s.", oldUsername, newUsername)
							log.Println(changeMsg)
							s.broadcastMessage(websocket.TextMessage, []byte(changeMsg), nil)
						}

						client.conn.WriteMessage(websocket.TextMessage, []byte("[Server] Username set to "+newUsername))
					}
				} else {
					client.conn.WriteMessage(websocket.TextMessage, []byte("[Server] Invalid username."))
				}
			} else {
				client.conn.WriteMessage(websocket.TextMessage, []byte("[Server] Usage: /nick <username>"))
			}
			continue
		} else if message == "/list" {
			log.Printf("Client %s requested user list", client.ip)
			s.sendClientList(client)
			continue
		} else if message == "/listips" {
			log.Printf("Client %s (%s) requested IP list", client.username, client.ip)
			s.sendClientIPList(client)
			continue
		} else if message == "/exit" {
			log.Printf("Client %s (%s) requested disconnect.", client.username, client.ip)
			return nil
		} else if strings.HasPrefix(message, "/pm ") {
			parts := strings.SplitN(message, " ", 3)
			if len(parts) == 3 {
				targetUsername := strings.TrimSpace(parts[1])
				pmText := strings.TrimSpace(parts[2])
				if targetUsername != "" && pmText != "" {
					s.sendPrivateMessage(client, targetUsername, pmText)
				} else {
					client.conn.WriteMessage(websocket.TextMessage, []byte("[Server] Usage: /pm <username> <message>"))
				}
			} else {
				client.conn.WriteMessage(websocket.TextMessage, []byte("[Server] Usage: /pm <username> <message>"))
			}
			continue
		}

		if client.username == "" {
			err := client.conn.WriteMessage(websocket.TextMessage, []byte("[Server] Please set a username first using /nick <username>"))
			if err != nil {
				log.Printf("Error sending username prompt to %s: %v", client.ip, err)
				return err
			}
			continue
		}

		if err := s.broadcastMessage(messageType, []byte(message), client); err != nil {
			log.Printf("Error broadcasting message: %v", err)
		}
	}
}

func (s *Server) sendResponse(client *Client, messageType int, msg []byte) error {
	response := fmt.Sprintf("Hi dear client with %s IP; your message was: %s", client.ip, msg)
	return client.conn.WriteMessage(messageType, []byte(response))
}

func (s *Server) broadcastMessage(messageType int, msg []byte, sender *Client) error {
	var formattedMsg string
	if sender != nil {
		timestamp := time.Now().Format("15:04:05")
		formattedMsg = fmt.Sprintf("[%s] %s: %s", timestamp, sender.username, msg)
		log.Printf("Broadcasting from %s (%s): %s", sender.username, sender.ip, msg)
	} else {
		formattedMsg = string(msg)
		log.Println("Broadcasting server message:", formattedMsg)
	}

	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()

	for client := range s.clients {
		if client != sender {
			if err := client.conn.WriteMessage(messageType, []byte(formattedMsg)); err != nil {
				log.Printf("Error sending message to %s: %v. Removing client.", client.ip, err)
			}
		}
	}
	return nil
}

func (s *Server) sendClientList(requestingClient *Client) {
	var usernames []string
	s.clientsMux.RLock()
	for c := range s.clients {
		username := c.username
		if username == "" {
			continue
		}
		usernames = append(usernames, username)
	}
	s.clientsMux.RUnlock()

	listMsg := "[Server] Connected users: " + strings.Join(usernames, ", ")
	if err := requestingClient.conn.WriteMessage(websocket.TextMessage, []byte(listMsg)); err != nil {
		log.Printf("Error sending user list to %s: %v", requestingClient.ip, err)
	}
}

func (s *Server) isUsernameTaken(username string, requestingClient *Client) bool {
	s.clientsMux.Lock()
	defer s.clientsMux.Unlock()
	for c := range s.clients {
		if c != requestingClient && c.username == username {
			return true
		}
	}
	return false
}

func (s *Server) sendClientIPList(requestingClient *Client) {
	var clientIPs []string
	s.clientsMux.RLock()
	for c := range s.clients {
		clientIPs = append(clientIPs, c.ip)
	}
	s.clientsMux.RUnlock()

	listMsg := "[Server] Connected IPs: " + strings.Join(clientIPs, ", ")
	if err := requestingClient.conn.WriteMessage(websocket.TextMessage, []byte(listMsg)); err != nil {
		log.Printf("Error sending IP list to %s: %v", requestingClient.ip, err)
	}
}

func (s *Server) sendPrivateMessage(sender *Client, targetUsername string, message string) {
	var targetClient *Client
	s.clientsMux.RLock()
	for c := range s.clients {
		if c.username == targetUsername {
			targetClient = c
			break
		}
	}
	s.clientsMux.RUnlock()

	if targetClient == nil {
		errMsg := fmt.Sprintf("[Server] User '%s' not found.", targetUsername)
		if err := sender.conn.WriteMessage(websocket.TextMessage, []byte(errMsg)); err != nil {
			log.Printf("Error sending PM error to %s: %v", sender.ip, err)
		}
		return
	}

	timestamp := time.Now().Format("15:04:05")
	msgToTarget := fmt.Sprintf("[%s] [PM from %s]: %s", timestamp, sender.username, message)
	msgToSender := fmt.Sprintf("[%s] [PM to %s]: %s", timestamp, targetUsername, message)

	if err := targetClient.conn.WriteMessage(websocket.TextMessage, []byte(msgToTarget)); err != nil {
		log.Printf("Error sending PM to target %s: %v", targetClient.ip, err)
	}

	if err := sender.conn.WriteMessage(websocket.TextMessage, []byte(msgToSender)); err != nil {
		log.Printf("Error sending PM confirmation to sender %s: %v", sender.ip, err)
	}

	log.Printf("PM from %s to %s relayed.", sender.username, targetUsername)
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
