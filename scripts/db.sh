#!/bin/bash

# ==========================================
# CarFlow SaaS - Database Management Script
# ==========================================

set -e

command=$1

case $command in
  "migrate")
    echo "🔄 Running database migrations..."
    docker-compose run --rm migrate
    echo "✅ Migrations complete!"
    ;;

  "migrate-down")
    echo "⬇️  Rolling back migrations..."
    docker-compose run --rm migrate -direction -1 -steps 1
    echo "✅ Rollback complete!"
    ;;

  "reset")
    echo "⚠️  WARNING: This will delete all data!"
    read -p "Are you sure? (yes/no): " confirm
    if [ "$confirm" = "yes" ]; then
      echo "🗑️  Dropping database..."
      docker-compose exec postgres psql -U carflow -d postgres -c "DROP DATABASE IF EXISTS carflow;"
      echo "📦 Creating fresh database..."
      docker-compose exec postgres psql -U carflow -d postgres -c "CREATE DATABASE carflow;"
      echo "🔄 Running migrations..."
      docker-compose run --rm migrate
      echo "✅ Database reset complete!"
    else
      echo "❌ Reset cancelled."
    fi
    ;;

  "seed")
    echo "🌱 Seeding database with demo data..."
    docker-compose run --rm migrate -path /app/migrations -database "postgres://carflow:carflow123@postgres:5432/carflow?sslmode=disable"
    echo "✅ Database seeded!"
    ;;

  "shell")
    echo "🐘 Opening PostgreSQL shell..."
    docker-compose exec postgres psql -U carflow -d carflow
    ;;

  "backup")
    timestamp=$(date +%Y%m%d_%H%M%S)
    backup_file="backups/carflow_${timestamp}.sql"
    mkdir -p backups
    echo "💾 Creating backup: $backup_file"
    docker-compose exec -T postgres pg_dump -U carflow carflow > $backup_file
    echo "✅ Backup created!"
    ;;

  "restore")
    if [ -z "$2" ]; then
      echo "❌ Please provide backup file path"
      echo "Usage: ./scripts/db.sh restore backups/carflow_20231123_120000.sql"
      exit 1
    fi
    backup_file=$2
    if [ ! -f "$backup_file" ]; then
      echo "❌ Backup file not found: $backup_file"
      exit 1
    fi
    echo "⚠️  WARNING: This will restore from backup and overwrite current data!"
    read -p "Are you sure? (yes/no): " confirm
    if [ "$confirm" = "yes" ]; then
      echo "📦 Restoring from backup..."
      cat $backup_file | docker-compose exec -T postgres psql -U carflow carflow
      echo "✅ Restore complete!"
    else
      echo "❌ Restore cancelled."
    fi
    ;;

  *)
    echo "CarFlow Database Management"
    echo ""
    echo "Usage: ./scripts/db.sh [command]"
    echo ""
    echo "Commands:"
    echo "  migrate          - Run pending migrations"
    echo "  migrate-down     - Rollback last migration"
    echo "  reset            - Drop and recreate database (WARNING: deletes all data)"
    echo "  seed             - Seed database with demo data"
    echo "  shell            - Open PostgreSQL shell"
    echo "  backup           - Create database backup"
    echo "  restore <file>   - Restore from backup file"
    echo ""
    ;;
esac
