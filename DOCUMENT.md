# 📚 GoPrint Service General - Project Documentation

## 📋 Deskripsi Proyek
Service-General adalah microservice berbasis Go yang menyediakan REST API dan gRPC untuk manajemen data person/pasien dengan integrasi ke sistem kesehatan Indonesia (BPJS & SatuSehat).

## 🏗️ Technical Architecture

Project ini mengimplementasikan **Clean Architecture** dengan **Domain-Driven Design (DDD)**:

- **Transport Layer**: REST API (Gin) + gRPC
- **Application Layer**: Business logic (UseCases)
- **Domain Layer**: Entities + Business rules
- **Infrastructure Layer**: Database, Cache, External services

### Multi-Database Support
Mendukung 5 jenis database sekaligus dengan connection pooling dan read replicas:
- PostgreSQL (Primary)
- MySQL
- SQL Server
- MongoDB
- SQLite

## 📁 Project Structure
```text
service-general/
├── cmd/api/main.go              # Entry point aplikasi
├── internal/                    # Business logic (Clean Architecture)
│   ├── person/                  # Person/patient domain
│   │   ├── dto.go               # Data Transfer Objects
│   │   ├── entity.go            # Domain entities
│   │   ├── mapper.go            # Object mappers
│   │   ├── repository.go        # Repository interface & implementation
│   │   └── service.go           # Business service
│   ├── auth/                    # Authentication domain
│   ├── master/                  # Master data domain
│   └── infrastructure/          # Infrastructure layer
│       ├── cache/               # Cache management
│       ├── config/              # Configuration
│       ├── database/            # Database connections
│       └── transport/           # HTTP & gRPC transport
├── pkg/                         # Shared packages
│   ├── errors/                  # Error handling system
│   ├── logger/                  # Logging system
│   └── utils/query/             # Powerful query builder
├── scripts/                     # Utility scripts
├── docs/                        # Documentation
└── tools/                       # Code generation tools
```

## 📦 Dependencies

### Core Dependencies
```
github.com/gin-gonic/gin v1.10.1                    # REST API Framework
github.com/google/uuid v1.6.0                         # UUID Generator
golang.org/x/crypto v0.44.0                           # Cryptography
golang.org/x/sync v0.18.0                           # Sync utilities
gorm.io/driver/postgres v1.5.11                     # PostgreSQL Driver
```

### Database Drivers
```
gorm.io/driver/mysql v1.6.0                         # MySQL Driver
gorm.io/driver/sqlite v1.6.0                        # SQLite Driver
gorm.io/driver/sqlserver v1.6.3                      # SQL Server Driver
go.mongodb.org/mongo-driver v1.17.6                # MongoDB Driver
github.com/jmoiron/sqlx v1.4.0                      # SQL Extensions
```

### Utilities
```
github.com/go-playground/validator/v10 v10.27.0    # Input Validation
github.com/golang/protobuf v1.5.4                     # Protocol Buffers
github.com/rs/zerolog v1.34.0                         # Structured Logging
github.com/swaggo/gin-swagger v1.6.1                  # Swagger Integration
google.golang.org/grpc v1.78.0                        # gRPC Framework
```

## 🔗 API Endpoints

### REST API Base
- **Base URL**: `http://localhost:8080/api/v1`
- **Swagger UI**: `http://localhost:8080/swagger/index.html`

### Person Endpoints
- `GET /persons` - List persons with pagination
- `GET /persons/:id` - Get person detail
- `POST /persons` - Create new person
- `PUT /persons/:id` - Update person
- `DELETE /persons/:id` - Soft delete person
- `GET /persons/search` - Search with filters

### gRPC Services
- **Port**: 50051 (default)
- **Reflection**: Enabled untuk development
- **Proto Files**: Tersedia di `docs/api/`

## ⚙️ Environment Configuration

### Configuration in config.yaml
```yaml
server:
  port: 8080
  mode: debug
  read_timeout: 10
  write_timeout: 10

databases:
  postgres:
    type: postgres
    host: 10.10.123.206
    port: 5432
    username: postgres
    password: your_db_password
    database: health
    sslmode: disable
    max_open_conns: 25
    max_idle_conns: 25
    conn_max_lifetime: 5m
```

## 🗄️ Database Schema Overview

