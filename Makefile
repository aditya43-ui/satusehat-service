.PHONY: dev prod clean

dev:
	docker-compose -f docker-compose.dev.yml up --build

logs:
	docker-compose -f docker-compose.dev.yml logs -f

prod:
	docker-compose -f docker-compose.prod.yml up --build

clean:
	docker-compose -f docker-compose.dev.yml down -v
	docker-compose -f docker-compose.prod.yml down -v

security-check:
	@echo "🔍 Running Static Application Security Testing (SAST)..."
	gosec -quiet ./...

audit:
	@echo "🔍 Checking for known vulnerabilities in dependencies..."
	govulncheck ./...
help:
	@echo "Available targets:"
	@echo "  proto          - Generate all protobuf files"
	@echo "  proto-all      - Generate all protobuf files with cleanup"
	@echo "  generate-proto - Generate specific proto file (usage: make generate-proto FILE=path/to/file.proto)"
	@echo "  generate-ethnic-grpc  - Generate ethnic gRPC service"
	@echo "  generate-person-grpc  - Generate person gRPC service"
	@echo "  clean          - Clean generated files"
	@echo "  build          - Build the application"
	@echo "  run            - Run the application"
	@echo "  test           - Run tests"
	@echo "  test-coverage  - Run tests and generate coverage report"
	@echo "  govulncheck    - Check for vulnerabilities"
	@echo "  seeder-build   - Build the seeder CLI tool"
	@echo "  seeder-list    - List all available tables for seeding"
	@echo "  seeder-seed    - Seed specific table (usage: make seeder-seed TABLE=province)"
	@echo "  seeder-seed-all - Seed all tables"
	@echo "  seeder-dry-run - Dry run seeding for table (usage: make seeder-dry-run TABLE=province)"
	@echo "  seeder-validate - Validate CSV files and configuration"

# Generate specific proto file dengan penanganan error
generate-proto:
	@if [ -z "$(FILE)" ]; then \
		echo "❌ Usage: make generate-proto FILE=path/to/file.proto"; \
		echo "Example: make generate-proto FILE=internal/infrastructure/transport/grpc/proto/v1/ethnic.proto"; \
		exit 1; \
	fi
	@echo "🚀 Generating protobuf code untuk: $(FILE)"
	@if command -v protoc-gen-validate &> /dev/null; then \
		echo "✅ protoc-gen-validate ditemukan, generate dengan validasi..."; \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			--validate_out="lang=go:." \
			$(FILE); \
	else \
		echo "⚠️  protoc-gen-validate tidak ditemukan, generate tanpa validasi..."; \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			$(FILE); \
	fi

# Seeder targets
seeder-build:
	@echo "🔨 Building seeder CLI tool..."
	@go build -o bin/seeder cmd/seeder/main.go
	@echo "✅ Seeder CLI built successfully at bin/seeder"

seeder-list: seeder-build
	@echo "📋 Listing all available tables for seeding..."
	@./bin/seeder list

seeder-seed: seeder-build
	@if [ -z "$(TABLE)" ]; then \
		echo "❌ Usage: make seeder-seed TABLE=<table-name>"; \
		echo "Example: make seeder-seed TABLE=province"; \
		echo "Available tables:"; \
		./bin/seeder list; \
		exit 1; \
	fi
	@echo "🌱 Seeding table: $(TABLE)"
	@./bin/seeder seed $(TABLE)

seeder-seed-all: seeder-build
	@echo "🌱 Seeding all tables..."
	@./bin/seeder seed all

seeder-dry-run: seeder-build
	@if [ -z "$(TABLE)" ]; then \
		echo "❌ Usage: make seeder-dry-run TABLE=<table-name>"; \
		echo "Example: make seeder-dry-run TABLE=province"; \
		exit 1; \
	fi
	@echo "🔍 Dry-run seeding for table: $(TABLE)"
	@./bin/seeder dry-run $(TABLE)

seeder-validate: seeder-build
	@echo "✅ Validating all tables and CSV files..."
	@./bin/seeder validate all

