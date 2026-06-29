# GoPrint Service General API

A comprehensive Go microservice providing REST API and gRPC support. Built with **Clean Architecture**, **Domain-Driven Design (DDD)**, and **CQRS Pattern**, featuring multi-database support, intelligent caching, and direct integration with Indonesian Health Systems (BPJS VClaim & SatuSehat).

## 🏗️ Architecture

This project strongly follows **Clean Architecture** principles, **CQRS Pattern** (Command and Query Responsibility Segregation), and **Domain-Driven Design (DDD)** patterns:

```
┌─────────────────────────────────────────────────────────────┐
│                    Transport Layer                           │
│  ┌─────────────────────┐    ┌─────────────────────────────┐  │
│  │   REST API (Gin)    │    │      gRPC API              │  │
│  └─────────────────────┘    └─────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                   Application Layer                          │
│  ┌─────────────────────┐    ┌─────────────────────────────┐  │
│  │    Auth Service     │    │   Role/Master Service       │  │
│  └─────────────────────┘    └─────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                     Domain Layer                             │
│  ┌─────────────────────┐    ┌─────────────────────────────┐  │
│  │  Command Repository │    │     Query Repository        │  │
│  └─────────────────────┘    └─────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
                                │
┌─────────────────────────────────────────────────────────────┐
│                Infrastructure Layer                           │
│  ┌─────────────────────┐    ┌─────────────────────────────┐  │
│  │ Multi-DB & Cache    │    │  BPJS VClaim & SatuSehat    │  │
│  └─────────────────────┘    └─────────────────────────────┘  │
└─────────────────────────────────────────────────────────────┘
```

## 🚀 Features

### Core Features
- ✅ **Dual API Support**: REST (Gin) + gRPC endpoints
- ✅ **Clean Architecture**: Domain → Application → Infrastructure → Transport
- ✅ **Domain-Driven Design**: Bounded Contexts (Auth, Patient)
- ✅ **Multi-Database Support**: PostgreSQL, MySQL, SQL Server, MongoDB, SQLite
- ✅ **Connection Pooling**: Optimized database connections with health monitoring
- ✅ **Read Replicas**: Load balancing with round-robin strategy
- ✅ **Audit Trail**: Complete audit logging for data changes
- ✅ **Soft Delete**: Data retention with active/inactive views

### Security Features
- ✅ **JWT Authentication**: Multiple auth providers (JWT, Keycloak, Static)
- ✅ **CORS Support**: Configurable cross-origin resource sharing
- ✅ **Rate Limiting**: Configurable request rate limiting
- ✅ **Security Headers**: XSS, CSRF, and other security headers
- ✅ **Input Validation**: Comprehensive data validation
- ✅ **SQL Injection Prevention**: Parameterized queries and input sanitization

### Integration Features
- ✅ **BPJS Integration**: Indonesian health insurance system integration
- ✅ **SatuSehat Integration**: Indonesian health data platform
- ✅ **Swagger Documentation**: Auto-generated API documentation
- ✅ **Health Checks**: Database and service health monitoring
- ✅ **Metrics Collection**: Application performance metrics

### Development Features
- ✅ **Hot Reload**: Development server with auto-restart
- ✅ **Code Generation**: Automated entity and service generation
- ✅ **Docker Support**: Containerized deployment
- ✅ **Environment Configuration**: YAML + Environment variables
- ✅ **Structured Logging**: Comprehensive logging system

## 📁 Project Structure

```text
service-general/                    # 🏠 Root Project
├── cmd/api/                        # 🚀 Application Entry Point (main.go)
├── internal/                       # 🔒 Private Application Code
│   ├── auth/                       # 🔐 Authentication Context
│   ├── master/                     # 📊 Master Data Context
│   ├── infrastructure/             # 🏗️ Technical Infrastructure (DB, HTTP, gRPC)
│   └── interfaces/                 # 🔌 External Services (BPJS, SatuSehat)
├── pkg/                            # 📦 Public Reusable Libraries
│   ├── errors/                     # 🚨 Error handling system
│   ├── logger/                     # 📝 Logging system
│   └── utils/query/                # 🔍 Advanced Query Builder
├── scripts/                        # 🔄 Automation Scripts
├── tools/                          # 🔧 Development Tools
├── docs/                           # 📚 Documentation
├── config.yaml                     # ⚙️ Configuration file
└── Makefile                        # 🔨 Build automation
```

## � Requirements

- Go 1.21+
- PostgreSQL 12+ (recommended)
- Redis 6+ (for rate limiting)
- Docker & Docker Compose (optional)

## 🛠️ Installation

### 1. Clone the Repository
```bash
git clone <repository-url>
cd service-general
```

