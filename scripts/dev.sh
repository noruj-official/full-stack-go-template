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
echo -e "${GREEN}▶️  Starting Go server...${NC}"
echo "   http://localhost:3000"
echo ""

# Run the Go server
go run ./cmd/server
