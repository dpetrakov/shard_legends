#!/bin/bash

# Start script for SLCW frontend development environment

set -e

echo "🚀 Starting SLCW Frontend (Development)..."

# Check if .env file exists, if not copy from example
if [ ! -f .env ]; then
    echo "📋 Creating .env file from .env.example..."
    cp .env.example .env
    echo "⚠️  Please update .env file with your actual configuration values!"
fi

# Build and start containers
echo "🔨 Building Docker images..."
docker-compose build

echo "🏃 Starting containers..."
docker-compose up -d

echo "✅ Frontend is starting on port 8092"
echo ""
echo "🌐 Access points:"
echo "  - Direct: http://localhost:8092"
echo "  - Via nginx: https://dev.slcw.dimlight.online"
echo ""
echo "📝 Useful commands:"
echo "  - View logs: docker-compose logs -f frontend"
echo "  - Stop: ./stop.sh"
echo "  - Rebuild: docker-compose build frontend"