### 2. Install Dependencies
```bash
go mod download
go mod tidy
```

### 3. Install Development Tools
```bash
# Install gRPC code generation tools
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-go@v1.31.0
go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v1.31.0
go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.31.0

# Install Swagger generation tool
go install github.com/swaggo/swag/cmd/swag@v1.16.2
```

### 4. Setup Database
```bash
# Create PostgreSQL database
createdb api_service

# Run migrations
psql -d api_service -f migrations/001_initial_schema.sql
```

### 5. Configure Environment
```bash
# Copy configuration template
cp config.yaml.example config.yaml

# Edit configuration
vim config.yaml
```

### 6. Run the Application
```bash
# Development mode
make run

# Or manually
go run cmd/api/main.go
```

## 🔧 Configuration

### Environment Variables
```bash
# Server Configuration
PORT=8080
GIN_MODE=debug

# Database Configuration
DB_HOST=localhost
DB_PORT=5432
DB_USERNAME=postgres
DB_PASSWORD=password
DB_DATABASE=api_service

# JWT Configuration
JWT_SECRET=your-secret-key

# BPJS Configuration
BPJS_CONSID=your-cons-id
BPJS_SECRETKEY=your-secret-key
BPJS_USERKEY=your-user-key

# Rate Limiting
RATE_LIMIT_REQUESTS_PER_MINUTE=60
```

### Configuration File (config.yaml)
```yaml
server:
  rest:
    enabled: true
    port: 8080
  grpc:
    enabled: true
    port: 50051
  mode: debug
  read_timeout: 30
  write_timeout: 30

databases:
  primary:
    type: postgres
    host: localhost
    port: 5432
    username: postgres
    password: password
    database: api_service
    max_open_conns: 25
    max_idle_conns: 25
    conn_max_lifetime: 5m

auth:
  type: jwt
  static_tokens: []

bpjs:
  base_url: https://apijkn.bpjs-kesehatan.go.id
  cons_id: ""
  secret_key: ""
  user_key: ""
  timeout: 30s

security:
  trusted_origins:
    - http://localhost:3000
    - http://localhost:8080
  max_input_length: 500
  rate_limit:
    requests_per_minute: 60
```

## 📚 API Documentation

### REST API Endpoints

#### Authentication
- `POST /api/v1/auth/login` - User login
- `POST /api/v1/auth/logout` - User logout
- `POST /api/v1/auth/refresh` - Refresh token
- `GET /api/v1/auth/profile` - Get user profile

#### Patients
- `GET /api/v1/patients` - List patients (paginated)
- `GET /api/v1/patients/{id}` - Get patient by ID
- `POST /api/v1/patients` - Create new patient
- `PUT /api/v1/patients/{id}` - Update patient
- `DELETE /api/v1/patients/{id}` - Delete patient

#### Medical Records
- `GET /api/v1/patients/{id}/medical-records` - Get patient medical records
- `POST /api/v1/patients/{id}/medical-records` - Create medical record
- `PUT /api/v1/medical-records/{id}` - Update medical record

#### Health Check
- `GET /api/v1/health` - Application health status
- `GET /api/v1/health/database` - Database connection status

### gRPC API Endpoints

#### Auth Service
```protobuf
service AuthService {
  rpc Login(LoginRequest) returns (LoginResponse);
  rpc Logout(LogoutRequest) returns (LogoutResponse);
  rpc RefreshToken(RefreshTokenRequest) returns (RefreshTokenResponse);
  rpc GetProfile(GetProfileRequest) returns (GetProfileResponse);
}
```

#### Patient Service
```protobuf
service PatientService {
  rpc ListPatients(ListPatientsRequest) returns (ListPatientsResponse);
  rpc GetPatient(GetPatientRequest) returns (GetPatientResponse);
  rpc CreatePatient(CreatePatientRequest) returns (CreatePatientResponse);
  rpc UpdatePatient(UpdatePatientRequest) returns (UpdatePatientResponse);
  rpc DeletePatient(DeletePatientRequest) returns (DeletePatientResponse);
}
```

## 🐳 Docker Deployment

### Build Docker Image
```bash
docker build -t service-general .
```

### Run with Docker Compose
```bash
docker-compose up -d
```

### Docker Compose Configuration
```yaml
version: '3.8'
services:
  service-general:
    build: .
    ports:
      - "8080:8080"
      - "50051:50051"
    environment:
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: api_service
      POSTGRES_USER: postgres
      POSTGRES_PASSWORD: password
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"

volumes:
  postgres_data:
```

## 🧪 Testing

