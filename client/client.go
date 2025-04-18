package main

import (
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textarea"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/gorilla/websocket"
)

var (
	serverAddr = "ws://localhost:8000/ws"
	style      = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FAFAFA")).
			Background(lipgloss.Color("#5A56E0"))
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF5370")).
			PaddingLeft(2)

	// Define styles for different UI elements
	senderStyle             = lipgloss.NewStyle().Foreground(lipgloss.Color("240")) // Dim for sender name
	messageStyle            = lipgloss.NewStyle().Foreground(lipgloss.Color("#FAFAFA"))
	serverMsgStyle          = lipgloss.NewStyle().Foreground(lipgloss.Color("#00ADD8")) // Cyan for server messages
	statusConnectedStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("#C3E88D")) // Green for connected
	statusDisconnectedStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#FFCB6B")) // Orange for disconnected
	statusReconnectingStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("#82AAFF")) // Blue for reconnecting
	viewportStyle           = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("#5A56E0")).
				PaddingRight(1). // Add some padding inside the border
				PaddingLeft(1)
	textareaStyle = lipgloss.NewStyle().
			PaddingTop(1)
)

type model struct {
	viewport     viewport.Model
	textarea     textarea.Model
	conn         *websocket.Conn
	messages     []string
	err          error
	connected    bool
	username     string
	reconnecting bool
	done         chan struct{}
	msgChan      chan receivedMsg // Add a channel for messages
}

type connectedMsg struct{ conn *websocket.Conn }
type disconnectedMsg struct{}
type receivedMsg struct{ text string }

func initialModel() model {
	ta := textarea.New()
	ta.Placeholder = "Type a message..."
	ta.Focus()

	ta.Prompt = "┃ "
	ta.CharLimit = 280

	ta.SetWidth(40)
	ta.SetHeight(3)

	// Remove cursor line styling
	ta.FocusedStyle.CursorLine = lipgloss.NewStyle()

	ta.ShowLineNumbers = false

	vp := viewport.New(40, 20)
	vp.Style = viewportStyle

	return model{
		textarea:     ta,
		viewport:     vp,
		messages:     []string{},
		username:     fmt.Sprintf("user-%d", rand.Intn(1000)),
		reconnecting: false,
		done:         make(chan struct{}),
		msgChan:      make(chan receivedMsg, 100), // Buffer 100 messages
	}
}

func (m model) Init() tea.Cmd {
	return tea.Batch(
		textarea.Blink,
		m.attemptConnection(),
		m.waitForMessages(), // Add a command to wait for messages
	)
}