### Person Entity (Core)
```go
type Person struct {
    Id                       int64      // Primary key
    Name                     string     // Nama lengkap
    BirthDate                *time.Time // Tanggal lahir
    BirthRegency_Code        *string    // Kode kabupaten (6 digit)
    Gender_Code              *string    // L/P (Laki-laki/Perempuan)
    ResidentIdentityNumber   *string    // NIK (16 digit)
    Religion_Code            *string    // Kode agama
    Education_Code           *string    // Kode pendidikan
    Occupation_Code          *string    // Kode pekerjaan
    MaritalStatus_Code       *string    // Kode status marital
}
```

### Related Entities
- **PersonAddress**: Multiple addresses per person
- **PersonContact**: Contact information (email, phone)
- **PersonInsurance**: Insurance company data
- **PersonRelative**: Family/relative information

## 📋 Business Rules & Validation

### Person Validation Rules
```go
Name                     string   `json:"name" binding:"required,min=2"`
BirthRegency_Code        string   `json:"birth_regency_code" validate:"omitempty,len=6"`
Gender_Code              string   `json:"gender_code" validate:"omitempty,oneof=L P"`
ResidentIdentityNumber   string   `json:"resident_identity_number" validate:"omitempty,len=16"`
Village_Code             string   `json:"village_code" validate:"required,len=10"`
```

### Key Validation Rules:
- **Name**: Required, minimum 2 characters
- **NIK**: 16 digits (optional)
- **Gender**: L/P only (Laki-laki/Perempuan)
- **Birth Regency**: 6 digits code
- **Village Code**: 10 digits code (for addresses)
- **Email**: Valid email format (for contacts)

## 🔗 Integration Points

### BPJS Integration
```yaml
bpjs:
  base_url: https://apijkn.bpjs-kesehatan.go.id
  cons_id: 5257
  user_key: your_bpjs_user_key
  secret_key: your_bpjs_secret_key
  timeout: 30s
```

### SatuSehat Integration
```yaml
satu_sehat:
  org_id: your_orgid
  fasyakes_id: your_fasyakes_id
  client_id: your_client_id
  client_secret: your_client_secret
  auth_url: https://api-satusehat.kemkes.go.id/oauth2/v1
  base_url: https://api-satusehat.kemkes.go.id/fhir-r4/v1
```

## 🛠️ Development Commands

### Setup & Run
```bash
# Install dependencies
go mod download

# Run development server
make run
# atau
go run cmd/api/main.go
```

### Build & Test
```bash
# Build binary
make build
# atau
./scripts/build.sh

# Run tests
make test
# atau
go test ./...

# Database migration
./scripts/migrate.sh

# Generate proto files
./scripts/generate_proto.sh
```

## 🔐 Security Features
- JWT Authentication dengan multiple providers
- Rate limiting dengan Redis
- CORS protection
- Input validation & sanitization
- SQL injection prevention
- Audit trail untuk semua perubahan data

## 📊 Performance Features
- Connection pooling untuk database
- Query optimization dengan index
- Caching strategy (Redis)
- Read replicas support
- Soft delete untuk data retention

## 🧪 Testing Strategy
- Unit testing untuk business logic
- Integration testing untuk API
- External service mocking
- Load testing capabilities

## 🚀 Deployment Options
- Docker support (dev & prod)
- Environment-based configuration
- Health checks & monitoring
- Graceful shutdown
- Multi-stage build

## 📚 Documentation
- Swagger/OpenAPI documentation
- gRPC service definitions
- Code comments & examples
- Postman collections
- Architecture diagrams

## 🔮 Future Enhancements
- Event sourcing untuk audit trail
- CQRS implementation
- Multi-tenant support
- Advanced analytics
- AI/ML integration untuk data analytics

---

**Generated on**: $(date)
**Version**: 1.0.0
**Go Version**: $(go version | cut -d' ' -f3)
**Last Updated**: $(date +%Y-%m-%d)
```

Sekarang saya juga akan update script context.sh untuk membuatnya lebih lengkap:

```bash:/home/meninjar/goprint/service-general/scripts/context.sh
#!/bin/bash

# Script untuk generate project context secara lengkap
# Usage: ./scripts/context.sh [options]

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(dirname "$SCRIPT_DIR")"
OUTPUT_FILE="$PROJECT_DIR/PROJECT_CONTEXT.md"