### Run Tests
```bash
# Run all tests
make test

# Run tests with coverage
make test-coverage

# Run integration tests
make test-integration
```

### Test Examples
```bash
# Test health endpoint
curl http://localhost:8080/api/v1/health

# Test patient creation
curl -X POST http://localhost:8080/api/v1/patients \
  -H "Content-Type: application/json" \
  -d '{
    "nik": "1234567890123456",
    "name": "John Doe",
    "date_of_birth": "1990-01-01",
    "gender": "L"
  }'
```

## 📊 Monitoring & Metrics

### Health Check Endpoints
- `/api/v1/health` - Overall application health
- `/api/v1/health/database` - Database connection status
- `/api/v1/health/external` - External service status

### Metrics Collection
- Request count and duration
- Database connection pool metrics
- Error rates and types
- Authentication success/failure rates

### Logging
- Structured JSON logging
- Request/Response logging
- Error stack traces
- Performance metrics

## 🔒 Security

### Authentication
- JWT tokens with configurable expiration
- Multiple auth providers (JWT, Keycloak, Static)
- Token refresh mechanism
- Session management

### Authorization
- Role-based access control (RBAC)
- Permission-based endpoints
- API key authentication
- Rate limiting per user/IP

### Data Protection
- SQL injection prevention
- XSS protection headers
- CSRF protection
- Input validation and sanitization

## 🚀 Performance

### Database Optimization
- Connection pooling with configurable limits
- Read replica support with load balancing
- Query optimization and indexing
- Health monitoring and automatic reconnection

### Caching
- Redis-based session storage
- Application-level caching
- Database query result caching
- Static asset caching

### Scalability
- Horizontal scaling support
- Load balancer ready
- Microservice architecture
- Container orchestration ready

## 🛠️ Development

### Code Generation
```bash
# Generate new entity
make generate entity=Patient

# Generate service for entity
make generate service=Patient

# Generate CRUD handlers
make generate crud=Patient
```

### Database Migrations
```bash
# Create new migration
make migration create=add_new_table

# Run migrations
make migration up

# Rollback migration
make migration down

# Create a new migration using goose
goose -dir internal/infrastructure/database/sql create create_users_table sql
```

### API Documentation
```bash
# Generate Swagger docs
make swagger

# Serve Swagger UI
make swagger-serve
```

## 📈 Production Deployment

### Environment Setup
```bash
# Production environment variables
export GIN_MODE=release
export PORT=8080
export DB_HOST=your-production-db
export JWT_SECRET=your-production-secret
```

### Monitoring Setup
- Prometheus metrics endpoint
- Grafana dashboard templates
- Log aggregation with ELK stack
- Alert configuration

### CI/CD Pipeline
- GitHub Actions workflow
- Automated testing
- Docker image building
- Deployment automation

## 🤝 Contributing

1. Fork the repository
2. Create feature branch (`git checkout -b feature/amazing-feature`)
3. Commit changes (`git commit -m 'Add amazing feature'`)
4. Push to branch (`git push origin feature/amazing-feature`)
5. Open Pull Request

### Code Style
- Follow Go conventions
- Use meaningful variable names
- Add comments for complex logic
- Write unit tests for new features

