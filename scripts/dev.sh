#!/bin/bash
# Development server startup script for Unix/macOS/Linux
# Run from anywhere - navigates to project root automatically

set -e

# Get the directory where this script is located
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Navigate to project root (one level up from scripts folder)
cd "$SCRIPT_DIR/.."

echo ""
echo "🚀 Starting Full Stack Go Template Development Server..."
echo "   Project Root: $(pwd)"
echo ""

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if .env exists
if [ ! -f .env ]; then
    echo -e "${YELLOW}⚠️  No .env file found. Copying from .env.example...${NC}"
    cp .env.example .env
    echo -e "${GREEN}✓ Created .env file${NC}"
fi

# Check if node_modules exists for Tailwind
if [ ! -d "node_modules" ]; then
    echo -e "${YELLOW}📦 Installing npm dependencies...${NC}"
    npm install
    echo -e "${GREEN}✓ npm dependencies installed${NC}"
fi

# Build Tailwind CSS
echo -e "${YELLOW}🎨 Building Tailwind CSS...${NC}"
npm run build
echo -e "${GREEN}✓ Tailwind CSS compiled${NC}"

# Check if PostgreSQL is running (optional)
if command -v docker &> /dev/null; then
    if ! docker ps | grep -q go_starter_db; then
        echo -e "${YELLOW}🐘 Starting PostgreSQL container...${NC}"
        docker-compose up -d
        echo -e "${GREEN}✓ PostgreSQL started${NC}"
        sleep 2
    fi
fi

echo ""

# Check if dev server is already running on port 3000
SERVER_LISTENING=false
if command -v ss >/dev/null 2>&1; then
  if ss -ltn | grep -q ':3000 '; then SERVER_LISTENING=true; fi
elif command -v lsof >/dev/null 2>&1; then
  if lsof -iTCP:3000 -sTCP:LISTENING >/dev/null 2>&1; then SERVER_LISTENING=true; fi
else
  if netstat -an 2>/dev/null | grep -q 'LISTENING'; then
    if netstat -an 2>/dev/null | grep -q ':3000'; then SERVER_LISTENING=true; fi
  fi
fi

if [ "$SERVER_LISTENING" = true ]; then
  echo "✅ Dev server already running on http://localhost:3000"
  read -p "[?] Open preview in browser now? (y/N) " OPEN_PREVIEW
  case "$OPEN_PREVIEW" in
    [yY]|[yY][eE][sS])
      if command -v xdg-open >/dev/null 2>&1; then
        xdg-open "http://localhost:3000/"
      elif command -v open >/dev/null 2>&1; then
        open "http://localhost:3000/"
      else
        echo "Open http://localhost:3000/ in your browser"
      fi
      ;;
    *)
      echo "Skipping preview open."
      ;;
  esac
else
  read -p "[?] Dev server not running. Start it here to stream logs? (y/N) " START_SERVER
  case "$START_SERVER" in
    [yY]|[yY][eE][sS])
      echo -e "${GREEN}▶️  Starting Go server (logs will stream here)...${NC}"
      go run ./cmd/server
      ;;
    *)
      echo "Run 'go run ./cmd/server' to start the server."
      ;;
  esac
fi