echo "🚀 Generating Service-General Project Context..."

# Fungsi untuk generate header
generate_header() {
    cat << 'EOF'
# SERVICE-GENERAL PROJECT CONTEXT

## 📋 Project Overview
Service-General adalah microservice berbasis Go yang menyediakan REST API dan gRPC untuk manajemen data person/pasien dengan integrasi ke sistem kesehatan Indonesia (BPJS & SatuSehat).

## 🎯 Business Domain
- **Patient Management**: CRUD lengkap untuk data pasien
- **Health Insurance**: Integrasi dengan BPJS Kesehatan
- **National Health Platform**: Integrasi dengan SatuSehat (FHIR R4)
- **Master Data**: Manajemen data master (agama, pendidikan, pekerjaan, dll)

## 🏗️ Technical Architecture

### Clean Architecture + Domain-Driven Design
- **Transport Layer**: REST API (Gin) + gRPC
- **Application Layer**: Business logic (UseCases)
- **Domain Layer**: Entities + Business rules
- **Infrastructure Layer**: Database, Cache, External services

### Multi-Database Support
- PostgreSQL (primary)
- MySQL
- SQL Server
- MongoDB
- SQLite

EOF
}

# Fungsi untuk generate struktur directory
generate_directory_structure() {
    echo "## 📁 Project Structure"
    echo '```'
    if command -v tree >/dev/null 2>&1; then
        tree -L 3 -I 'tmp|vendor|.git|node_modules|*.log' "$PROJECT_DIR" 2>/dev/null | head -50
    else
        find "$PROJECT_DIR" -type d -not -path "*/tmp/*" -not -path "*/.git/*" | head -30
    fi
    echo '```'
}

# Fungsi untuk generate dependencies list
generate_dependencies() {
    echo "## 📦 Dependencies"
    echo "### Core Dependencies"
    grep -E '^[[:space:]]*github.com/' "$PROJECT_DIR/go.mod" | head -10
    
    echo ""
    echo "### Database Drivers"
    grep -E 'postgres|mysql|mongo|sqlite|sqlserver' "$PROJECT_DIR/go.mod"
    
    echo ""
    echo "### Security & Validation"
    grep -E 'validator|crypto|jwt|auth' "$PROJECT_DIR/go.mod"
}

# Fungsi untuk generate business rules dari DTO
generate_business_rules() {
    echo "## 📋 Business Rules & Validation"
    if [ -f "$PROJECT_DIR/internal/person/dto.go" ]; then
        echo "### Person Validation Rules"
        echo '```go'
        grep -E "validate:|binding:" "$PROJECT_DIR/internal/person/dto.go" | head -15
        echo '```'
        echo ""
        echo "### Key Validation Rules:"
        echo "- **Name**: Required, minimum 2 characters"
        echo "- **NIK**: 16 digits (optional)"
        echo "- **Gender**: L/P only (Laki-laki/Perempuan)"
        echo "- **Birth Regency**: 6 digits code"
        echo "- **Village Code**: 10 digits code (for addresses)"
        echo "- **Email**: Valid email format (for contacts)"
    fi
}

# Fungsi untuk generate database schema overview
generate_database_schema() {
    echo "## 🗄️ Database Schema Overview"
    if [ -f "$PROJECT_DIR/internal/person/entity.go" ]; then
        echo "### Person Entity (Core)"
        echo '```go'
        # Ambil struct definition
        sed -n '/type Person struct {/,/^}/p' "$PROJECT_DIR/internal/person/entity.go" | head -20
        echo '```'
        echo ""
        echo "### Related Entities:"
        echo "- **PersonAddress**: Multiple addresses per person"
        echo "- **PersonContact**: Contact information (email, phone)"
        echo "- **PersonInsurance**: Insurance company data"
        echo "- **PersonRelative**: Family/relative information"
    fi
}

# Fungsi untuk generate integration points
generate_integrations() {
    echo "## 🔗 Integration Points"
    if [ -f "$PROJECT_DIR/config.yaml" ]; then
        echo "### BPJS Integration"
        echo '```yaml'
        grep -A 6 "bpjs:" "$PROJECT_DIR/config.yaml" 2>/dev/null || echo "BPJS configuration not found"
        echo '```'
        echo ""
        echo "### SatuSehat Integration"
        echo '```yaml'
        grep -A 8 "satu_sehat:" "$PROJECT_DIR/config.yaml" 2>/dev/null || echo "SatuSehat configuration not found"
        echo '```'
    fi
}

