#!/bin/bash

# Colors for terminal output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}Checking for processes using port 8000...${NC}"

# Get PIDs of processes using port 8000
SERVER_PIDS=$(lsof -t -i:8000)

if [ -z "$SERVER_PIDS" ]; then
    echo -e "${GREEN}No processes found using port 8000.${NC}"
    exit 0
fi

echo -e "${RED}Found processes using port 8000:${NC}"
echo "$SERVER_PIDS"

# Ask for confirmation
read -p "Kill these processes? (y/n): " -n 1 -r
echo
if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo -e "${YELLOW}Operation cancelled.${NC}"
    exit 0
fi

# Kill processes
for PID in $SERVER_PIDS; do
    echo -e "${YELLOW}Killing process $PID...${NC}"
    kill -9 "$PID" 2>/dev/null
done

echo -e "${GREEN}All processes on port 8000 have been terminated.${NC}" 