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

2. Make the scripts executable:
   ```
   chmod +x run_server.sh run_client.sh kill_server.sh
   ```

## Usage

### Quick Start

Use the provided shell scripts to quickly start the server and client. The scripts will automatically check for and install any missing dependencies:

```bash
# Start the server (in one terminal)
./run_server.sh

# Start a client (in another terminal)
./run_client.sh
```

You can run multiple client instances to simulate multiple users chatting.

### Managing Server Processes

If you encounter port conflicts or need to kill the server:

```bash
# Kill any processes using port 8000
./kill_server.sh
```

### Manual Start (Advanced)

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

## License

[MIT License](LICENSE)

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.