# Advanced seeder with options
seeder-advanced: seeder-build
	@if [ -z "$(TABLE)" ]; then \
		echo "❌ Usage: make seeder-advanced TABLE=<table-name> [OPTIONS]"; \
		echo "Options:"; \
		echo "  BATCH_SIZE=<n>     - Set batch size"; \
		echo "  DELETE_BEFORE=1    - Delete existing data before seeding"; \
		echo "  CSV_PATH=<path>    - Override CSV file path"; \
		exit 1; \
	fi
	@echo "🌱 Advanced seeding for table: $(TABLE)"
	@./bin/seeder seed $(TABLE) -batch-size=$(or $(BATCH_SIZE),100) $(if $(DELETE_BEFORE),-delete-before,) $(if $(CSV_PATH),-csv-path=$(CSV_PATH),)

clean-seeder:
	@echo "🧹 Cleaning seeder binary..."
	@rm -f bin/seeder
	@echo "✅ Seeder binary cleaned"
	@echo "✅ Proto code berhasil digenerate"

# Generate all protobuf files
proto:
	@echo "Generating all protobuf files..."
	@for proto in internal/infrastructure/transport/grpc/proto/v1/*.proto; do \
		echo "Processing $$proto..."; \
		if command -v protoc-gen-validate &> /dev/null; then \
			protoc --go_out=. --go_opt=paths=source_relative \
				--go-grpc_out=. --go-grpc_opt=paths=source_relative \
				--validate_out="lang=go:." \
				$$proto; \
		else \
			protoc --go_out=. --go_opt=paths=source_relative \
				--go-grpc_out=. --go-grpc_opt=paths=source_relative \
				$$proto; \
		fi; \
	done
	@echo "✅ All protobuf files generated"

# Generate all protobuf files with cleanup
proto-all: clean-gen proto

# Generate specific service (will be added by generator script)
generate-ethnic-grpc:
	@echo "Generating ethnic gRPC service..."
	@if [ -f "internal/infrastructure/transport/grpc/proto/v1/ethnic.proto" ]; then \
		if command -v protoc-gen-validate &> /dev/null; then \
			protoc --go_out=. --go_opt=paths=source_relative \
				--go-grpc_out=. --go-grpc_opt=paths=source_relative \
				--validate_out="lang=go:." \
				internal/infrastructure/transport/grpc/proto/v1/ethnic.proto; \
		else \
			protoc --go_out=. --go_opt=paths=source_relative \
				--go-grpc_out=. --go-grpc_opt=paths=source_relative \
				internal/infrastructure/transport/grpc/proto/v1/ethnic.proto; \
		fi; \
	else \
		echo "⚠️  File ethnic.proto tidak ditemukan, skip..."; \
	fi
	@echo "✅ Ethnic gRPC service generated"

generate-person-grpc:
	@echo "Generating person gRPC service..."
	@if command -v protoc-gen-validate &> /dev/null; then \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			--validate_out="lang=go:." \
			internal/infrastructure/transport/grpc/proto/v1/person.proto; \
	else \
		protoc --go_out=. --go_opt=paths=source_relative \
			--go-grpc_out=. --go-grpc_opt=paths=source_relative \
			internal/infrastructure/transport/grpc/proto/v1/person.proto; \
	fi
	@echo "✅ Person gRPC service generated"

# Clean generated files
clean-gen:
	@echo "Cleaning generated files..."
	@rm -rf internal/infrastructure/transport/grpc/gen/
	@mkdir -p internal/infrastructure/transport/grpc/gen
	@echo "✅ Generated files cleaned"
# Generate village gRPC service
generate-village-grpc:
	@echo "Generating village gRPC service..."
	@protoc --go_out=. --go_opt=paths=source_relative \
		--go-grpc_out=. --go-grpc_opt=paths=source_relative \
		--validate_out="lang=go:." \
		--experimental_allow_proto3_optional \
		internal/infrastructure/transport/grpc/proto/v1/village.proto
	@echo "✅ village gRPC service generated"

.PHONY: test test-coverage

# Run tests
test:
	@echo "🧪 Running tests..."
	@go test -v ./...

# Run tests with coverage
test-coverage:
	@echo "🧪 Running tests with coverage..."
	@go test -v -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html
	@echo "✅ Coverage report generated at coverage.html"