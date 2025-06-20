#!/bin/bash

# Stop script for SLCW frontend development environment

echo "🛑 Stopping SLCW Frontend..."

docker-compose down

echo "✅ Frontend stopped successfully"