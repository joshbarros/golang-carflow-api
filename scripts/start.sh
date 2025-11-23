#!/bin/bash

# ==========================================
# CarFlow SaaS - Quick Start Script
# ==========================================

set -e

echo "🚀 Starting CarFlow SaaS Platform..."
echo ""

# Check if .env exists, if not copy from .env.example
if [ ! -f .env ]; then
    echo "📝 Creating .env file from .env.example..."
    cp .env.example .env
    echo "✅ .env file created. Please update it with your configuration."
    echo ""
fi

# Check if Docker is running
if ! docker info > /dev/null 2>&1; then
    echo "❌ Docker is not running. Please start Docker and try again."
    exit 1
fi

echo "🐳 Starting Docker containers..."
docker-compose up -d

echo ""
echo "⏳ Waiting for services to be ready..."
sleep 5

# Wait for postgres to be ready
echo "📊 Waiting for PostgreSQL..."
until docker-compose exec -T postgres pg_isready -U carflow > /dev/null 2>&1; do
    echo "   PostgreSQL is starting up..."
    sleep 2
done
echo "✅ PostgreSQL is ready!"

# Run migrations
echo ""
echo "🔄 Running database migrations..."
docker-compose run --rm migrate

echo ""
echo "✅ CarFlow SaaS is running!"
echo ""
echo "📍 Access points:"
echo "   API:            http://localhost:8080"
echo "   Health Check:   http://localhost:8080/healthz"
echo "   API Docs:       http://localhost:8080/api-docs"
echo "   Metrics:        http://localhost:8080/metrics"
echo ""
echo "💾 Database:"
echo "   PostgreSQL:     localhost:5432"
echo "   User:           carflow"
echo "   Password:       carflow123"
echo "   Database:       carflow"
echo ""
echo "📝 View logs: docker-compose logs -f"
echo "🛑 Stop: docker-compose down"
echo ""
