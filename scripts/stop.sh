#!/bin/bash

# ==========================================
# CarFlow SaaS - Stop Script
# ==========================================

set -e

echo "🛑 Stopping CarFlow SaaS..."

# Check which compose files exist
if [ -f docker-compose.dev.yml ]; then
    docker-compose -f docker-compose.yml -f docker-compose.dev.yml down
else
    docker-compose down
fi

echo "✅ All services stopped."
echo ""
echo "💡 To remove volumes (delete data): docker-compose down -v"
echo ""
