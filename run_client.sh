#!/bin/bash

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting WebSocket Chat Client...${NC}"

# Check if server is running on port 8000
if ! netstat -tuln | grep -q ":8000 "; then
    echo -e "${YELLOW}Warning: No server detected on port 8000.${NC}"
    echo -e "${YELLOW}Make sure the server is running before connecting.${NC}"
    echo -e "Run ${GREEN}./run_server.sh${NC} in another terminal to start the server."
    
    read -p "Continue anyway? (y/n): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        echo -e "${RED}Client startup cancelled.${NC}"
        exit 1
    fi
fi

# Change to client directory
cd client || {
    echo -e "${RED}Error: client directory not found${NC}"
    exit 1
}

# Install dependencies if needed
echo -e "${BLUE}Checking dependencies...${NC}"
if ! go list -m github.com/gorilla/websocket &>/dev/null || \
   ! go list -m github.com/charmbracelet/bubbletea &>/dev/null || \
   ! go list -m github.com/charmbracelet/bubbles &>/dev/null || \
   ! go list -m github.com/charmbracelet/lipgloss &>/dev/null; then
    echo -e "${YELLOW}Installing required dependencies...${NC}"
    go get github.com/gorilla/websocket
    go get github.com/charmbracelet/bubbletea
    go get github.com/charmbracelet/bubbles/textarea
    go get github.com/charmbracelet/bubbles/viewport
    go get github.com/charmbracelet/lipgloss
    echo -e "${GREEN}Dependencies installed successfully.${NC}"
else
    echo -e "${GREEN}All dependencies already installed.${NC}"
fi

# Run the client
echo -e "${GREEN}Starting client...${NC}"
go run client.go

# Handle exit
echo -e "${RED}Client stopped.${NC}" 