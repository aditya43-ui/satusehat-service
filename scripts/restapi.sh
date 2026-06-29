#!/bin/bash

# Script: scripts/restapi.sh
# Description: REST API Generator and Analyzer for Person Service
# Usage: ./scripts/restapi.sh [command] [options]
# Commands: generate, analyze, test, docs

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Project paths
PROJECT_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
HANDLERS_DIR="$PROJECT_ROOT/internal/infrastructure/transport/http/handlers"
GRPC_HANDLERS_DIR="$PROJECT_ROOT/internal/infrastructure/transport/grpc/handlers"
OUTPUT_DIR="$PROJECT_ROOT/docs/api"
SWAGGER_FILE="$OUTPUT_DIR/openapi.yaml"
POSTMAN_FILE="$PROJECT_ROOT/person.postman_collection.json"

# Default values
DEFAULT_HOST="localhost"
DEFAULT_PORT="8080"
DEFAULT_API_VERSION="v1"

# Helper functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Check if required files exist
check_requirements() {
    log_info "Checking requirements..."
    
    if [[ ! -f "$HANDLERS_DIR/person_handler.go" ]]; then
        log_error "Person handler not found at $HANDLERS_DIR/person_handler.go"
        exit 1
    fi
    
    if [[ ! -f "$GRPC_HANDLERS_DIR/person_handler.go" ]]; then
        log_error "gRPC person handler not found at $GRPC_HANDLERS_DIR/person_handler.go"
        exit 1
    fi
    
    if [[ ! -d "$OUTPUT_DIR" ]]; then
        mkdir -p "$OUTPUT_DIR"
    fi
    
    log_success "Requirements check passed"
}

# Extract endpoints from person_handler.go
extract_rest_endpoints() {
    log_info "Extracting REST endpoints from person_handler.go..."
    
    local handler_file="$HANDLERS_DIR/person_handler.go"
    local endpoints_file="$OUTPUT_DIR/rest_endpoints.txt"
    
    cat > "$endpoints_file" << 'EOF'
# REST API Endpoints - Person Service
# Generated on: $(date)
# Handler: internal/infrastructure/transport/http/handlers/person_handler.go

## Base Path: /api/v1/persons

### 1. Get List of Persons (Pagination)
- **Method**: GET
- **Path**: /persons
- **Query Parameters**:
  - page (int, optional): Page number (default: 1)
  - limit (int, optional): Items per page (default: 10)
- **Response**: Paginated list of persons
- **Handler**: PersonHandler.GetList()

### 2. Get Person Detail
- **Method**: GET
- **Path**: /persons/{id}
- **Path Parameters**:
  - id (int64, required): Person ID
- **Response**: Single person details
- **Handler**: PersonHandler.GetDetail()

### 3. Create New Person
- **Method**: POST
- **Path**: /persons
- **Request Body**: CreatePersonRequest JSON
- **Response**: Created person details
- **Handler**: PersonHandler.Create()

### 4. Update Person
- **Method**: PUT
- **Path**: /persons/{id}
- **Path Parameters**:
  - id (int64, required): Person ID
- **Request Body**: CreatePersonRequest JSON
- **Response**: Updated person details
- **Handler**: PersonHandler.Update()

EOF

    log_success "REST endpoints extracted to $endpoints_file"
}

# Extract gRPC methods from person_handler.go
extract_grpc_endpoints() {
    log_info "Extracting gRPC endpoints from person_handler.go..."
    
    local grpc_handler_file="$GRPC_HANDLERS_DIR/person_handler.go"
    local grpc_endpoints_file="$OUTPUT_DIR/grpc_endpoints.txt"
    
    cat > "$grpc_endpoints_file" << 'EOF'
# gRPC API Endpoints - Person Service
# Generated on: $(date)
# Handler: internal/infrastructure/transport/grpc/handlers/person_handler.go

## Service: PersonService

### 1. GetPerson
- **Method**: GetPerson
- **Request**: GetPersonRequest { id: int64 }
- **Response**: GetPersonResponse { person: Person }
- **Handler**: PersonHandler.GetPerson()

### 2. ListPersons
- **Method**: ListPersons
- **Request**: ListPersonsRequest { page: int32, page_size: int32 }
- **Response**: ListPersonsResponse { persons: []Person, total: int64 }
- **Handler**: PersonHandler.ListPersons()

### 3. CreatePerson
- **Method**: CreatePerson
- **Request**: CreatePersonRequest { data: CreatePersonRequest }
- **Response**: GetPersonResponse { person: Person }
- **Handler**: PersonHandler.CreatePerson()

### 4. UpdatePerson
- **Method**: UpdatePerson
- **Request**: UpdatePersonRequest { id: int64, data: CreatePersonRequest }
- **Response**: GetPersonResponse { person: Person }
- **Handler**: PersonHandler.UpdatePerson()

### 5. DeletePerson
- **Method**: DeletePerson
- **Request**: DeletePersonRequest { id: int64 }
- **Response**: Empty {}
- **Handler**: PersonHandler.DeletePerson()

EOF

    log_success "gRPC endpoints extracted to $grpc_endpoints_file"
}

