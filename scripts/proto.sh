#!/bin/bash

# Flexible gRPC and Protobuf code generator for the service project.
#
# This script can:
# 1. Generate code for a specific proto directory.
# 2. Automatically find and generate code for all proto files in the project.

# --- Configuration ---
set -e # Exit immediately if a command exits with a non-zero status.
set -o pipefail # Return value of a pipeline is the value of the last command to exit with a non-zero status

# --- Colors for output ---
readonly NC='\033[0m' # No Color
readonly RED='\033[0;31m'
readonly GREEN='\033[0;32m'
readonly YELLOW='\033[1;33m'
readonly BLUE='\033[0;34m'
readonly PURPLE='\033[0;35m'
readonly CYAN='\033[0;36m'

# --- Helper Functions ---
log_info() { echo -e "${BLUE}[INFO]${NC} $1"; }
log_success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
log_warning() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
log_error() { echo -e "${RED}[ERROR]${NC} $1"; }
log_header() { echo -e "\n${PURPLE}==== $1 ====${NC}"; }

# --- Show Help ---
show_help() {
    cat << EOF
${YELLOW}Flexible gRPC and Protobuf Code Generator${NC}

This script generates Go code from .proto files using the recommended 'paths=import' method.

${YELLOW}USAGE:${NC}
  $(basename "$0") [path_to_proto_dir]

${YELLOW}DESCRIPTION:${NC}
  - If a ${CYAN}[path_to_proto_dir]${NC} is provided, it generates code only for .proto files in that directory.
    ${CYAN}Example:${NC} $(basename "$0") internal/infrastructure/transport/grpc/proto/permission/v1

  - If no path is provided, it automatically finds and generates code for ${CYAN}all .proto files${NC} within the project.
    ${CYAN}Example:${NC} $(basename "$0")

${YELLOW}REQUIREMENTS:${NC}
  - protoc
  - protoc-gen-go
  - protoc-gen-go-grpc

Make sure these are installed and available in your system's PATH.
To install them:
  go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
  go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
EOF
}

# --- Check for required tools ---
check_dependencies() {
    log_header "Checking Dependencies"
    local missing_deps=false
    for cmd in protoc protoc-gen-go protoc-gen-go-grpc; do
        if ! command -v "$cmd" &> /dev/null; then
            log_error "Dependency not found: ${CYAN}$cmd${NC}"
            missing_deps=true
        else
            log_success "Found: ${CYAN}$cmd${NC}"
        fi
    done

    if [ "$missing_deps" = true ]; then
        log_error "Please install the missing dependencies. See help for instructions."
        show_help
        exit 1
    fi
}

# --- Main Logic ---
main() {
    # Handle help flag
    if [[ "$1" == "-h" || "$1" == "--help" ]]; then
        show_help
        exit 0
    fi

    check_dependencies

    # Get project root directory
    local project_root
    project_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
    cd "$project_root"
    log_info "Operating from project root: ${CYAN}$project_root${NC}"

    local proto_files=()
    if [ -n "$1" ]; then
        # Case 1: A specific directory is provided
        local proto_dir_path="$1"
        if [ ! -d "$project_root/$proto_dir_path" ]; then
            log_info "Directory not found. Creating new directory: ${CYAN}$proto_dir_path${NC}"
            mkdir -p "$project_root/$proto_dir_path"
        fi
        log_header "Generating for specific directory: ${CYAN}$proto_dir_path${NC}"
        while IFS= read -r -d '' file; do
            proto_files+=("$file")
        done < <(find "$proto_dir_path" -name '*.proto' -print0)
    else
        # Case 2: No directory provided, find all protos
        log_header "Generating for all .proto files in the project"
        while IFS= read -r -d '' file; do
            proto_files+=("$file")
        done < <(find . -name '*.proto' -not -path './vendor/*' -print0)
    fi

    if [ ${#proto_files[@]} -eq 0 ]; then
        log_warning "No .proto files found to generate."
        exit 0
    fi

    log_info "Found the following .proto files to process:"
    for f in "${proto_files[@]}"; do echo -e "${CYAN}$f${NC}"; done

    # --- Generation Process ---
    log_header "Starting Code Generation"

    # The protoc command uses 'paths=import', which respects the 'go_package' option in .proto files.
    # This is the modern and recommended approach.
    # -I. : Search for imports in the project root.
    # --go_out=. : Output generated files relative to the project root.
    protoc -I. \
           --experimental_allow_proto3_optional \
           --go_out=. --go_opt=paths=import \
           --go-grpc_out=. --go-grpc_opt=paths=import \
           "${proto_files[@]}"

    if [ $? -ne 0 ]; then
        log_error "Protocol Buffer code generation failed."
        log_error "Please check the output from 'protoc' above for details."
        exit 1
    fi

    log_success "Protocol Buffer code generated successfully."

    # --- Post-generation ---
    log_header "Post-generation Steps"
    log_info "Running 'go mod tidy' to sync dependencies..."
    go mod tidy
    log_success "'go mod tidy' completed."
    echo ""
    log_success "All tasks finished successfully! ✨"
}

# --- Run main function ---
main "$@"
