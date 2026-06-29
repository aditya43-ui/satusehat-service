#!/bin/bash

# gRPC Layer Generator from Existing Context
# Usage: ./scripts/grpc.sh [OPTIONS]
# Options:
#   -d, --dir PATH         Custom directory structure of the existing context (e.g., master/reference/province) (required)
#   -e, --entity-file PATH Optional path to the entity file if not named 'entity.go'
#   -v, --verbose          Verbose output
#   -h, --help             Show help

set -uo pipefail

# Colors for output
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'
readonly WHITE='\033[1;37m'
readonly NC='\033[0m' # No Color

# Configuration
readonly SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
readonly PROJECT_ROOT="$(cd "${SCRIPT_DIR}/.." && pwd)"
readonly INTERNAL_DIR="${PROJECT_ROOT}/internal"

# Global variables
VERBOSE=false
CUSTOM_DIR=""
ENTITY_FILE_NAME="entity.go"

# --- Helper Functions ---

# Logging functions
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_verbose() { [[ "$VERBOSE" == true ]] && echo -e "${CYAN}[VERBOSE]${NC} $1"; }
log_header() { echo -e "\n${PURPLE}==== $1 ====${NC}\n"; }

# Help function
show_help() {
    cat << EOF
 ${WHITE}gRPC Layer Generator from Existing Context${NC}

 This script generates the gRPC layer (.proto, handler, mapper)
 by parsing an existing Go entity file.

 ${YELLOW}Usage:${NC}
    $(basename "$0") -d CONTEXT_DIR [OPTIONS]

 ${YELLOW}Required Options:${NC}
    -d, --dir PATH         Path to the existing context directory relative to 'internal/'
                           (e.g., master/reference/province)

 ${YELLOW}Optional Options:${NC}
    -e, --entity-file NAME File name of the source entity (default: entity.go)
    -v, --verbose          Verbose output
    -h, --help             Show this help message

 ${YELLOW}Example:${NC}
    # Generate gRPC layer for an existing 'province' context
    $(basename "$0") -d master/reference/province
EOF
}

# PascalCase from snake_case or kebab-case
to_pascal_case() {
    echo "$1" | sed -E 's/(^|[-_.])([a-zA-Z])/\U\2/g'
}

# snake_case from PascalCase
to_snake_case() {
    echo "$1" | sed -E 's/([A-Z])/_\L\1/g' | sed 's/^_//'
}

# camelCase from snake_case or kebab-case
to_camel_case() {
    local pascal
    pascal=$(to_pascal_case "$1")
    echo "$(echo "${pascal:0:1}" | tr '[:upper:]' '[:lower:]')${pascal:1}"
}

# Extract package name from custom directory structure
# For path master/reference/province, returns "province"
extract_package_name() {
    basename "$1"
}

# Convert Go type to Protobuf type
go_to_proto_type() {
    local go_type="$1"
    local is_pointer="${2:-false}"

    local proto_type="string" # Default
    case "$go_type" in
        "int"|"int32"|"int64") proto_type="int64" ;;
        "string") proto_type="string" ;;
        "bool") proto_type="bool" ;;
        "time.Time") proto_type="google.protobuf.Timestamp" ;;
        "float32"|"float64") proto_type="double" ;;
        "uuid.UUID") proto_type="string" ;;
    esac

    # Use optional for pointer fields (nullable) in proto3
    if [[ "$is_pointer" == "true" ]]; then
        echo "optional $proto_type"
    else
        echo "$proto_type"
    fi
}

# --- Core Logic ---

# Parse command line arguments
parse_args() {
    if [[ "$#" -eq 0 ]]; then
        show_help
        exit 1
    fi

    while [[ $# -gt 0 ]]; do
        case $1 in
            -d|--dir)
                CUSTOM_DIR="$2"
                shift 2
                ;;
            -e|--entity-file)
                ENTITY_FILE_NAME="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE=true
                shift
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done

    # Validate required arguments
    if [[ -z "$CUSTOM_DIR" ]]; then
        log_error "Context directory is required. Use -d or --dir."
        exit 1
    fi
}