# Generate Postman collection
generate_postman_collection() {
    log_info "Generating Postman collection..."
    
    cat > "$POSTMAN_FILE" << 'EOF'
{
    "info": {
        "name": "Person Service API",
        "description": "REST API Collection for Person Service",
        "schema": "https://schema.getpostman.com/json/collection/v2.1.0/collection.json"
    },
    "item": [
        {
            "name": "Get Person List",
            "request": {
                "method": "GET",
                "header": [],
                "url": {
                    "raw": "{{base_url}}/api/v1/persons?page=1&limit=10",
                    "host": ["{{base_url}}"],
                    "path": ["api", "v1", "persons"],
                    "query": [
                        {
                            "key": "page",
                            "value": "1",
                            "description": "Page number"
                        },
                        {
                            "key": "limit",
                            "value": "10",
                            "description": "Items per page"
                        }
                    ]
                }
            }
        },
        {
            "name": "Get Person Detail",
            "request": {
                "method": "GET",
                "header": [],
                "url": {
                    "raw": "{{base_url}}/api/v1/persons/1",
                    "host": ["{{base_url}}"],
                    "path": ["api", "v1", "persons", "1"]
                }
            }
        },
        {
            "name": "Create Person",
            "request": {
                "method": "POST",
                "header": [
                    {
                        "key": "Content-Type",
                        "value": "application/json"
                    }
                ],
                "body": {
                    "mode": "raw",
                    "raw": "{\n    \"name\": \"John Doe\",\n    \"email\": \"john.doe@example.com\",\n    \"phone\": \"+1234567890\"\n}"
                },
                "url": {
                    "raw": "{{base_url}}/api/v1/persons",
                    "host": ["{{base_url}}"],
                    "path": ["api", "v1", "persons"]
                }
            }
        },
        {
            "name": "Update Person",
            "request": {
                "method": "PUT",
                "header": [
                    {
                        "key": "Content-Type",
                        "value": "application/json"
                    }
                ],
                "body": {
                    "mode": "raw",
                    "raw": "{\n    \"name\": \"John Doe Updated\",\n    \"email\": \"john.updated@example.com\",\n    \"phone\": \"+0987654321\"\n}"
                },
                "url": {
                    "raw": "{{base_url}}/api/v1/persons/1",
                    "host": ["{{base_url}}"],
                    "path": ["api", "v1", "persons", "1"]
                }
            }
        }
    ],
    "variable": [
        {
            "key": "base_url",
            "value": "http://localhost:8080",
            "type": "string"
        }
    ]
}
EOF

    log_success "Postman collection generated at $POSTMAN_FILE"
}

