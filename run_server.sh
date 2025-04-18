#!/bin/bash

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting WebSocket Chat Server...${NC}"

# Check if there's already a server running on port 8000
if netstat -tuln | grep -q ":8000 "; then
    echo -e "${RED}Error: Port 8000 is already in use.${NC}"
    echo -e "${YELLOW}You might have another instance of the server running.${NC}"
    echo -e "To force kill any process using port 8000, run: ${YELLOW}./kill_server.sh${NC}"
    exit 1
fi

# Change to server directory
cd server || {
    echo -e "${RED}Error: server directory not found${NC}"
    exit 1
}

# Install dependencies if needed
echo -e "${BLUE}Checking dependencies...${NC}"
if ! go list -m github.com/gorilla/websocket &>/dev/null || \
   ! go list -m github.com/labstack/echo/v4 &>/dev/null; then
    echo -e "${YELLOW}Installing required dependencies...${NC}"
    go get github.com/gorilla/websocket
    go get github.com/labstack/echo/v4
    echo -e "${GREEN}Dependencies installed successfully.${NC}"
else
    echo -e "${GREEN}All dependencies already installed.${NC}"
fi

# Run the server
echo -e "${GREEN}Starting server...${NC}"
go run server.go

# Handle exit
echo -e "${RED}Server stopped.${NC}" 