// waitForMessages waits for messages on the channel
func (m model) waitForMessages() tea.Cmd {
	return func() tea.Msg {
		return <-m.msgChan
	}
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var (
		tiCmd tea.Cmd
		vpCmd tea.Cmd
		cmds  []tea.Cmd
	)

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			close(m.done)
			if m.conn != nil {
				err := m.conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseNormalClosure, ""))
				if err != nil {
					log.Printf("Error writing close message: %v", err)
				}
				m.conn.Close()
				m.conn = nil
			}
			return m, tea.Quit
		case tea.KeyEnter:
			if m.conn != nil && m.connected {
				message := strings.TrimSpace(m.textarea.Value())
				if message == "" {
					return m, nil
				}

				// Send any non-empty message to the server
				err := m.conn.WriteMessage(websocket.TextMessage, []byte(message))
				if err != nil {
					// Handle potential write errors (e.g., connection closed)
					m.err = fmt.Errorf("failed to send message: %v", err)
					log.Printf("Send error: %v", err)
					// Trigger disconnection logic if write fails
					tea.Println(disconnectedMsg{})
					return m, nil
				}

				// Display own message in the UI (add this code)
				if !strings.HasPrefix(message, "/") {
					timestamp := time.Now().Format("15:04:05")
					// Style own message differently
					styledMsg := lipgloss.JoinHorizontal(
						lipgloss.Top,
						senderStyle.Render(fmt.Sprintf("[%s] ", timestamp)),
						senderStyle.Foreground(lipgloss.Color("#C3E88D")).Render(m.username+":"),
						messageStyle.Render(" "+message),
					)
					m.messages = append(m.messages, styledMsg)
					m.viewport.SetContent(strings.Join(m.messages, "\n"))
					m.viewport.GotoBottom()
				}

				m.textarea.Reset()
				// Check if the message sent was a /nick command and update local username if successful
				if strings.HasPrefix(message, "/nick ") {
					parts := strings.SplitN(message, " ", 2)
					if len(parts) == 2 {
						newUsername := strings.TrimSpace(parts[1])
						// Basic client-side validation mirroring server (optional but good practice)
						if newUsername != "" && len(newUsername) < 20 {
							// Assume success for now, server confirms via broadcast
							m.username = newUsername
						}
					}
				}
			} else {
				m.err = fmt.Errorf("not connected to server")
			}
		}
	case connectedMsg:
		if m.conn != nil && m.conn != msg.conn {
			log.Println("Closing potentially old connection before assigning new one.")
			m.conn.Close()
		}
		m.conn = msg.conn
		m.connected = true
		m.err = nil
		m.reconnecting = false

		// Add a connection message to the UI
		connectMsg := fmt.Sprintf("[Server] Connected as %s", m.username)
		m.messages = append(m.messages, serverMsgStyle.Render(connectMsg))
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

		log.Printf("Client connected successfully. Starting listener.")
		// Return a command to send the /nick message AFTER connection is established
		nickCmd := func() tea.Msg {
			if m.conn == nil {
				return fmt.Errorf("cannot send nick: connection is nil")
			}
			nickMsg := fmt.Sprintf("/nick %s", m.username)
			err := m.conn.WriteMessage(websocket.TextMessage, []byte(nickMsg))
			if err != nil {
				log.Printf("Failed to send initial nick command: %v", err)
				// Handle error, maybe queue for retry or signal disconnection
				return disconnectedMsg{}
			}
			log.Printf("Sent initial nick command: %s", nickMsg)
			return nil // Indicate success, no state change needed directly
		}
		// Start listener AND send nick command
		go m.listenForMessages() // Restart listener for the new connection

		// Continue waiting for more messages
		cmds = append(cmds, m.waitForMessages())

		return m, tea.Batch(nickCmd, m.waitForMessages())
	case disconnectedMsg:
		m.connected = false
		if m.conn != nil {
			m.conn.Close()
			m.conn = nil
		}
		if !m.reconnecting {
			m.reconnecting = true
			m.err = fmt.Errorf("connection lost")
			// Add a disconnection message to the UI
			m.messages = append(m.messages, errorStyle.Render("Disconnected from server. Attempting to reconnect..."))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()

			// Continue waiting for more messages
			cmds = append(cmds, m.waitForMessages())

			return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				log.Println("Attempting to reconnect...")
				c, _, err := websocket.DefaultDialer.Dial(serverAddr, nil)
				if err != nil {
					log.Printf("Reconnection failed: %v", err)
					return fmt.Errorf("reconnection failed: %w", err)
				}
				return connectedMsg{conn: c}
			})
		}
	case error:
		currentErr := msg.(error)
		if m.reconnecting {
			m.err = currentErr
			log.Printf("Reconnection attempt failed: %v. Retrying in 5s.", currentErr)

			// Add error message to UI
			errMsg := fmt.Sprintf("Reconnection failed: %v. Retrying...", currentErr)
			m.messages = append(m.messages, errorStyle.Render(errMsg))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()

			// Continue waiting for messages
			cmds = append(cmds, m.waitForMessages())

			return m, tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
				return m.attemptConnection()()
			})
		} else {
			m.err = currentErr
			m.connected = false
			m.reconnecting = true
			log.Printf("Initial connection failed: %v. Starting reconnection process.", currentErr)

			// Add error message to UI
			errMsg := fmt.Sprintf("Connection failed: %v. Attempting to reconnect...", currentErr)
			m.messages = append(m.messages, errorStyle.Render(errMsg))
			m.viewport.SetContent(strings.Join(m.messages, "\n"))
			m.viewport.GotoBottom()

			// Continue waiting for messages
			cmds = append(cmds, m.waitForMessages())

			return m, tea.Tick(2*time.Second, func(t time.Time) tea.Msg {
				return m.attemptConnection()()
			})
		}
	case receivedMsg:
		log.Printf("Received raw message: %s", msg.text) // ADD LOGGING
		// Parse and style the received message
		styledMsg := msg.text // Default to raw text
		if strings.HasPrefix(msg.text, "[Server]") {
			styledMsg = serverMsgStyle.Render(msg.text)
		} else if strings.HasPrefix(msg.text, "[PM from") || strings.HasPrefix(msg.text, "[PM to") {
			// Simple styling for PMs, could be more sophisticated
			styledMsg = serverMsgStyle.Foreground(lipgloss.Color("#FFCB6B")).Render(msg.text) // Use orange for PMs
		} else if strings.HasPrefix(msg.text, "[") && strings.Contains(msg.text, "]") && strings.Contains(msg.text, ":") {
			// Attempt to parse timestamped message: [15:04:05] user: message
			timestampEnd := strings.Index(msg.text, "]")
			colonPos := strings.Index(msg.text, ":")
			if timestampEnd != -1 && colonPos != -1 && colonPos > timestampEnd {
				timestamp := msg.text[:timestampEnd+1]
				senderAndMsg := strings.TrimSpace(msg.text[timestampEnd+1:])
				parts := strings.SplitN(senderAndMsg, ":", 2)
				if len(parts) == 2 {
					sender := strings.TrimSpace(parts[0])
					message := strings.TrimSpace(parts[1])

					// Render parts with styles
					timestampStyled := senderStyle.Render(timestamp) // Use dim style for timestamp
					senderStyled := senderStyle.Render(sender + ":")
					messageStyled := messageStyle.Render(" " + message)

					// Highlight own messages
					if sender == m.username {
						senderStyled = senderStyle.Foreground(lipgloss.Color("#C3E88D")).Render(sender + ":")
					}
					styledMsg = lipgloss.JoinHorizontal(lipgloss.Top, timestampStyled, senderStyled, messageStyled)
				}
			}
		} // else: Keep raw styledMsg for unparsed lines

		log.Printf("Styled message: %s", styledMsg) // ADD LOGGING

		// Add the styled message to the list
		m.messages = append(m.messages, styledMsg)

		// Update the viewport content and scroll to bottom
		m.viewport.SetContent(strings.Join(m.messages, "\n"))
		m.viewport.GotoBottom()

		log.Printf("Viewport content set. Total messages: %d", len(m.messages)) // ADD LOGGING

		// Continue waiting for more messages
		cmds = append(cmds, m.waitForMessages())
	}

	m.textarea, tiCmd = m.textarea.Update(msg)
	m.viewport, vpCmd = m.viewport.Update(msg)

	cmds = append(cmds, tiCmd, vpCmd)
	return m, tea.Batch(cmds...)
}

