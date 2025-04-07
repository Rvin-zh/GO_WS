#!/bin/bash

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${GREEN}Starting WebSocket Chat Server...${NC}"

# Check if there's already a server running on port 8000
if netstat -tuln | grep -q ":8000 "; then
    echo -e "${RED}Error: Port 8000 is already in use.${NC}"
    echo -e "${YELLOW}You might have another instance of the server running.${NC}"
    echo -e "To force kill any process using port 8000, run: ${YELLOW}kill \$(lsof -t -i:8000)${NC}"
    exit 1
fi

# Change to server directory
cd server || {
    echo -e "${RED}Error: server directory not found${NC}"
    exit 1
}

# Run the server
go run server.go

# Handle exit
echo -e "${RED}Server stopped.${NC}" 