#!/bin/bash

# Script untuk menjalankan database migrations

set -e  # Exit immediately if a command exits with a non-zero status

echo "🔧 Running database migrations..."

# Warna untuk output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Load environment variables
if [ -f .env ]; then
    export $(cat .env | grep -v '^#' | xargs)
    echo -e "${GREEN}✅ Environment variables loaded${NC}"
else
    echo -e "${YELLOW}⚠️  .env file not found, using system environment${NC}"
fi

# Fungsi untuk menjalankan migration
run_migration() {
    local db_type=$1
    
    case $db_type in
        "postgres")
            echo -e "${YELLOW}📊 Running PostgreSQL migrations...${NC}"
            # Jika menggunakan golang-migrate
            if command -v migrate &> /dev/null; then
                migrate -path internal/infrastructure/database/migrations -database "postgresql://$DB_USER:$DB_PASSWORD@$DB_HOST:$DB_PORT/$DB_NAME?sslmode=disable" up
            # Jika menggunakan go run dengan gorm
            elif command -v go &> /dev/null; then
                go run cmd/migrate/main.go
            else
                echo -e "${RED}❌ No migration tool found${NC}"
                exit 1
            fi
            ;;
        "mysql")
            echo -e "${YELLOW}📊 Running MySQL migrations...${NC}"
            if command -v migrate &> /dev/null; then
                migrate -path internal/infrastructure/database/migrations -database "mysql://$DB_USER:$DB_PASSWORD@tcp($DB_HOST:$DB_PORT)/$DB_NAME" up
            elif command -v go &> /dev/null; then
                go run cmd/migrate/main.go
            else
                echo -e "${RED}❌ No migration tool found${NC}"
                exit 1
            fi
            ;;
        "mongodb")
            echo -e "${YELLOW}📊 Running MongoDB migrations...${NC}"
            # MongoDB biasanya tidak menggunakan traditional migrations
            echo -e "${GREEN}✅ MongoDB migrations skipped (schema-less)${NC}"
            ;;
        *)
            echo -e "${RED}❌ Unsupported database type: $db_type${NC}"
            exit 1
            ;;
    esac
}

# Cek apakah ada database configuration
if [ -z "$DB_TYPE" ]; then
    echo -e "${YELLOW}⚠️  DB_TYPE not set, defaulting to postgres${NC}"
    DB_TYPE="postgres"
fi

# Cek dan buat directory migrations jika belum ada
MIGRATIONS_DIR="internal/infrastructure/database/migrations"
if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo -e "${YELLOW}📁 Creating migrations directory...${NC}"
    mkdir -p $MIGRATIONS_DIR
fi

# Cek koneksi database
echo -e "${YELLOW}🔍 Testing database connection...${NC}"
if go run scripts/test_db_connection.go; then
    echo -e "${GREEN}✅ Database connection successful${NC}"
else
    echo -e "${RED}❌ Database connection failed${NC}"
    exit 1
fi

# Jalankan migration
run_migration $DB_TYPE

echo -e "${GREEN}✅ Database migrations completed successfully${NC}"