# Extract struct info from Go entity file
parse_go_context() {
    local context_path="$1"
    local entity_filename="$2"
    local entity_file_path="${INTERNAL_DIR}/${context_path}/${entity_filename}"

    log_info "Parsing Go context from: $entity_file_path"

    if [[ ! -f "$entity_file_path" ]]; then
        log_error "Entity file not found: $entity_file_path"
        return 1
    fi

    # Extract struct name (e.g., Province)
    local struct_name
    struct_name=$(grep -m 1 -E "type .* struct" "$entity_file_path" | awk '{print $2}')
    if [[ -z "$struct_name" ]]; then
        log_error "Could not find a struct definition in $entity_file_path"
        return 1
    fi
    log_verbose "Found struct: $struct_name"

    # Reset arrays
    PARSED_FIELDS=()
    PARSED_PRIMARY_KEY_GO_NAME=""
    PARSED_PRIMARY_KEY_GO_TYPE=""

    # Use awk to isolate the struct definition block
    local fields_block
    fields_block=$(awk -v struct_name="$struct_name" '$0 ~ "type " struct_name " struct \\{" {p=1; next} p && /}/ {p=0} p' "$entity_file_path")

    while IFS= read -r line; do
        # Trim leading/trailing whitespace
        local trimmed_line
        trimmed_line=$(echo "$line" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')

        # Skip empty, comment, or closing brace lines
        if [[ -z "$trimmed_line" || "$trimmed_line" =~ ^// || "$trimmed_line" == "}" ]]; then
            continue
        fi

        local go_field_name go_type is_pointer db_name
        go_field_name=$(echo "$trimmed_line" | awk '{print $1}')
        go_type=$(echo "$trimmed_line" | awk '{print $2}')
        is_pointer="false"

        if [[ "$go_type" == "*"* ]]; then
            is_pointer="true"
            go_type=${go_type#\*} # Remove the leading '*'
        fi

        # Extract db tag for snake_case name, fallback to json tag, then to converting field name
        db_name=$(echo "$trimmed_line" | grep -o 'db:"[^"]*"' | cut -d'"' -f2)
        if [[ -z "$db_name" ]]; then
            db_name=$(echo "$trimmed_line" | grep -o 'json:"[^"]*"' | cut -d'"' -f2)
        fi
        if [[ -z "$db_name" ]]; then
            db_name=$(to_snake_case "$go_field_name")
        fi

        PARSED_FIELDS+=("${go_field_name}|${go_type}|${is_pointer}|${db_name}")
        log_verbose "Found field: $go_field_name ($go_type, pointer: $is_pointer, db_name: $db_name)"

        # Heuristic to find the Primary Key
        if [[ -z "$PARSED_PRIMARY_KEY_GO_NAME" && ("$go_field_name" == "Id" || "$db_name" == "id") ]]; then
            PARSED_PRIMARY_KEY_GO_NAME="$go_field_name"
            PARSED_PRIMARY_KEY_GO_TYPE="$go_type"
        fi

    done <<< "$fields_block"

    # If no PK found by heuristic, assume the first field is the PK
    if [[ -z "$PARSED_PRIMARY_KEY_GO_NAME" && ${#PARSED_FIELDS[@]} -gt 0 ]]; then
        IFS='|' read -r go_field_name go_type _ _ <<< "${PARSED_FIELDS[0]}"
        PARSED_PRIMARY_KEY_GO_NAME="$go_field_name"
        PARSED_PRIMARY_KEY_GO_TYPE="$go_type"
        log_warning "Could not determine primary key, assuming first field '${go_field_name}' is the PK."
    fi

    # Store in global variables for later use
    export CLEAN_NAME="$struct_name"
    export GO_PK_NAME="$PARSED_PRIMARY_KEY_GO_NAME"
    export GO_PK_TYPE="$PARSED_PRIMARY_KEY_GO_TYPE"
    export PROTO_PK_TYPE=$(go_to_proto_type "$PARSED_PRIMARY_KEY_GO_TYPE" "false")

    log_success "Successfully parsed context for: $struct_name"
}

# Generate gRPC proto file
generate_proto_file() {
    local package_name
    package_name=$(extract_package_name "$CUSTOM_DIR")

    local proto_dir="${INTERNAL_DIR}/infrastructure/transport/grpc/proto/${CUSTOM_DIR}/v1"
    log_info "Generating proto file in: $proto_dir"
    mkdir -p "$proto_dir"

    local proto_file="${proto_dir}/${package_name}.proto"

    if [ -f "$proto_file" ]; then
        log_warning "File already exists, skipping: $proto_file"
        return
    fi

    log_info "Generating file: $proto_file"
    cat > "$proto_file" << EOF
syntax = "proto3";

package ${package_name}.v1;

option go_package = "service/internal/infrastructure/transport/grpc/gen/${CUSTOM_DIR}/v1;${package_name}v1";

import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";

// Service definition for ${CLEAN_NAME}.
service ${CLEAN_NAME}Service {
  rpc Get${CLEAN_NAME}(Get${CLEAN_NAME}Request) returns (${CLEAN_NAME}Response);
  rpc List${CLEAN_NAME}s(List${CLEAN_NAME}sRequest) returns (List${CLEAN_NAME}sResponse);
  rpc Create${CLEAN_NAME}(Create${CLEAN_NAME}Request) returns (${CLEAN_NAME}Response);
  rpc Update${CLEAN_NAME}(Update${CLEAN_NAME}Request) returns (${CLEAN_NAME}Response);
  rpc Delete${CLEAN_NAME}(Delete${CLEAN_NAME}Request) returns (google.protobuf.Empty);
}

// The main message representing a ${CLEAN_NAME}.
message ${CLEAN_NAME} {
EOF

    local field_index=1
    for field in "${PARSED_FIELDS[@]}"; do
        IFS='|' read -r go_field_name go_type is_pointer db_name <<< "$field"
        local proto_field_name="$db_name"
        local proto_type
        proto_type=$(go_to_proto_type "$go_type" "$is_pointer")

        echo "  ${proto_type} ${proto_field_name} = ${field_index};" >> "$proto_file"
        ((field_index++))
    done

    cat >> "$proto_file" << EOF
}

// --- Request/Response Messages ---

message Get${CLEAN_NAME}Request {
  ${PROTO_PK_TYPE} id = 1;
}

message ${CLEAN_NAME}Response {
  ${CLEAN_NAME} data = 1;
}

message List${CLEAN_NAME}sRequest {
  int32 page = 1;
  int32 page_size = 2;
  // TODO: Add filter fields here if needed
}

message List${CLEAN_NAME}sResponse {
  repeated ${CLEAN_NAME} data = 1;
  int64 total = 2;
}

message Create${CLEAN_NAME}Request {
EOF

    local create_field_index=1
    for field in "${PARSED_FIELDS[@]}"; do
        IFS='|' read -r go_field_name go_type is_pointer db_name <<< "$field"
        # Skip PK, created_at, updated_at, deleted_at for create requests
        if [[ "$db_name" == "id" || "$db_name" == "created_at" || "$db_name" == "updated_at" || "$db_name" == "deleted_at" ]]; then
            continue
        fi
        local proto_field_name="$db_name"
        local proto_type
        proto_type=$(go_to_proto_type "$go_type" "false") # All fields are required for create

        echo "  ${proto_type} ${proto_field_name} = ${create_field_index};" >> "$proto_file"
        ((create_field_index++))
    done

    cat >> "$proto_file" << EOF
}

message Update${CLEAN_NAME}Request {
  ${PROTO_PK_TYPE} id = 1;
EOF

    local update_field_index=2
    for field in "${PARSED_FIELDS[@]}"; do
        IFS='|' read -r go_field_name go_type is_pointer db_name <<< "$field"
        if [[ "$db_name" == "id" || "$db_name" == "created_at" || "$db_name" == "updated_at" || "$db_name" == "deleted_at" ]]; then
            continue
        fi
        local proto_field_name="$db_name"
        # For update, all fields are optional
        local proto_type
        proto_type=$(go_to_proto_type "$go_type" "true")

        echo "  ${proto_type} ${proto_field_name} = ${update_field_index};" >> "$proto_file"
        ((update_field_index++))
    done

    cat >> "$proto_file" << EOF
}

message Delete${CLEAN_NAME}Request {
  ${PROTO_PK_TYPE} id = 1;
}
EOF

    log_success "Generated proto file: ${package_name}.proto"

    log_info "Running proto compiler to generate Go code from this new .proto file..."
    # Assuming you have a script to compile protos, or you can do it directly.
    # This is an example command.
    if command -v protoc &> /dev/null; then
        protoc --go_out=. --go_opt=paths=source_relative \
               --go-grpc_out=. --go-grpc_opt=paths=source_relative \
               "${proto_dir}/${package_name}.proto"
        log_success "protoc compilation successful."
    else
        log_warning "protoc command not found. Please compile the .proto file manually."
    fi
}

# Generate gRPC handler and mapper file
generate_grpc_handler() {
    local package_name
    package_name=$(extract_package_name "$CUSTOM_DIR")

    local target_dir="${INTERNAL_DIR}/infrastructure/transport/grpc/handlers/${CUSTOM_DIR}"
    log_info "Generating gRPC handler and mapper in: $target_dir"
    mkdir -p "$target_dir"

    local handler_file_name="${target_dir}/${package_name}_grpc_handler.go"
    local mapper_file_name="${target_dir}/${package_name}_grpc_mapper.go"

    if [ -f "$handler_file_name" ]; then
        log_warning "File already exists, skipping: $handler_file_name"
    else
        log_info "Generating file: $handler_file_name"
        cat > "$handler_file_name" << EOF
package handlers

import (
	"context"

	"service/internal/infrastructure/transport/grpc/gen/${CUSTOM_DIR}/v1"
	${package_name}Service "service/internal/${CUSTOM_DIR}"
	"service/pkg/errors"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/types/known/emptypb"
)

type ${CLEAN_NAME}GrpcHandler struct {
	${package_name}v1.Unimplemented${CLEAN_NAME}ServiceServer
	service ${package_name}Service.Service
}

func New${CLEAN_NAME}GrpcHandler(service ${package_name}Service.Service) *${CLEAN_NAME}GrpcHandler {
	return &${CLEAN_NAME}GrpcHandler{service: service}
}

func (h *${CLEAN_NAME}GrpcHandler) Get${CLEAN_NAME}(ctx context.Context, req *${package_name}v1.Get${CLEAN_NAME}Request) (*${package_name}v1.${CLEAN_NAME}Response, error) {
	res, err := h.service.GetDetail(ctx, req.GetId())
	if err != nil {
		appErr := errors.FromError(err)
		return nil, status.Error(appErr.GRPCStatus(), appErr.Error())
	}
	return &${package_name}v1.${CLEAN_NAME}Response{Data: mapResponseToProto(res)}, nil
}

// Add other handler methods (List, Create, Update, Delete) here...

EOF
        log_success "Generated gRPC handler file: ${package_name}_grpc_handler.go"
    fi

    if [ -f "$mapper_file_name" ]; then
        log_warning "File already exists, skipping: $mapper_file_name"
    else
        log_info "Generating file: $mapper_file_name"
        cat > "$mapper_file_name" << EOF
package handlers

import (
	"service/internal/infrastructure/transport/grpc/gen/${CUSTOM_DIR}/v1"
	${package_name}Service "service/internal/${CUSTOM_DIR}"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// mapResponseToProto converts the service DTO to a Protobuf message.
func mapResponseToProto(dto *${package_name}Service.${CLEAN_NAME}Response) *${package_name}v1.${CLEAN_NAME} {
	if dto == nil {
		return nil
	}
	return &${package_name}v1.${CLEAN_NAME}{
EOF
        for field in "${PARSED_FIELDS[@]}"; do
            IFS='|' read -r go_field_name go_type is_pointer db_name <<< "$field"
            local proto_field_name="$db_name"
            local go_pascal_field
            go_pascal_field=$(to_pascal_case "$go_field_name")
            
            # Handle time.Time and *time.Time specifically
            if [[ "$go_type" == "time.Time" ]]; then
                if [[ "$is_pointer" == "true" ]]; then
                    echo "		${proto_field_name}: timestamppb.New(*dto.${go_pascal_field})," >> "$mapper_file_name"
                else
                    echo "		${proto_field_name}: timestamppb.New(dto.${go_pascal_field})," >> "$mapper_file_name"
                fi
            else
                 echo "		${proto_field_name}: dto.${go_pascal_field}," >> "$mapper_file_name"
            fi
        done
        cat >> "$mapper_file_name" << EOF
	}
}

// Add other mappers (e.g., mapCreateProtoToRequest) here...
EOF
        log_success "Generated gRPC mapper file: ${package_name}_grpc_mapper.go"
    fi
}

# --- Main Execution ---

main() {
    log_header "gRPC Layer Generator"
    parse_args "$@"

    if ! parse_go_context "$CUSTOM_DIR" "$ENTITY_FILE_NAME"; then
        exit 1
    fi

    generate_proto_file
    generate_grpc_handler

    local package_name
    package_name=$(extract_package_name "$CUSTOM_DIR")

    echo -e "\n${PURPLE}================================================================${NC}"
    echo -e "${GREEN}✨ gRPC LAYER FOR [${CLEAN_NAME}] GENERATED SUCCESSFULLY! ✨${NC}"
    echo -e "${PURPLE}================================================================${NC}\n"

    echo -e "${YELLOW}🚀 NEXT STEPS TO ACTIVATE YOUR gRPC MODULE:${NC}\n"

    echo -e "${CYAN}1. Review Generated Files:${NC}"
    echo -e "   - ${WHITE}internal/infrastructure/transport/grpc/proto/${CUSTOM_DIR}/v1/${package_name}.proto${NC}"
    echo -e "   - ${WHITE}internal/infrastructure/transport/grpc/handlers/${CUSTOM_DIR}/${package_name}_grpc_handler.go${NC}"
    echo -e "   - ${WHITE}internal/infrastructure/transport/grpc/handlers/${CUSTOM_DIR}/${package_name}_grpc_mapper.go${NC}"
    echo -e "   Lengkapi implementasi untuk method List, Create, Update, Delete di handler dan mapper.\n"

    echo -e "${CYAN}2. Registrasi gRPC Handler (misal di cmd/grpc/server.go):${NC}"
    echo -e "   Import packages:"
    echo -e "   ${WHITE}gen${CLEAN_NAME} \"service/internal/infrastructure/transport/grpc/gen/${CUSTOM_DIR}/v1\"${NC}"
    echo -e "   ${WHITE}grpc${CLEAN_NAME}Handler \"service/internal/infrastructure/transport/grpc/handlers/${CUSTOM_DIR}\"${NC}\n"
    echo -e "   Inisialisasi dan registrasi handler:"
    echo -e "   ${WHITE}// (Asumsikan ${package_name}Svc sudah diinisialisasi)${NC}"
    echo -e "   ${WHITE}${package_name}GrpcHandler := grpc${CLEAN_NAME}Handler.New${CLEAN_NAME}GrpcHandler(${package_name}Svc)${NC}"
    echo -e "   ${WHITE}gen${CLEAN_NAME}.Register${CLEAN_NAME}ServiceServer(grpcServer, ${package_name}GrpcHandler)${NC}\n"

    echo -e "${CYAN}3. Rapihkan Dependencies:${NC}"
    echo -e "   ${WHITE}go mod tidy${NC}\n"
}

# Run main function with all arguments
main "$@"