## 📄 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For support and questions:
- Create an issue on GitHub
- Check the [Wiki](wiki) for documentation
- Join our [Discord](https://discord.gg/your-server) community

## 🗺️ Roadmap

service-general/                    # 🏠 Root Project
├── 📁 cmd/                         # 🚀 Application Entry Points
│   └── 📁 api/                     # API Application
│       └── 📄 main.go              # 🎯 Main entry point (wires everything together)
│
├── 📁 internal/                    # 🔒 Private Application Code
│   ├── 📁 auth/                    # 🔐 Bounded Context: Authentication
│   │   ├── 📁 domain/              # 🏛️ Core auth entities & business rules
│   │   │   └── 📄 model.go         # User, Role entities
│   │   ├── 📁 repository/          # 💾 Data access interfaces
│   │   │   └── 📄 repo.go          # Auth repository interface
│   │   └── 📁 usecase/             # ⚙️ Application business logic
│   │       └── 📄 service.go       # Auth services (login, register, etc)
│   │ 
│   ├── 📁 master/                  # 📊 Bounded Context: Master Data
│   │   ├── 📄 dto.go               # 📋 Data Transfer Objects
│   │   ├── 📄 entity.go            # 🗃️ Master entities
│   │   ├── 📄 repository.go        # 🏪 Master repository interface
│   │   └── 📄 service.go           # 🛠️ Master business logic
│   │ 
│   ├── 📁 infrastructure/          # 🏗️ Technical Infrastructure Layer
│   │   ├── 📁 config/              # ⚙️ Configuration management
│   │   │   └── 📄 config.go        # Load config.yaml
│   │   ├── 📁 container/           # 📦 Dependency injection container
│   │   ├── 📁 database/              # 💽 Database connections & migrations
│   │   │   ├── 📄 database.go      # Multi-database support (PostgreSQL, MySQL, SQL Server, MongoDB)
│   │   │   └── 📁 migrations/      # 🗄️ Database migrations
│   │   └── 📁 transport/           # 🌐 API Transport Layer
│   │       ├── 📁 grpc/            # 🚀 gRPC implementation
│   │       │   ├── 📁 gen/         # 📝 Generated gRPC code
│   │       │   ├── 📁 proto/       # 📄 Protocol buffer definitions
│   │       │   └── 📁 servers/     # 🖥️ gRPC server implementations
│   │       └── 📁 http/            # 🌐 REST API implementation
│   │           ├── 📁 handlers/      # 🎯 HTTP request handlers
│   │           ├── 📁 middleware/    # 🛡️ HTTP middlewares
│   │           ├── 📁 routes/        # 🗺️ Route definitions
│   │           └── 📁 servers/       # 🖥️ HTTP server setup
│   │ 
│   └── 📁 interfaces/              # 🔌 External Service Interfaces
│       ├── 📁 bpjs/                # 🏥 BPJS Integration
│       │   └── 📁 vclaim/          # 💳 BPJS VClaim service
│       │       ├── 📄 VCLAIM.md    # 📖 Documentation
│       │       ├── 📄 client.go    # 🌐 HTTP client
│       │       ├── 📄 crypto.go    # 🔐 Cryptography utilities
│       │       ├── 📄 service.go   # 🛠️ VClaim service logic
│       │       └── 📄 types.go     # 📋 Type definitions
│       └── 📁 satusehat/           # 🏥 SatuSehat Integration (HL7 FHIR)
│
├── 📁 pkg/                         # 📦 Public Reusable Libraries
│   ├── 📁 errors/                  # 🚨 Error handling system
│   │   ├── 📄 ERRORS.md            # 📖 Error documentation
│   │   ├── 📄 builder.go           # 🏗️ Error builder
│   │   ├── 📄 codes.go             # 🔢 Error codes
│   │   ├── 📄 grpc.go              # 🌐 gRPC error handling
│   │   ├── 📄 http.go              # 🌐 HTTP error handling
│   │   └── 📄 validator.go         # ✅ Validation errors
│   ├── 📁 logger/                  # 📝 Logging system
│   │   └── 📄 logger.go            # 🪵 Configurable logger
│   ├── 📁 metrics/                 # 📊 Application metrics
│   ├── 📁 response/                # 📤 HTTP response utilities
│   │   └── 📄 response.go          # 🎯 Standardized responses
│   └── 📁 utils/                   # 🛠️ Utility functions
│       └── 📁 query/               # 🔍 Advanced Query Builder
│           ├── 📄 BUILDER.md       # 📖 Query builder docs
│           ├── 📄 builder_query.go # 🏗️ Main query builder
│           ├── 📄 dialects.go      # 🗣️ Database dialect support
│           ├── 📄 sql_builder.go   # 🏗️ SQL query builder
│           ├── 📄 mongo_builder.go # 🍃 MongoDB query builder
│           └── 📄 sql_security.go  # 🔒 SQL injection protection
│
├── 📁 scripts/                     # 🔄 Automation Scripts
│   ├── 📄 build.sh               # 🔨 Build application
│   ├── 📄 generate_proto.sh      # 📝 Generate gRPC protobuf
│   ├── 📄 migrate.sh             # 🗄️ Run database migrations
│   └── 📄 seed.sh                # 🌱 Seed initial data
│
├── 📁 tools/                       # 🔧 Development Tools
│   ├── 📄 generate.go            # 🏭 Code generator
│   └── 📄 generate_config.yaml   # ⚙️ Generator configuration
│
├── 📄 config.yaml                # ⚙️ Main configuration file
├── 📄 go.mod                     # 📦 Go module dependencies
├── 📄 go.sum                     # 🔒 Dependency checksums
├── 📄 Makefile                   # 🔨 Build automation
├── 📄 README.md                  # 📖 Project documentation
├── 📄 Dockerfile                 # 🐳 Production container
├── 📄 Dockerfile.dev             # 🐳 Development container
├── 📄 docker-compose.dev.yml   # 🐳 Dev environment
├── 📄 docker-compose.prod.yml  # 🐳 Production environment
└── 📄 .air.toml                 # 🔄 Live reload configuration


goose -dir internal/infrastructure/database/sql create create_users_table sql