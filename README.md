# Go WebSocket Chat Application

A simple terminal-based chat application built with Go and WebSockets. This project allows multiple users to connect to a central chat server and exchange messages in real-time.

## Features

- Terminal-based UI with [Bubble Tea](https://github.com/charmbracelet/bubbletea)
- Real-time WebSocket communication
- Username customization
- Private messaging
- Connection status display
- Automatic reconnection on connection loss
- User list display

## Requirements

- Go 1.16+
- Internet connection for package downloads

## Installation

1. Clone this repository:

   ```
   git clone <your-repository-url>
   cd <repository-directory>
   ```

2. Install dependencies:
   ```
   go mod download
   ```

## Usage

### Quick Start

Use the provided shell scripts to quickly start the server and client:

```bash
# Start the server
./run_server.sh

# Start a client (can run multiple instances)
./run_client.sh
```

### Manual Start

Alternatively, you can start the server and client manually:

```bash
# Start the server
cd server
go run server.go

# Start a client (in a new terminal)
cd client
go run client.go
```

## Chat Commands

The following commands are available in the chat:

| Command                    | Description                          | Example                |
| -------------------------- | ------------------------------------ | ---------------------- |
| `/nick <username>`         | Change your username                 | `/nick alice`          |
| `/pm <username> <message>` | Send a private message               | `/pm bob Hello there!` |
| `/list`                    | List all connected users             | `/list`                |
| `/listips`                 | List IP addresses of connected users | `/listips`             |
| `/exit`                    | Disconnect from the server           | `/exit`                |

## User Interface

The interface is divided into three main sections:

- Status bar (top): Shows connection status and your username
- Message area (middle): Displays chat messages
- Input area (bottom): For typing messages

## Connection Management

- The client automatically attempts to connect to the server at startup
- If the connection is lost, it automatically attempts to reconnect
- The status bar shows your current connection state
