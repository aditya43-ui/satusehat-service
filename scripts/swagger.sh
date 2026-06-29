#!/bin/bash

set -eo pipefail

# Colors for output
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

echo -e "${BLUE}📚 Generating Flexible Swagger documentation...${NC}"

# 1. Periksa apakah swag CLI sudah terinstall
if ! command -v swag &> /dev/null; then
    echo -e "${YELLOW}⚠️  'swag' command not found. Installing github.com/swaggo/swag/cmd/swag@latest...${NC}"
    go install github.com/swaggo/swag/cmd/swag@latest
    
    # Tambahkan GOPATH/bin ke PATH jika belum ada
    export PATH=$PATH:$(go env GOPATH)/bin
fi

# 2. Format anotasi Swagger di seluruh project (Auto-format)
echo -e "${BLUE}📝 Formatting Swagger annotations...${NC}"
swag fmt -d ./cmd/api,./internal -g main.go

# 3. Generate Swagger docs
# -d : Root directory scanning (cmd/api untuk anotasi global, internal untuk semua modul domain DDD)
# -g : Entrypoint utama aplikasi relatif terhadap argumen -d pertama (main.go)
# -o : Output folder untuk file docs/swagger
# --parseDependency : Membaca struct dari external package (pkg/response dll)
# --parseInternal : Membaca struct/model di dalam folder internal
echo -e "${BLUE}⚙️  Building Swagger JSON/YAML files...${NC}"
if swag init -d ./cmd/api,./internal -g main.go -o ./docs/swagger --parseDependency --parseInternal; then
    echo -e "${GREEN}✅ Swagger documentation generated successfully in ./docs/swagger${NC}"
else
    echo -e "${RED}❌ Failed to generate Swagger documentation${NC}"
    exit 1
fi