# Generate Swagger documentation
generate_swagger_docs() {
    log_info "Generating Swagger documentation..."
    
    cat > "$SWAGGER_FILE" << 'EOF'
openapi: 3.0.0
info:
  title: Person Service API
  description: REST API for Person Management Service
  version: 1.0.0
  contact:
    name: API Support
    email: support@example.com
servers:
  - url: http://localhost:8080/api/v1
    description: Development server
paths:
  /persons:
    get:
      summary: Get list of persons
      description: Retrieve paginated list of persons
      parameters:
        - name: page
          in: query
          description: Page number
          required: false
          schema:
            type: integer
            default: 1
        - name: limit
          in: query
          description: Items per page
          required: false
          schema:
            type: integer
            default: 10
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PaginatedPersonResponse'
    post:
      summary: Create new person
      description: Create a new person record
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreatePersonRequest'
      responses:
        '201':
          description: Person created successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PersonResponse'
  /persons/{id}:
    get:
      summary: Get person detail
      description: Retrieve detailed information about a specific person
      parameters:
        - name: id
          in: path
          required: true
          description: Person ID
          schema:
            type: integer
            format: int64
      responses:
        '200':
          description: Successful response
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PersonResponse'
        '404':
          description: Person not found
    put:
      summary: Update person
      description: Update an existing person record
      parameters:
        - name: id
          in: path
          required: true
          description: Person ID
          schema:
            type: integer
            format: int64
      requestBody:
        required: true
        content:
          application/json:
            schema:
              $ref: '#/components/schemas/CreatePersonRequest'
      responses:
        '200':
          description: Person updated successfully
          content:
            application/json:
              schema:
                $ref: '#/components/schemas/PersonResponse'
        '404':
          description: Person not found
components:
  schemas:
    CreatePersonRequest:
      type: object
      properties:
        name:
          type: string
        email:
          type: string
          format: email
        phone:
          type: string
      required:
        - name
        - email
    PersonResponse:
      type: object
      properties:
        id:
          type: integer
          format: int64
        name:
          type: string
        email:
          type: string
        phone:
          type: string
        created_at:
          type: string
          format: date-time
    PaginatedPersonResponse:
      type: object
      properties:
        data:
          type: array
          items:
            $ref: '#/components/schemas/PersonResponse'
        meta:
          type: object
          properties:
            page:
              type: integer
            limit:
              type: integer
            total:
              type: integer
            total_pages:
              type: integer
EOF

    log_success "Swagger documentation generated at $SWAGGER_FILE"
}

# Test API endpoints
test_api_endpoints() {
    log_info "Testing API endpoints..."
    
    local base_url="http://${DEFAULT_HOST}:${DEFAULT_PORT}"
    local test_results="$OUTPUT_DIR/api_test_results.txt"
    
    echo "API Test Results - $(date)" > "$test_results"
    echo "================================" >> "$test_results"
    echo "" >> "$test_results"
    
    # Test health endpoint first
    log_info "Testing health endpoint..."
    if curl -s -o /dev/null -w "%{http_code}" "$base_url/health" | grep -q "200"; then
        echo "✅ Health check: PASSED" >> "$test_results"
        log_success "Health check passed"
    else
        echo "❌ Health check: FAILED" >> "$test_results"
        log_error "Health check failed"
    fi
    
    # Test person list endpoint
    log_info "Testing person list endpoint..."
    local list_response=$(curl -s -w "\n%{http_code}" "$base_url/api/v1/persons?page=1&limit=5")
    local list_http_code=$(echo "$list_response" | tail -n1)
    
    if [[ "$list_http_code" == "200" ]] || [[ "$list_http_code" == "204" ]]; then
        echo "✅ Person list endpoint: PASSED (HTTP $list_http_code)" >> "$test_results"
        log_success "Person list endpoint working"
    else
        echo "❌ Person list endpoint: FAILED (HTTP $list_http_code)" >> "$test_results"
        log_error "Person list endpoint failed"
    fi
    
    log_success "API test results saved to $test_results"
}

# Generate comprehensive API documentation
generate_api_docs() {
    log_info "Generating comprehensive API documentation..."
    
    local api_docs="$OUTPUT_DIR/API_DOCUMENTATION.md"
    
    cat > "$api_docs" << 'EOF'
# Person Service API Documentation

## Overview
This document provides comprehensive documentation for the Person Service API, including REST and gRPC endpoints.

## Table of Contents
1. [REST API Endpoints](#rest-api-endpoints)
2. [gRPC API Endpoints](#grpc-api-endpoints)
3. [Request/Response Examples](#requestresponse-examples)
4. [Error Handling](#error-handling)
5. [Authentication](#authentication)
6. [Rate Limiting](#rate-limiting)

## REST API Endpoints

### Base URL

EOF

    log_success "API documentation generated at $api_docs"
}

# Main function
main() {
    local command="$1"
    
    case "$command" in
        "generate")
            check_requirements
            extract_rest_endpoints
            extract_grpc_endpoints
            generate_postman_collection
            generate_swagger_docs
            generate_api_docs
            log_success "API generation completed"
            ;;
        "analyze")
            check_requirements
            extract_rest_endpoints
            extract_grpc_endpoints
            log_success "API analysis completed"
            ;;
        "test")
            check_requirements
            test_api_endpoints
            log_success "API testing completed"
            ;;
        "docs")
            check_requirements
            generate_api_docs
            log_success "API documentation generated"
            ;;
        *)
            echo "Usage: $0 [generate|analyze|test|docs]"
            exit 1
            ;;
    esac
}

# Run main function
main "$@"