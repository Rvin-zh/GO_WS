#!/bin/bash

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
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

# Run the client
go run client.go

# Handle exit
echo -e "${RED}Client stopped.${NC}" 