func (m *model) listenForMessages() {
	if m.conn == nil {
		log.Println("listenForMessages called with nil connection")
		m.msgChan <- receivedMsg{text: "[Error] Not connected to server"}
		return
	}

	localConn := m.conn

	defer func() {
		if localConn != nil {
			localConn.Close()
		}
		log.Println("Listener goroutine stopped.")
	}()

	log.Println("Listener goroutine started.")

	for {
		select {
		case <-m.done:
			log.Println("Listener goroutine stopping due to done channel.")
			return
		default:
			if localConn == nil {
				log.Println("Listener goroutine found nil connection during loop.")
				m.msgChan <- receivedMsg{text: "[Error] Connection lost"}
				return
			}

			log.Println("Waiting to read message from server...")
			messageType, message, err := localConn.ReadMessage()
			if err != nil {
				log.Printf("Read error in listener: %v", err)
				netErr, ok := err.(net.Error)
				if ok && netErr.Timeout() {
					log.Println("Read timeout occurred.")
					continue
				}

				m.msgChan <- receivedMsg{text: "[Error] Connection error: " + err.Error()}
				return
			}
			log.Printf("Received message from server - Type: %d, Content: %s", messageType, string(message))
			// Send message to UI for processing through channel
			m.msgChan <- receivedMsg{text: string(message)}
		}
	}
}

func (m model) View() string {
	var status string
	var statusStyle lipgloss.Style
	if m.connected {
		status = fmt.Sprintf("Connected as %s", m.username)
		statusStyle = statusConnectedStyle
	} else if m.reconnecting {
		status = "Reconnecting..."
		statusStyle = statusReconnectingStyle
	} else {
		status = "Disconnected"
		statusStyle = statusDisconnectedStyle
	}

	var errorMsg string
	if m.err != nil {
		errorMsg = errorStyle.Render(fmt.Sprintf("Error: %v", m.err))
	}

	// Layout using lipgloss.JoinVertical
	statusLine := statusStyle.Render(status)
	if errorMsg != "" {
		statusLine = lipgloss.JoinVertical(lipgloss.Left,
			statusLine,
			errorMsg,
		)
	}

	return lipgloss.JoinVertical(lipgloss.Left,
		statusLine,
		m.viewport.View(),                       // Viewport now uses viewportStyle
		textareaStyle.Render(m.textarea.View()), // Apply style to textarea container
	)
}

func (m *model) attemptConnection() tea.Cmd {
	return func() tea.Msg {
		log.Println("Attempting initial connection...")
		c, _, err := websocket.DefaultDialer.Dial(serverAddr, nil)
		if err != nil {
			log.Printf("Initial connection failed: %v", err)
			return fmt.Errorf("initial connection failed: %w", err)
		}
		return connectedMsg{conn: c}
	}
}

func main() {
	// Set up logging to a file instead of stdout
	logFile, err := os.OpenFile("client.log", os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Error opening log file: %v\n", err)
		os.Exit(1)
	}
	defer logFile.Close()
	log.SetOutput(logFile)

	rand.Seed(time.Now().UnixNano())
	p := tea.NewProgram(initialModel(), tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Printf("Error running program: %v", err)
		os.Exit(1)
	}
}
