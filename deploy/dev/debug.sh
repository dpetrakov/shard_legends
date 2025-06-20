#!/bin/bash

# Debug script for SLCW frontend development environment

echo "🔍 Debugging SLCW Frontend setup..."
echo ""

echo "1. Container status:"
docker compose ps

echo ""
echo "2. Container logs (last 20 lines):"
docker compose logs --tail=20 frontend

echo ""
echo "3. Port bindings:"
docker compose port frontend 3000

echo ""
echo "4. Network inspection:"
docker network ls | grep slcw

echo ""
echo "5. Testing local connection:"
curl -I http://localhost:8092 || echo "❌ Cannot connect to localhost:8092"

echo ""
echo "6. Container health:"
docker compose ps --format "table {{.Service}}\t{{.Status}}"