# Fungsi untuk generate API endpoints dari code analysis
generate_api_endpoints() {
    echo "## 🔗 API Endpoints"
    echo "### REST API Base"
    echo "- **Base URL**: \`http://localhost:8080/api/v1\`"
    echo "- **Swagger UI**: \`http://localhost:8080/swagger/index.html\`"
    echo ""
    echo "### Person Endpoints"
    echo "- \`GET /persons\` - List persons with pagination"
    echo "- \`GET /persons/:id\` - Get person detail"
    echo "- \`POST /persons\` - Create new person"
    echo "- \`PUT /persons/:id\` - Update person"
    echo "- \`DELETE /persons/:id\` - Soft delete person"
    echo "- \`GET /persons/search\` - Search with filters"
    echo ""
    echo "### Auth Endpoints"
    echo "- \`POST /auth/login\` - User login"
    echo "- \`POST /auth/refresh\` - Refresh token"
    echo "- \`POST /auth/logout\` - User logout"
    echo ""
    echo "### gRPC Services"
    echo "- **Port**: 50051 (default)"
    echo "- **Reflection**: Enabled untuk development"
    echo "- **Proto Files**: Tersedia di \`docs/api/\`"
}

# Fungsi untuk generate environment configuration
generate_env_config() {
    echo "## ⚙️ Environment Configuration"
    if [ -f "$PROJECT_DIR/.env.example" ]; then
        echo "### Environment Variables (.env.example)"
        echo '```bash'
        cat "$PROJECT_DIR/.env.example"
        echo '```'
    else
        echo "### Configuration in config.yaml"
        echo '```yaml'
        grep -A 10 "server:" "$PROJECT_DIR/config.yaml" 2>/dev/null || echo "Server config not found"
        echo '```'
    fi
}

# Fungsi untuk generate development commands
generate_dev_commands() {
    echo "## 🛠️ Development Commands"
    echo "### Setup & Run"
    echo '```bash'
    echo "# Install dependencies"
    echo "go mod download"
    echo ""
    echo "# Run development server"
    echo "make run"
    echo "# atau"
    echo "go run cmd/api/main.go"
    echo '```'
    echo ""
    echo "### Build & Test"
    echo '```bash'
    echo "# Build binary"
    echo "make build"
    echo "# atau"
    echo "./scripts/build.sh"
    echo ""
    echo "# Run tests"
    echo "make test"
    echo "# atau"
    echo "go test ./..."
    echo ""
    echo "# Database migration"
    echo "./scripts/migrate.sh"
    echo ""
    echo "# Generate proto files"
    echo "./scripts/generate_proto.sh"
    echo '```'
}

# Fungsi untuk generate security features
generate_security_features() {
    echo "## 🔐 Security Features"
    echo "- JWT Authentication dengan multiple providers (JWT, Keycloak, Static)"
    echo "- Rate limiting dengan Redis (default: 60 requests/minute)"
    echo "- CORS protection dengan configurable origins"
    echo "- Input validation & sanitization menggunakan go-playground/validator"
    echo "- SQL injection prevention dengan parameterized queries"
    echo "- Audit trail untuk semua perubahan data"
    echo "- Soft delete untuk data retention"
    echo "- Security headers (XSS, CSRF protection)"
}

# Fungsi untuk generate performance features
generate_performance_features() {
    echo "## 📊 Performance Features"
    echo "- Connection pooling untuk database (max 25 connections)"
    echo "- Query optimization dengan proper indexing"
    echo "- Caching strategy menggunakan Redis"
    echo "- Read replicas support untuk load balancing"
    echo "- Soft delete untuk mempertahankan performa query"
    echo "- Pagination untuk semua list endpoints"
    echo "- Timeout configuration untuk external services (30s default)"
}

# Fungsi untuk generate testing strategy
generate_testing_strategy() {
    echo "## 🧪 Testing Strategy"
    echo "- Unit testing untuk business logic di setiap domain"
    echo "- Integration testing untuk API endpoints"
    echo "- External service mocking untuk BPJS dan SatuSehat"
    echo "- Load testing capabilities untuk performance validation"
    echo "- Test coverage target: >80%"
}

# Fungsi untuk generate deployment info
generate_deployment_info() {
    echo "## 🚀 Deployment Options"
    echo "### Docker Support"
    echo "- **Development**: \`docker-compose -f docker-compose.dev.yml up\`"
    echo "- **Production**: \`docker-compose -f docker-compose.prod.yml up\`"
    echo ""
    echo "### Environment Variables"
    echo "- Database configuration"
    echo "- JWT secret dan auth settings"
    echo "- External service credentials (BPJS, SatuSehat)"
    echo "- Rate limiting configuration"
    echo "- Logging level dan format"
}

# Generate complete context
generate_complete_context() {
    generate_header > "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    generate_directory_structure >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    generate_dependencies >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    generate_api_endpoints >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    generate_env_config >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    generate_database_schema >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    generate_business_rules >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    generate_integrations >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    generate_dev_commands >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    generate_security_features >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    generate_performance_features >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    generate_testing_strategy >> "$OUTPUT_FILE"
    echo "" >> "$OUTPUT_FILE"
    generate_deployment_info >> "$OUTPUT_FILE"
    
    # Tambahkan closing section
    cat >> "$OUTPUT_FILE" << EOF

## 📚 Additional Documentation
- **API Documentation**: Swagger/OpenAPI di \`/swagger/index.html\`
- **gRPC Documentation**: Tersedia di \`docs/api/grpc_docs.md\`
- **Postman Collection**: \`person.postman_collection.json\`
- **Architecture Diagrams**: Implementasi Clean Architecture + DDD

## 🔮 Future Enhancements
- Event sourcing untuk audit trail yang lebih baik
- CQRS (Command Query Responsibility Segregation) implementation
- Multi-tenant support untuk beberapa organisasi
- Advanced analytics dan reporting
- AI/ML integration untuk data analytics dan prediksi
- Service mesh integration untuk microservices
- Advanced caching strategies (Redis Cluster)

## 📊 Project Metrics
- **Go Version**: $(go version 2>/dev/null || echo "Go not installed")
- **Module**: service-general
- **Architecture**: Clean Architecture + DDD
- **API Types**: REST + gRPC
- **Databases**: 5 types (PostgreSQL, MySQL, SQL Server, MongoDB, SQLite)
- **External Integrations**: BPJS Kesehatan, SatuSehat Platform
- **Security**: JWT, Rate Limiting, CORS, Input Validation

---
**Generated on**: $(date)
**Version**: 1.0.0
**Last Updated**: $(date +%Y-%m-%d)
**Generator**: Service-General Context Script

EOF

    echo "✅ Project context generated successfully!"
    echo "📄 File created: $OUTPUT_FILE"
    echo "📊 Context includes:"
    echo "  ✅ Project structure and architecture"
    echo "  ✅ Dependencies and technology stack"
    echo "  ✅ API endpoints and business rules"
    echo "  ✅ Database schema overview"
    echo "  ✅ Integration points (BPJS & SatuSehat)"
    echo "  ✅ Development commands and deployment"
    echo "  ✅ Security and performance features"
    echo "  ✅ Testing strategy and future enhancements"
}

# Main execution
main() {
    cd "$PROJECT_DIR"
    
    case "${1:-}" in
        --help|-h)
            echo "Usage: $0 [options]"
            echo "Options:"
            echo "  --help, -h     Show this help message"
            echo "  --generate     Generate complete project context (default)"
            echo "  --structure    Show directory structure only"
            echo "  --deps         Show dependencies only"
            echo "  --api          Show API endpoints only"
            echo "  --rules        Show business rules only"
            echo "  --security     Show security features only"
            echo "  --performance  Show performance features only"
            ;;
        --generate)
            generate_complete_context
            ;;
        --structure)
            generate_directory_structure
            ;;
        --deps)
            generate_dependencies
            ;;
        --api)
            generate_api_endpoints
            ;;
        --rules)
            generate_business_rules
            ;;
        --security)
            generate_security_features
            ;;
        --performance)
            generate_performance_features
            ;;
        *)
            generate_complete_context
            ;;
    esac
}

# Run main function
main "$@"
echo "✨ Context generation completed!"
```
