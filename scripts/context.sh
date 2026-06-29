#!/bin/bash

# Advanced Context Generator for service
# Usage: ./scripts/context.sh [OPTIONS]
# Options:
#   -s, --sql FILE         Parse SQL file and generate context (mutually exclusive with -j)
#   -j, --json FILE        Parse JSON response file and generate context (mutually exclusive with -s)
#   -t, --table NAME       Specify explicit table name (useful when using JSON)
#   -d, --dir PATH         Custom directory structure (e.g., master/reference/province) (required)
#   -g, --generate TYPE    What to generate: domain|handler|all (default: all)
#   -f, --format           Output format for docs: code|markdown|json|yaml (default: code)
#   -o, --output           Output directory for documentation (default: ./generated-contexts)
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
readonly DEFAULT_OUTPUT_DIR="${PROJECT_ROOT}/generated-contexts"
readonly DEFAULT_FORMAT="code"
readonly INTERNAL_DIR="${PROJECT_ROOT}/internal"

# Global variables
VERBOSE=false
OUTPUT_DIR="${DEFAULT_OUTPUT_DIR}"
FORMAT="${DEFAULT_FORMAT}"
GENERATE_TYPE="all" # domain, handler, all
SQL_FILE=""
JSON_FILE=""
TABLE_NAME=""
CUSTOM_DIR=""

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
 ${WHITE}Advanced Context Generator for service${NC}

 ${YELLOW}Usage:${NC}
    $(basename "$0") -s SQL_FILE -d CUSTOM_DIR [OPTIONS]
    $(basename "$0") -j JSON_FILE -d CUSTOM_DIR -t TABLE_NAME [OPTIONS]

 ${YELLOW}Required Options:${NC}
    -s, --sql FILE         SQL file with CREATE TABLE statement
    -j, --json FILE        JSON response payload file
    -d, --dir PATH         Custom directory structure (e.g., master/reference/province)

 ${YELLOW}Optional Options:${NC}
    -t, --table NAME       Explicit table name (recommended when using -j)
    -g, --generate TYPE    What to generate: domain|handler|proto|grpc|all (default: all)
                           - 'domain': Generates entity, dto, mapper, repository, service
                           - 'handler': Generates HTTP handler
                           - 'proto': Generates gRPC .proto file
                           - 'grpc': Generates gRPC handler and mapper
                           - 'all': Generates both domain and handler
    -o, --output           Output directory for documentation (default: ${DEFAULT_OUTPUT_DIR})
    -v, --verbose          Verbose output
    -h, --help             Show this help message

 ${YELLOW}Examples:${NC}
    # Generate domain + handler for Ethnic
    $(basename "$0") -s db/migrations/001_create_ethnic_table.sql -d master/reference/ethnic

    # Generate only domain files for Province
    $(basename "$0") -s db/migrations/002_create_province_table.sql -d master/reference/province -g domain

    # Generate only gRPC proto file for Pages
    $(basename "$0") -s db/migrations/001_create_role_pages.sql -d master/role/pages -g proto

    # Generate only HTTP handler for District
    $(basename "$0") -s db/migrations/003_create_district_table.sql -d master/reference/district -g handler
EOF
}

# PascalCase from snake_case or kebab-case
to_pascal_case() {
    local str="$1"
    echo "$str" | sed -E 's/(^|[-_.])([a-zA-Z])/\U\2/g'
}

to_snake_case() {
    local str="$1"
    # Tambahkan underscore sebelum huruf besar (kecuali di awal), lalu ubah ke huruf kecil
    echo "$str" | sed 's/\([A-Z]\)/_\L\1/g' | sed 's/^_//'
}

# camelCase from snake_case or kebab-case
to_camel_case() {
    local str="$1"
    local pascal=$(to_pascal_case "$str")
    echo "$(echo "${pascal:0:1}" | tr '[:upper:]' '[:lower:]')${pascal:1}"
}
to_lower_case() {
    local str="$1"
    echo "$str" | tr '[:upper:]' '[:lower:]'
}

# Helper: Cek apakah field ini merupakan kolom otomatis (managed)
is_managed_field() {
    local col_name="$1"
    local lower_col=$(to_lower_case "$col_name")
    if [[ "$col_name" == "$DB_PK_NAME" || "$lower_col" == "created_at" || "$lower_col" == "createdat" || "$lower_col" == "updated_at" || "$lower_col" == "updatedat" || "$lower_col" == "deleted_at" || "$lower_col" == "deletedat" ]]; then
        return 0
    fi
    return 1
}

# Extract package name from custom directory structure
# For path master/reference/province, returns "province"
extract_package_name() {
    local custom_dir="$1"
    basename "$custom_dir"
}

# Convert SQL type to Go type
sql_to_go_type() {
    local sql_type="$1"
    local nullable="${2:-false}"
    
    local go_type="interface{}"
    case "$sql_type" in
        "smallserial"|"serial"|"bigserial"|"smallint"|"integer"|"bigint"|"int") go_type="int64" ;;
        "varchar"|"text"|"char"|"character varying") go_type="string" ;;
        "boolean"|"bool") go_type="bool" ;;
        "timestamp"|"timestamptz"|"date") go_type="time.Time" ;;
        "decimal"|"numeric"|"real"|"double precision"|"float") go_type="float64" ;;
        "uuid") go_type="uuid.UUID" ;;
    esac

    # Gunakan pointer untuk kolom nullable agar tipe data strict
    if [[ "$nullable" == "true" && "$go_type" != "interface{}" ]]; then
        echo "*$go_type"
    else
        echo "$go_type"
    fi
}

# Convert SQL type to Protobuf type
sql_to_proto_type() {
    local sql_type="$1"
    local nullable="${2:-false}"

    local proto_type="string" # Default
    case "$sql_type" in
        "smallserial"|"serial"|"bigserial"|"smallint"|"integer"|"bigint"|"int") proto_type="int64" ;;
        "varchar"|"text"|"char"|"character varying"|"uuid") proto_type="string" ;;
        "boolean"|"bool") proto_type="bool" ;;
        "timestamp"|"timestamptz"|"date") proto_type="google.protobuf.Timestamp" ;;
        "decimal"|"numeric") proto_type="string" ;; # Represent numeric as string for precision
        "real"|"double precision"|"float") proto_type="double" ;;
    esac

    # Use optional for nullable fields (proto3)
    if [[ "$nullable" == "true" ]]; then
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
            -s|--sql)
                SQL_FILE="$2"
                shift 2
                ;;
            -j|--json)
                JSON_FILE="$2"
                shift 2
                ;;
            -t|--table)
                TABLE_NAME="$2"
                shift 2
                ;;
            -d|--dir)
                CUSTOM_DIR="$2"
                shift 2
                ;;
            -g|--generate)
                GENERATE_TYPE="$2"
                if [[ "$GENERATE_TYPE" != "domain" && "$GENERATE_TYPE" != "handler" && "$GENERATE_TYPE" != "proto" && "$GENERATE_TYPE" != "grpc" && "$GENERATE_TYPE" != "all" ]]; then
                    log_error "Invalid generate type: $GENERATE_TYPE. Use 'domain', 'handler', 'proto', 'grpc', or 'all'."
                    exit 1
                fi
                shift 2
                ;;
            -f|--format)
                FORMAT="$2"
                shift 2
                ;;
            -o|--output)
                OUTPUT_DIR="$2"
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
    if [[ -z "$SQL_FILE" && -z "$JSON_FILE" ]]; then
        log_error "Either SQL file (-s) or JSON file (-j) is required."
        exit 1
    fi
    if [[ -n "$SQL_FILE" && ! -f "$SQL_FILE" ]]; then
        log_error "SQL file not found: $SQL_FILE"
        exit 1
    fi
    if [[ -n "$JSON_FILE" && ! -f "$JSON_FILE" ]]; then
        log_error "JSON file not found: $JSON_FILE"
        exit 1
    fi
    if [[ -z "$CUSTOM_DIR" ]]; then
        log_error "Custom directory is required. Use -d or --dir."
        exit 1
    fi
}

# Extract table info from SQL
parse_sql_table() {
    local sql_file="$1"
    log_info "Parsing SQL table from: $sql_file"

    # Extract table name (more robust regex)
    local table_name
    table_name=$(grep -i "CREATE TABLE" "$sql_file" | sed -E 's/.*CREATE[[:space:]]+TABLE[[:space:]]*(IF[[:space:]]+NOT[[:space:]]+EXISTS[[:space:]]+)?[[:space:]]*"?([^"[:space:]]+)"?.*/\2/i' | head -n 1)
    
    if [[ -z "$table_name" ]]; then
        log_error "Could not extract table name from SQL file"
        return 1
    fi
    log_verbose "Found table: $table_name"

    # Reset arrays
    PARSED_COLUMNS=()
    PARSED_PRIMARY_KEY=""

    # Use a temp file to process line by line, preserving formatting
    local temp_file=$(mktemp)
    # Pre-process: remove comments and ensure each column definition is on one line
    grep -v '^[[:space:]]*--' "$sql_file" | sed ':a;N;$!ba;s/\n[[:space:]]*,/ /g' > "$temp_file"

    # Extract columns
    while IFS= read -r line; do
        # Remove leading/trailing spaces
        local trimmed=$(echo "$line" | sed -e 's/^[[:space:]]*//' -e 's/[[:space:]]*$//')
        
        # Skip empty lines, brackets, or constraints
        if [[ -z "$trimmed" ]] || [[ "$trimmed" == ")" ]] || [[ "$trimmed" == ");" ]] || [[ "$trimmed" =~ ^CREATE[[:space:]]+TABLE ]] || [[ "$trimmed" =~ ^CONSTRAINT ]] || [[ "$trimmed" =~ ^PRIMARY ]] || [[ "$trimmed" =~ ^FOREIGN ]] || [[ "$trimmed" =~ ^UNIQUE ]]; then
            continue
        fi

        # Extract column name (first word, remove quotes)
        local col_name=$(echo "$trimmed" | awk '{print $1}' | tr -d '"')
        
        # Extract column type (second word, remove size parameters like (100) and commas)
        local col_type=$(echo "$trimmed" | awk '{print $2}' | sed -E 's/\([0-9,]+\)//g' | tr -d ',' | tr '[:upper:]' '[:lower:]')
        
        # Check if nullable
        local nullable="true"
        if [[ "$trimmed" =~ NOT[[:space:]]+NULL ]]; then
            nullable="false"
        fi

        PARSED_COLUMNS+=("${col_name}|${col_type}|${nullable}")
        log_verbose "Found column: $col_name ($col_type, nullable: $nullable)"

    done < "$temp_file"

    # Extract primary key
    local pk_line=$(grep -i "PRIMARY KEY" "$sql_file")
    if [[ -n "$pk_line" ]]; then
        PARSED_PRIMARY_KEY=$(echo "$pk_line" | sed -E 's/.*PRIMARY[[:space:]]+KEY[[:space:]]*\(([^)]+)\).*/\1/' | tr -d '"')
        log_verbose "Found Primary Key: $PARSED_PRIMARY_KEY"
    fi

    # Clean up temp file
    rm -f "$temp_file"

    # Store in global variables for later use
    PARSED_TABLE_NAME="$table_name"
    
    export HAS_CREATED_AT=false
    export HAS_UPDATED_AT=false
    export HAS_DELETED_AT=false
    export HAS_ACTIVE=false
    export GO_PK_NAME="Id"
    export DB_PK_NAME="id"
    export GO_PK_TYPE="int64"
    export PROTO_PK_TYPE="int64"

    local pk_col_type="integer"
    for col in "${PARSED_COLUMNS[@]}"; do
        IFS='|' read -r col_name col_type nullable <<< "$col"
        local lower_col=$(to_lower_case "$col_name")
        if [[ "$lower_col" == "created_at" || "$lower_col" == "createdat" ]]; then HAS_CREATED_AT=true; fi
        if [[ "$lower_col" == "updated_at" || "$lower_col" == "updatedat" ]]; then HAS_UPDATED_AT=true; fi
        if [[ "$lower_col" == "deleted_at" || "$lower_col" == "deletedat" ]]; then HAS_DELETED_AT=true; fi
        if [[ "$lower_col" == "active" ]]; then HAS_ACTIVE=true; fi
        if [[ "$col_name" == "$PARSED_PRIMARY_KEY" || (-z "$PARSED_PRIMARY_KEY" && "$lower_col" == "id") ]]; then
            PARSED_PRIMARY_KEY="$col_name"
            pk_col_type="$col_type"
        fi
    done
    if [[ -z "$PARSED_PRIMARY_KEY" && ${#PARSED_COLUMNS[@]} -gt 0 ]]; then
        IFS='|' read -r col_name col_type _ <<< "${PARSED_COLUMNS[0]}"
        PARSED_PRIMARY_KEY="$col_name"
        pk_col_type="$col_type"
    fi

    GO_PK_NAME=$(to_pascal_case "$PARSED_PRIMARY_KEY")
    DB_PK_NAME="$PARSED_PRIMARY_KEY"
    GO_PK_TYPE=$(sql_to_go_type "$pk_col_type" "false")
    PROTO_PK_TYPE=$(sql_to_proto_type "$pk_col_type" "false")

    log_success "Successfully parsed table: $table_name"
}

# Extract table info from JSON payload
parse_json_payload() {
    local json_file="$1"
    log_info "Parsing JSON payload from: $json_file"

    if ! command -v jq &> /dev/null; then
        log_error "'jq' is required to parse JSON. Please install it (e.g., sudo apt-get install jq)."
        exit 1
    fi

    local table_name="$TABLE_NAME"
    if [[ -z "$table_name" ]]; then
        table_name=$(basename "$json_file" | sed 's/\.[^.]*$//' | tr '-' '_')
        log_warning "Table name not provided, inferring from filename: $table_name"
    fi

    PARSED_COLUMNS=()
    PARSED_PRIMARY_KEY=""

    local temp_file=$(mktemp)
    # Analyze JSON structure using jq. It handles both array of objects and single object.
    jq -r '
    (if type == "array" then .[0] else . end) |
    to_entries | .[] |
    .key as $k |
    (.value | type) as $t |
    if $t == "string" then
      if (.value | test("^[0-9]{4}-[0-9]{2}-[0-9]{2}T")) then "\($k)|timestamp|true"
      else "\($k)|varchar|true" end
    elif $t == "number" then
      if (.value | tostring | test("\\.")) then "\($k)|float|true" else "\($k)|integer|true" end
    elif $t == "boolean" then "\($k)|boolean|true"
    elif $t == "object" or $t == "array" then "\($k)|jsonb|true"
    else "\($k)|varchar|true" end
    ' "$json_file" > "$temp_file"

    while IFS='|' read -r col_name col_type nullable; do
        if [[ -z "$col_name" ]]; then continue; fi
        PARSED_COLUMNS+=("${col_name}|${col_type}|${nullable}")
        log_verbose "Found property: $col_name ($col_type)"
    done < "$temp_file"
    rm -f "$temp_file"

    if [[ ${#PARSED_COLUMNS[@]} -eq 0 ]]; then
        log_error "No properties found in JSON file. Is it empty?"
        exit 1
    fi

    PARSED_TABLE_NAME="$table_name"
    export HAS_CREATED_AT=false
    export HAS_UPDATED_AT=false
    export HAS_DELETED_AT=false
    export HAS_ACTIVE=false
    export GO_PK_NAME="Id"
    export DB_PK_NAME="id"

    local pk_col_type="varchar" # Default fallback for JSON
    for col in "${PARSED_COLUMNS[@]}"; do
        IFS='|' read -r col_name col_type nullable <<< "$col"
        local lower_col=$(to_lower_case "$col_name")
        if [[ "$lower_col" == "created_at" || "$lower_col" == "createdat" ]]; then HAS_CREATED_AT=true; fi
        if [[ "$lower_col" == "updated_at" || "$lower_col" == "updatedat" ]]; then HAS_UPDATED_AT=true; fi
        if [[ "$lower_col" == "deleted_at" || "$lower_col" == "deletedat" ]]; then HAS_DELETED_AT=true; fi
        if [[ "$lower_col" == "active" ]]; then HAS_ACTIVE=true; fi
        
        if [[ -z "$PARSED_PRIMARY_KEY" ]]; then
            if [[ "$lower_col" == "id" || "$lower_col" == "${table_name}_id" || "$lower_col" == "kode" || "$lower_col" == "uuid" ]]; then
                PARSED_PRIMARY_KEY="$col_name"
                pk_col_type="$col_type"
            fi
        fi
    done

    if [[ -z "$PARSED_PRIMARY_KEY" ]]; then
        IFS='|' read -r col_name col_type _ <<< "${PARSED_COLUMNS[0]}"
        PARSED_PRIMARY_KEY="$col_name"
        pk_col_type="$col_type"
    fi

    GO_PK_NAME=$(to_pascal_case "$PARSED_PRIMARY_KEY")
    DB_PK_NAME="$PARSED_PRIMARY_KEY"
    GO_PK_TYPE=$(sql_to_go_type "$pk_col_type" "false")
    PROTO_PK_TYPE=$(sql_to_proto_type "$pk_col_type" "false")

    log_success "Successfully parsed JSON for table: $table_name"
}

# Generate domain files
generate_domain_files() {
    local table_name="$1"
    local clean_name=$(to_pascal_case "$table_name")
    local target_dir="${INTERNAL_DIR}/${CUSTOM_DIR}"
    
    local package_name=$(extract_package_name "$CUSTOM_DIR")

    log_info "Generating domain files in: $target_dir with package name: $package_name"
    mkdir -p "$target_dir"

    local entity_file="${target_dir}/entity.go"
    if [ -f "$entity_file" ]; then
        log_warning "File already exists, skipping: $entity_file"
    else
        log_info "Generating file: $entity_file"
        cat > "$entity_file" << EOF
package ${package_name}

import (
    "time"
)

// ${clean_name} entity represents the ${table_name} table in the database
type ${clean_name} struct {
    ${GO_PK_NAME}        ${GO_PK_TYPE}       \`json:"${DB_PK_NAME}" db:"${DB_PK_NAME}"\`
    DeletedAt *time.Time  \`json:"deleted_at" db:"deleted_at"\`
    CreatedAt *time.Time  \`json:"created_at" db:"created_at"\`
    UpdatedAt *time.Time  \`json:"updated_at" db:"updated_at"\`
EOF

    # Tambahkan kolom-kolom dari SQL (kecuali Id, CreatedAt, UpdatedAt, DeletedAt)
    for col in "${PARSED_COLUMNS[@]}"; do
        IFS='|' read -r col_name col_type nullable <<< "$col"
        # Lewati kolom standar yang sudah kita definisikan manual
        if [[ "$(to_lower_case "$col_name")" == "id" || "$(to_lower_case "$col_name")" == "created_at" || "$(to_lower_case "$col_name")" == "updated_at" || "$(to_lower_case "$col_name")" == "deleted_at" || "$col_name" == "${GO_PK_NAME}" || "$col_name" == "${DB_PK_NAME}" || "$col_name" == "CreatedAt" || "$col_name" == "UpdatedAt" || "$col_name" == "DeletedAt" ]]; then
            continue
        fi

        go_type=$(sql_to_go_type "$col_type" "$nullable")
        go_field_name=$(to_pascal_case "$col_name")
        db_tag=$(to_lower_case "$col_name") # Nama kolom di DB (Lower case)
        json_tag=$(to_lower_case "$col_name") # Nama kolom di JSON (camelCase)

        echo "	${go_field_name} ${go_type} \`json:\"${json_tag}\" db:\"${db_tag}\"\`" >> "$entity_file"
    done

    cat >> "$entity_file" << EOF
}

// TableName specifies the table name for ${clean_name}
func (${clean_name}) TableName() string {
    return "${table_name}"
}
EOF
    fi

    # --- Generate dto.go ---
    local dto_file="${target_dir}/dto.go"
    if [ -f "$dto_file" ]; then
        log_warning "File already exists, skipping: $dto_file"
    else
        log_info "Generating file: $dto_file"

        local dto_imports=""
        for col in "${PARSED_COLUMNS[@]}"; do
            IFS='|' read -r col_name col_type _ <<< "$col"
            if [[ "$col_type" == *"timestamp"* || "$col_type" == *"timestamptz"* || "$col_type" == *"date"* ]]; then
                dto_imports="import \"time\""
                break
            fi
        done

        cat > "$dto_file" << EOF
package ${package_name}

${dto_imports}

// ${clean_name}Request represents the request payload for ${clean_name}
type ${clean_name}Request struct {
EOF
    for col in "${PARSED_COLUMNS[@]}"; do
        IFS='|' read -r col_name col_type nullable <<< "$col"
        if is_managed_field "$col_name"; then
            continue
        fi
        go_type=$(sql_to_go_type "$col_type" "$nullable")
        go_field_name=$(to_pascal_case "$col_name")
        json_tag=$(to_lower_case "$col_name")
        validate_tag=""
        if [[ "$nullable" == "false" ]]; then
            validate_tag=" validate:\"required\""
        fi
        echo "	${go_field_name} ${go_type} \`json:\"${json_tag}\"${validate_tag}\`" >> "$dto_file"
    done
    cat >> "$dto_file" << EOF
}

// ${clean_name}Response represents the response payload for ${clean_name}
type ${clean_name}Response struct {
EOF
    for col in "${PARSED_COLUMNS[@]}"; do
        IFS='|' read -r col_name col_type nullable <<< "$col"
        go_type=$(sql_to_go_type "$col_type" "$nullable")
        go_field_name=$(to_pascal_case "$col_name")
        json_tag=$(to_lower_case "$col_name")
        echo "	${go_field_name} ${go_type} \`json:\"${json_tag}\"\`" >> "$dto_file"
    done
    echo "}" >> "$dto_file"
    fi

    # --- Generate mapper.go ---
    local mapper_file="${target_dir}/mapper.go"
    if [ -f "$mapper_file" ]; then
        log_warning "File already exists, skipping: $mapper_file"
    else
        log_info "Generating file: $mapper_file"
        cat > "$mapper_file" << EOF
package ${package_name}

// mapRequestToEntity converts ${clean_name}Request to *${clean_name} entity
func mapRequestToEntity(req ${clean_name}Request) *${clean_name} {
    return &${clean_name}{
EOF
    for col in "${PARSED_COLUMNS[@]}"; do
        IFS='|' read -r col_name col_type nullable <<< "$col"
        if is_managed_field "$col_name"; then
            continue
        fi
        go_field_name=$(to_pascal_case "$col_name")
        echo "		${go_field_name}: req.${go_field_name}," >> "$mapper_file"
    done
    cat >> "$mapper_file" << EOF
    }
}

// mapEntityToResponse converts *${clean_name} entity to *${clean_name}Response DTO
func mapEntityToResponse(e *${clean_name}) *${clean_name}Response {
    if e == nil {
        return nil
    }
    return &${clean_name}Response{
EOF
    for col in "${PARSED_COLUMNS[@]}"; do
        IFS='|' read -r col_name col_type nullable <<< "$col"
        go_field_name=$(to_pascal_case "$col_name")
        echo "		${go_field_name}: e.${go_field_name}," >> "$mapper_file"
    done
    cat >> "$mapper_file" << EOF
    }
}
EOF
    fi

    # --- Generate repository.go ---
    local repo_file="${target_dir}/repository.go"
    if [ -f "$repo_file" ]; then
        log_warning "File already exists, skipping: $repo_file"
    else
        log_info "Generating file: $repo_file"
        cat > "$repo_file" << EOF
package ${package_name}

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "strconv"
    "strings"

    "service/internal/infrastructure/database"
    "service/pkg/utils/query"
    "service/pkg/logger"
    "github.com/jmoiron/sqlx"
    "gorm.io/gorm"
)

type CommandRepository interface {
    Create(ctx context.Context, entity *${clean_name}) error
    Update(ctx context.Context, entity *${clean_name}) error
    Delete(ctx context.Context, id ${GO_PK_TYPE}) error
}

type QueryRepository interface {
    FindAll(ctx context.Context, limit, offset int) ([]${clean_name}, int64, error)
    FindByID(ctx context.Context, id ${GO_PK_TYPE}) (*${clean_name}, error)
    Search(ctx context.Context, filters map[string]interface{}, sorts []query.SortField, limit, offset int) ([]${clean_name}, int64, error)
}

type repository struct {
    dbManager database.Service
    dbName    string
    qb        query.QueryBuilder
    dbType    query.DBType
    allowedColumnsMap map[string]bool
}

func NewCommandRepository(dbManager database.Service, dbName string) CommandRepository {
    return NewRepository(dbManager, dbName)
}

func NewQueryRepository(dbManager database.Service, dbName string) QueryRepository {
    return NewRepository(dbManager, dbName)
}

func NewRepository(dbManager database.Service, dbName string) *repository {
    dbType := query.DBTypePostgreSQL
    allowedColumns := []string{
EOF
    for col in "${PARSED_COLUMNS[@]}"; do
        IFS='|' read -r col_name _ _ <<< "$col"
        echo "		\"$(to_lower_case "$col_name")\"," >> "$repo_file"
    done
    cat >> "$repo_file" << EOF
    }
    qb := query.NewSQLQueryBuilder(dbType).
        SetSecurityOptions(true, 1000).
        SetQueryLogging(true).
        SetQueryTimeout(30).
        SetAllowedColumns(allowedColumns)

    allowedMap := make(map[string]bool)
    for _, col := range allowedColumns {
        allowedMap[col] = true
    }

    return &repository{dbManager: dbManager, dbName: dbName, qb: qb, dbType: dbType, allowedColumnsMap: allowedMap}
}

// isColumnAllowed memvalidasi apakah kolom diizinkan untuk digunakan
func (r *repository) isColumnAllowed(column string) bool {
    return r.allowedColumnsMap[column]
}

// getWriteGormDB extracts *gorm.DB for Write/Command operations
func (r *repository) getWriteGormDB() (*gorm.DB, error) {
    return r.dbManager.GetGormDB(r.dbName)
}

// getReadSQLXDB extracts *sqlx.DB from read replicas for Read/Query operations
func (r *repository) getReadSQLXDB() (*sqlx.DB, error) {
    db, err := r.dbManager.GetReadDB(r.dbName)
    if err != nil {
        return nil, fmt.Errorf("failed to get read db: %w", err)
    }
    // Gunakan Unsafe() agar sqlx mengabaikan kolom hasil query yang tidak terdapat di dalam struct
    return sqlx.NewDb(db, "pgx").Unsafe(), nil // Gunakan pgx untuk PostgreSQL
}

EOF
    local default_sort_col="${DB_PK_NAME}"
    if [[ "$HAS_CREATED_AT" == "true" ]]; then default_sort_col="created_at"; fi
    cat >> "$repo_file" << EOF
// FindAll fetches all ${clean_name} with pagination
func (r *repository) FindAll(ctx context.Context, limit, offset int) ([]${clean_name}, int64, error) {
    db, err := r.getReadSQLXDB()
    if err != nil { return nil, 0, err }
    
    dq := query.DynamicQuery{
        From: "${table_name}",
EOF
    if [[ "$HAS_DELETED_AT" == "true" ]]; then
        echo "        Filters: []query.FilterGroup{{ Filters: []query.DynamicFilter{query.CreateFilter(\"deleted_at\", query.OpNull, nil)} }}," >> "$repo_file"
    fi
    cat >> "$repo_file" << EOF
        Limit: limit, Offset: offset,
        Sort: []query.SortField{query.CreateAscSort("${DB_PK_NAME}")},
    }
    
    var results []${clean_name}
    if err := r.qb.ExecuteQuery(ctx, db, dq, &results); err != nil {
        return nil, 0, fmt.Errorf("failed to execute find all query: %w", err)
    }
    
    count, err := r.qb.ExecuteCount(ctx, db, dq)
    if err != nil { return nil, 0, fmt.Errorf("failed to execute count query: %w", err) }
    
    return results, count, nil
}

// FindByID fetches a single ${clean_name} by ID
func (r *repository) FindByID(ctx context.Context, id ${GO_PK_TYPE}) (*${clean_name}, error) {
    db, err := r.getReadSQLXDB()
    if err != nil { return nil, err }
    
    var result ${clean_name}
    q := query.DynamicQuery{
        From: "${table_name}",
        Filters: []query.FilterGroup{{
            Filters: []query.DynamicFilter{
                query.CreateEqualFilter("${DB_PK_NAME}", id),
EOF
    if [[ "$HAS_DELETED_AT" == "true" ]]; then
        echo "                query.CreateFilter(\"deleted_at\", query.OpNull, nil)," >> "$repo_file"
    fi
    cat >> "$repo_file" << EOF
            },
        }},
        Limit: 1,
    }
    
    if err := r.qb.ExecuteQueryRow(ctx, db, q, &result); err != nil {
        if errors.Is(err, sql.ErrNoRows) { 
            return nil, nil // Return nil, nil jika data tidak ada, jangan lempar error query
        }
        // Wrapping error origin
        return nil, fmt.Errorf("failed to fetch ${clean_name}: %w", err)
    }
    return &result, nil
}

// Search fetches ${clean_name} based on dynamic filters and sorting
func (r *repository) Search(ctx context.Context, filters map[string]interface{}, sorts []query.SortField, limit, offset int) ([]${clean_name}, int64, error) {
    db, err := r.getReadSQLXDB()
    if err != nil { return nil, 0, err }
    
    var dynamicFilters []query.DynamicFilter
    for k, v := range filters {
        colName := strings.ToLower(k)
        if !r.isColumnAllowed(colName) {
            continue
        }

        switch val := v.(type) {
        case string:
            if val != "" {
EOF
    # Check text columns to use OpILike
    text_cols=()
    for col in "${PARSED_COLUMNS[@]}"; do
        IFS='|' read -r col_name col_type _ <<< "$col"
        if [[ "$col_type" == *"varchar"* || "$col_type" == *"text"* || "$col_type" == *"char"* ]]; then
            text_cols+=("colName == \"$(to_lower_case "$col_name")\"")
        fi
    done
    
    if [ ${#text_cols[@]} -gt 0 ]; then
        text_cond=$(printf " || %s" "${text_cols[@]}")
        text_cond=${text_cond:4}
        cat >> "$repo_file" << EOF
                if $text_cond {
                    dynamicFilters = append(dynamicFilters, query.CreateFilter(colName, query.OpILike, "%"+val+"%"))
                } else {
                    if boolVal, err := strconv.ParseBool(val); err == nil {
                        dynamicFilters = append(dynamicFilters, query.CreateEqualFilter(colName, boolVal))
                    } else {
                        dynamicFilters = append(dynamicFilters, query.CreateEqualFilter(colName, val))
                    }
                }
EOF
    else
        cat >> "$repo_file" << EOF
                if boolVal, err := strconv.ParseBool(val); err == nil {
                    dynamicFilters = append(dynamicFilters, query.CreateEqualFilter(colName, boolVal))
                } else {
                    dynamicFilters = append(dynamicFilters, query.CreateEqualFilter(colName, val))
                }
EOF
    fi

    cat >> "$repo_file" << EOF
            }
        default:
            dynamicFilters = append(dynamicFilters, query.CreateEqualFilter(colName, val))
        }
    }

EOF
    if [[ "$HAS_DELETED_AT" == "true" ]]; then
        echo "    dynamicFilters = append(dynamicFilters, query.CreateFilter(\"deleted_at\", query.OpNull, nil))" >> "$repo_file"
    fi
    cat >> "$repo_file" << EOF

    var sortFields []query.SortField
    for _, sort := range sorts {
        colName := strings.ToLower(sort.Column)
        if r.isColumnAllowed(colName) {
            sortFields = append(sortFields, query.SortField{
                Column: colName,
                Order:  sort.Order,
            })
        }
    }

    // Jika tidak ada sort yang valid, gunakan default
    if len(sortFields) == 0 {
        sortFields = []query.SortField{query.CreateAscSort("${DB_PK_NAME}")}
    }

    q := query.DynamicQuery{
        From: "${table_name}",
        Fields: []query.SelectField{
            {Expression: "*"},
        },
        Filters: []query.FilterGroup{{Filters: dynamicFilters}},
        Limit: limit, Offset: offset,
        Sort: sortFields,
    }

    logger.Default().Info("Built search query", logger.String("request", fmt.Sprintf("%+v", q)))
    var results []${clean_name}
    if err := r.qb.ExecuteQuery(ctx, db, q, &results); err != nil {
        return nil, 0, fmt.Errorf("failed to execute search query: %w", err)
    }
    
    count, err := r.qb.ExecuteCount(ctx, db, q)
    if err != nil { return nil, 0, fmt.Errorf("failed to execute search count query: %w", err) }
    
    return results, count, nil
}
// Create inserts a new ${clean_name} record
func (r *repository) Create(ctx context.Context, entity *${clean_name}) error {
    db, err := r.getWriteGormDB()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Table(entity.TableName()).Create(entity).Error
}

func (r *repository) Update(ctx context.Context, entity *${clean_name}) error {
	db, err := r.getWriteGormDB()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Table(entity.TableName()).Save(entity).Error
}

func (r *repository) Delete(ctx context.Context, id ${GO_PK_TYPE}) error {
	db, err := r.getWriteGormDB()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Delete(&${clean_name}{}, id).Error
}

EOF
    fi

    # --- Generate service.go ---
    local service_file="${target_dir}/service.go"
    if [ -f "$service_file" ]; then
        log_warning "File already exists, skipping: $service_file"
    else
        log_info "Generating file: $service_file"
        cat > "$service_file" << EOF
package ${package_name}

import (
    "context"
    "service/pkg/utils/query"
	"strings"
    "service/pkg/errors"
    "gorm.io/gorm"
)

type Service interface {
    GetList(ctx context.Context, page, pageSize int, sorts []string) (map[string]interface{}, error)
    GetDetail(ctx context.Context, id ${GO_PK_TYPE}) (*${clean_name}Response, error)
    Search(ctx context.Context, filters map[string]interface{}, sorts []string, page, pageSize int) (map[string]interface{}, error)
    Create(ctx context.Context, req ${clean_name}Request) (*${clean_name}Response, error)
    Update(ctx context.Context, id ${GO_PK_TYPE}, req ${clean_name}Request) (*${clean_name}Response, error)
    Delete(ctx context.Context, id ${GO_PK_TYPE}) error
}

type service struct {
    cmdRepo   CommandRepository
    queryRepo QueryRepository
}

func NewService(cmdRepo CommandRepository, queryRepo QueryRepository) Service {
    return &service{cmdRepo: cmdRepo, queryRepo: queryRepo}
}

func (s *service) GetList(ctx context.Context, page, pageSize int, sorts []string) (map[string]interface{}, error) {
    return s.Search(ctx, map[string]interface{}{}, sorts, page, pageSize)
}

EOF
    local pk_val_check="id <= 0"
    if [[ "$GO_PK_TYPE" == "string" ]]; then pk_val_check="id == \"\""; fi
    
    cat >> "$service_file" << EOF
func (s *service) GetDetail(ctx context.Context, id ${GO_PK_TYPE}) (*${clean_name}Response, error) {
    if ${pk_val_check} { return nil, errors.NewValidationError().Message("Invalid ID").Metadata("id", id).Build() }

    entity, err := s.queryRepo.FindByID(ctx, id)
    if err != nil { return nil, errors.InternalError().Message("Failed to retrieve ${clean_name} detail").Cause(err).Build() }
    if entity == nil { return nil, errors.NotFoundError().Message("${clean_name} not found").Metadata("id", id).Build() }

    return mapEntityToResponse(entity), nil
}

func (s *service) Search(ctx context.Context, filters map[string]interface{}, sorts []string, page, pageSize int) (map[string]interface{}, error) {
    if page < 1 { page = 1 }
    if pageSize < 1 || pageSize > 100 { pageSize = 10 }
    offset := (page - 1) * pageSize

    // --- 1. Implementasi Caching (Check) ---
    filterBytes, _ := json.Marshal(filters)
    sortBytes, _ := json.Marshal(sorts)
    hashInput := fmt.Sprintf("%s|%s|%d|%d", string(filterBytes), string(sortBytes), page, pageSize)
    hash := sha256.Sum256([]byte(hashInput))
    cacheKey := fmt.Sprintf("${snake_case_name}_search_v2:%s", hex.EncodeToString(hash[:]))

    if s.cache != nil {
        var strData string
        if err := s.cache.Get(ctx, cacheKey, &strData); err == nil && strData != "" {
            var cachedData struct {
                Data     []*${clean_name}Response \`json:"data"\`
                Total    int64                    \`json:"total"\`
                Page     int                      \`json:"page"\`
                PageSize int                      \`json:"page_size"\`
            }
            if err := json.Unmarshal([]byte(strData), &cachedData); err == nil {
                logger.Default().Debug("Cache hit for ${clean_name} Search", logger.String("key", cacheKey))
                return map[string]interface{}{
                    "data":      cachedData.Data,
                    "total":     cachedData.Total,
                    "page":      cachedData.Page,
                    "page_size": cachedData.PageSize,
                }, nil
            }
        }
    }

    // Konversi sorts string ke SortField
    var sortFields []query.SortField
    for _, sort := range sorts {
        if sort == "" {
            continue
        }

        // Default ASC, tandai dengan - untuk DESC
        order := "ASC"
        column := sort
        if strings.HasPrefix(sort, "-") {
            order = "DESC"
            column = strings.TrimPrefix(sort, "-")
        } else if strings.HasPrefix(sort, "+") {
            column = strings.TrimPrefix(sort, "+")
        }

        // Validasi kolom yang diizinkan (lowercase untuk mapping)
        allowedColumns := map[string]bool{
EOF
    # Add allowed columns
    for col in "${PARSED_COLUMNS[@]}"; do
        IFS='|' read -r col_name _ _ <<< "$col"
        lower_name=$(to_lower_case "$col_name")
        echo "            \"${lower_name}\": true," >> "$service_file"
    done
    cat >> "$service_file" << EOF
        }

        if allowedColumns[column] {
            sortFields = append(sortFields, query.SortField{
                Column: column,
                Order:  order,
            })
        }
    }

    // Jika tidak ada sort yang valid, gunakan default
    if len(sortFields) == 0 {
        sortFields = []query.SortField{query.CreateAscSort("${DB_PK_NAME}")}
    }

    entities, total, err := s.queryRepo.Search(ctx, filters, sortFields, pageSize, offset)
    if err != nil { return nil, errors.InternalError().Message("Failed to search ${clean_name}s").Cause(err).Build() }

    responses := make([]*${clean_name}Response, len(entities))
    for i, entity := range entities { responses[i] = mapEntityToResponse(&entity) }

    responseMap := map[string]interface{}{
        "data": responses, "total": total, "page": page, "page_size": pageSize,
    }

    // --- 2. Implementasi Caching (Set) ---
    if s.cache != nil {
        if bytes, err := json.Marshal(responseMap); err == nil {
            _ = s.cache.Set(ctx, cacheKey, string(bytes), 5*time.Minute)
        }
    }

    return responseMap, nil
}

func (s *service) Create(ctx context.Context, req ${clean_name}Request) (*${clean_name}Response, error) {
    // TODO: Add validation here if needed
    entity := mapRequestToEntity(req)
    if err := s.cmdRepo.Create(ctx, entity); err != nil {
        if errors.Is(err, gorm.ErrDuplicatedKey) {
            return nil, errors.AlreadyExistsError().Message("${clean_name} with this identifier already exists").Cause(err).Build()
        }
        return nil, errors.InternalError().Message("Failed to create ${clean_name}").Cause(err).Build()
    }
    // Re-fetch to get the complete created entity
    createdEntity, err := s.queryRepo.FindByID(ctx, entity.${GO_PK_NAME})
    if err != nil { return nil, errors.InternalError().Message("Failed to retrieve newly created ${clean_name}").Cause(err).Build() }

    if s.cache != nil {
        _ = s.cache.Delete(ctx, "${snake_case_name}_search_v2:*")
    }

    return mapEntityToResponse(createdEntity), nil
}

func (s *service) Update(ctx context.Context, id ${GO_PK_TYPE}, req ${clean_name}Request) (*${clean_name}Response, error) {
    if ${pk_val_check} { return nil, errors.NewValidationError().Message("Invalid ID").Metadata("id", id).Build() }

    // Cek apakah record ada
    existing, err := s.queryRepo.FindByID(ctx, id)
    if err != nil { return nil, errors.InternalError().Message("Failed to retrieve ${clean_name}").Cause(err).Build() }
    if existing == nil { return nil, errors.NotFoundError().Message("${clean_name} not found").Metadata("id", id).Build() }

    // Update fields from request on the existing entity
EOF
    for col in "${PARSED_COLUMNS[@]}"; do
        IFS='|' read -r col_name _ _ <<< "$col"
        if is_managed_field "$col_name"; then
            continue
        fi
        go_field_name=$(to_pascal_case "$col_name")
        echo "    existing.${go_field_name} = req.${go_field_name}" >> "$service_file"
    done
    cat >> "$service_file" << EOF

    if err := s.cmdRepo.Update(ctx, existing); err != nil {
        if errors.Is(err, gorm.ErrDuplicatedKey) {
            return nil, errors.AlreadyExistsError().Message("${clean_name} with this identifier already exists").Cause(err).Build()
        }
        return nil, errors.InternalError().Message("Failed to update ${clean_name}").Cause(err).Build()
    }

    if s.cache != nil {
        _ = s.cache.Delete(ctx, "${snake_case_name}_search_v2:*")
    }

    return mapEntityToResponse(existing), nil
}

func (s *service) Delete(ctx context.Context, id ${GO_PK_TYPE}) error {
    if ${pk_val_check} { return errors.NewValidationError().Message("Invalid ID").Metadata("id", id).Build() }

    // Cek apakah record ada sebelum dihapus untuk idempotency dan pre-delete logic
    existing, err := s.queryRepo.FindByID(ctx, id)
    if err != nil { return errors.InternalError().Message("Failed to retrieve ${clean_name} before deletion").Cause(err).Build() }
    if existing == nil { return nil } // Idempotent: jika tidak ada, anggap berhasil

    if err := s.cmdRepo.Delete(ctx, id); err != nil {
        return errors.InternalError().Message("Failed to delete ${clean_name}").Cause(err).Build()
    }

    if s.cache != nil {
        _ = s.cache.Delete(ctx, "${snake_case_name}_search_v2:*")
    }

    return nil
}
EOF
    fi

    # --- Generate service_test.go (Boilerplate Unit Test & Mocks) ---
    local test_file="${target_dir}/service_test.go"
    if [ -f "$test_file" ]; then
        log_warning "File already exists, skipping: $test_file"
    else
        log_info "Generating unit test and mock files..."
        cat > "$test_file" << EOF
package ${package_name}

import (
	"context"
	"testing"

	"service/pkg/utils/query"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// --- Mock Command Repository ---
type MockCommandRepository struct {
	mock.Mock
}

func (m *MockCommandRepository) Create(ctx context.Context, entity *${clean_name}) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockCommandRepository) Update(ctx context.Context, entity *${clean_name}) error {
	args := m.Called(ctx, entity)
	return args.Error(0)
}

func (m *MockCommandRepository) Delete(ctx context.Context, id ${GO_PK_TYPE}) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// --- Mock Query Repository ---
type MockQueryRepository struct {
	mock.Mock
}

func (m *MockQueryRepository) FindAll(ctx context.Context, limit, offset int) ([]${clean_name}, int64, error) {
	args := m.Called(ctx, limit, offset)
	return args.Get(0).([]${clean_name}), args.Get(1).(int64), args.Error(2)
}

func (m *MockQueryRepository) FindByID(ctx context.Context, id ${GO_PK_TYPE}) (*${clean_name}, error) {
	args := m.Called(ctx, id)
	if args.Get(0) != nil {
		return args.Get(0).(*${clean_name}), args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *MockQueryRepository) Search(ctx context.Context, filters map[string]interface{}, sorts []query.SortField, limit, offset int) ([]${clean_name}, int64, error) {
	args := m.Called(ctx, filters, sorts, limit, offset)
	return args.Get(0).([]${clean_name}), args.Get(1).(int64), args.Error(2)
}

// --- Test Suites ---
func TestGetDetail_Success(t *testing.T) {
	mockCmdRepo := new(MockCommandRepository)
	mockQueryRepo := new(MockQueryRepository)
	svc := NewService(mockCmdRepo, mockQueryRepo, nil)

    var dummyId ${GO_PK_TYPE}
    var notFoundId ${GO_PK_TYPE}
	expectedData := &${clean_name}{
		${GO_PK_NAME}: dummyId,
	}

	mockQueryRepo.On("FindByID", mock.Anything, dummyId).Return(expectedData, nil)

	result, err := svc.GetDetail(context.Background(), dummyId)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, dummyId, result.${GO_PK_NAME})

	mockQueryRepo.AssertExpectations(t)
}

func TestGetDetail_NotFound(t *testing.T) {
	mockCmdRepo := new(MockCommandRepository)
	mockQueryRepo := new(MockQueryRepository)
	svc := NewService(mockCmdRepo, mockQueryRepo, nil)

	mockQueryRepo.On("FindByID", mock.Anything, notFoundId).Return(nil, nil)

	result, err := svc.GetDetail(context.Background(), notFoundId)

	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "not found")

	mockQueryRepo.AssertExpectations(t)
}

func TestGetDetail_InvalidID(t *testing.T) {
	mockCmdRepo := new(MockCommandRepository)
	mockQueryRepo := new(MockQueryRepository)
	svc := NewService(mockCmdRepo, mockQueryRepo, nil)

	mockQueryRepo.AssertNotCalled(t, "FindByID")

}
EOF
    fi

    log_success "Generated domain files for $table_name"
}

# Generate handler file
generate_handler_file() {
    local table_name="$1"
    local clean_name=$(to_pascal_case "$table_name") # Misal: "Language"
    
    local package_name=$(extract_package_name "$CUSTOM_DIR") # Misal: "reference"
    
    local target_dir="${INTERNAL_DIR}/infrastructure/transport/http/handlers/${CUSTOM_DIR}"
    
    log_info "Generating handler file in: $target_dir"
    mkdir -p "$target_dir"

    # Konversi 'clean_name' (PascalCase) ke 'snake_case' untuk nama file
    local snake_case_name=$(to_snake_case "$clean_name")
    local handler_file_name="${target_dir}/${package_name}_handler.go"

    if [ -f "$handler_file_name" ]; then
        log_warning "File already exists, skipping: $handler_file_name"
        return
    fi

    local id_parse_snippet=""
    if [[ "$GO_PK_TYPE" == "string" ]]; then
        id_parse_snippet="id := c.Param(\"id\")"
    else
        id_parse_snippet="id, err := strconv.ParseInt(c.Param(\"id\"), 10, 64)
    if err != nil {
        response.Error(c, http.StatusBadRequest, \"Invalid ID format\", nil)
        return
    }"
    fi

    log_info "Generating file: $handler_file_name"
    cat > "$handler_file_name" << EOF
package handlers

import (
    "fmt"
    "math"
    "net/http"
    "strconv"
    "strings"

    ${package_name}Service "service/internal/${CUSTOM_DIR}"
    "service/pkg/errors"
    "service/pkg/logger"
    "service/pkg/response"

    "github.com/gin-gonic/gin"
)

type ${clean_name}Handler struct {
    service ${package_name}Service.Service
}

func New${clean_name}Handler(service ${package_name}Service.Service) *${clean_name}Handler {
    return &${clean_name}Handler{service: service}
}

func (h *${clean_name}Handler) RegisterRoutes(router *gin.RouterGroup) {
    group := router.Group("/${snake_case_name}s")
    {
        group.GET("", h.GetList)
        group.GET("/search", h.Search)
        group.GET("/:id", h.GetDetail)
        group.POST("", h.Create)
        group.PUT("/:id", h.Update)
        group.DELETE("/:id", h.Delete)
    }
}

// GetList godoc
// @Summary      Get list of ${clean_name}s
// @Description  Retrieve a paginated list of ${clean_name}
// @Tags         ${snake_case_name}s
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        page_size query int false "Number of items per page" default(10)
// @Param        sort query string false "Sort fields (e.g. +name,-created_at)"
${active_param_doc}
// @Success      200  {object}  response.Response
// @Security     BearerAuth
// @Router       /${snake_case_name}s [get]
func (h *${clean_name}Handler) GetList(c *gin.Context) {
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "10"))

${active_parsing}

    // Parse parameter sort (format: sort=column1,-column2,+column3)
    // -column untuk DESC, +column atau column untuk ASC
    var sorts []string
    if sortParam := c.Query("sort"); sortParam != "" {
        sorts = strings.Split(sortParam, ",")
        // Validasi dan bersihkan sort parameters
        for i, sort := range sorts {
            sorts[i] = strings.TrimSpace(sort)
        }
    }

    ctx := c.Request.Context()
    ${get_list_call}
    if err != nil {
        appErr := errors.FromError(err)
        response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
        return
    }

    data := result["data"]
    total := result["total"].(int64)
    totalPages := int(math.Ceil(float64(total) / float64(pageSize)))

    meta := response.Meta{Page: page, Limit: pageSize, Total: int(total), TotalPages: totalPages}
    response.Paginated(c, http.StatusOK, "Successfully retrieved ${clean_name} list", data, meta)
}

// GetDetail godoc
// @Summary      Get ${clean_name} detail
// @Description  Retrieve detailed information about a specific ${clean_name}
// @Tags         ${snake_case_name}s
// @Produce      json
// @Param        id   path      int  true  "${clean_name} ID"
// @Success      200  {object}  response.Response
// @Security     BearerAuth
// @Router       /${snake_case_name}s/{id} [get]
func (h *${clean_name}Handler) GetDetail(c *gin.Context) {
    ${id_parse_snippet}

    ctx := c.Request.Context()
    result, err := h.service.GetDetail(ctx, id)
    if err != nil {
        appErr := errors.FromError(err)
        response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
        return
    }

    response.Success(c, http.StatusOK, "Successfully retrieved ${clean_name} detail", result)
}

// Search godoc
// @Summary      Search ${clean_name}s
// @Description  Search ${clean_name} records using dynamic filters
// @Tags         ${snake_case_name}s
// @Produce      json
// @Param        page query int false "Page number" default(1)
// @Param        limit query int false "Limit per page" default(10)
// @Success      200  {object}  response.Response
// @Security     BearerAuth
// @Router       /${snake_case_name}s/search [get]
func (h *${clean_name}Handler) Search(c *gin.Context) {
    // Parse Query Params dengan default value
    page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
    limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))

    // Ambil parameter filter
    // Ambil parameter filter secara dinamis
    filters := make(map[string]interface{})

    for key, values := range c.Request.URL.Query() {
        if key == "page" || key == "limit" || key == "page_size" || key == "sort" {
            continue
        }
        if len(values) > 0 && values[0] != "" {
            filters[key] = values[0]
        }
    }
EOF
    cat >> "$handler_file_name" << EOF

    // Parse parameter sort (format: sort=column1,-column2,+column3)
    // -column untuk DESC, +column atau column untuk ASC
    var sorts []string
    if sortParam := c.Query("sort"); sortParam != "" {
        sorts = strings.Split(sortParam, ",")
        // Validasi dan bersihkan sort parameters
        for i, sort := range sorts {
            sorts[i] = strings.TrimSpace(sort)
        }
    }

    logger.Default().Info("Search request",
        logger.String("filters", fmt.Sprintf("%v", filters)),
        logger.String("sorts", fmt.Sprintf("%v", sorts)),
        logger.Int("page", page),
        logger.Int("limit", limit))

    ctx := c.Request.Context()

    // Panggil service dengan parameter sort tambahan
    result, err := h.service.Search(ctx, filters, sorts, page, limit)
    if err != nil {
        appErr := errors.FromError(err)
        response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
        return
    }

    // Extract data dari map service untuk response format
    data := result["data"]
    total := result["total"].(int64)

    // Hitung total pages
    totalPages := 0
    if limit > 0 {
        totalPages = int(math.Ceil(float64(total) / float64(limit)))
    }

    meta := response.Meta{
        Page:       page,
        Limit:      limit,
        Total:      int(total),
        TotalPages: totalPages,
    }

    response.Paginated(c, http.StatusOK, "Successfully retrieved ${clean_name} search results", data, meta)
}

// Create godoc
// @Summary      Create new ${clean_name}
// @Description  Create a new ${clean_name} record
// @Tags         ${snake_case_name}s
// @Accept       json
// @Produce      json
// @Param        request body ${package_name}Service.${clean_name}Request true "Payload"
// @Success      201  {object}  response.Response
// @Security     BearerAuth
// @Router       /${snake_case_name}s [post]
func (h *${clean_name}Handler) Create(c *gin.Context) {
    var req ${package_name}Service.${clean_name}Request
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
        return
    }

    ctx := c.Request.Context()
    created, err := h.service.Create(ctx, req)
    if err != nil {
        appErr := errors.FromError(err)
        response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
        return
    }

    response.Success(c, http.StatusCreated, "Successfully created ${clean_name}", created)
}

// Update godoc
// @Summary      Update an existing ${clean_name}
// @Description  Update details of an existing ${clean_name} record by ID
// @Tags         ${snake_case_name}s
// @Accept       json
// @Produce      json
// @Param        id   path      int  true  "${clean_name} ID"
// @Param        request body ${package_name}Service.${clean_name}Request true "Payload"
// @Success      200  {object}  response.Response
// @Security     BearerAuth
// @Router       /${snake_case_name}s/{id} [put]
func (h *${clean_name}Handler) Update(c *gin.Context) {
    ${id_parse_snippet}

    var req ${package_name}Service.${clean_name}Request
    if err := c.ShouldBindJSON(&req); err != nil {
        response.Error(c, http.StatusBadRequest, "Invalid request body", err.Error())
        return
    }

    ctx := c.Request.Context()
    updated, err := h.service.Update(ctx, id, req)
    if err != nil {
        appErr := errors.FromError(err)
        response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
        return
    }

    response.Success(c, http.StatusOK, "Successfully updated ${clean_name}", updated)
}

// Delete godoc
// @Summary      Delete a ${clean_name}
// @Description  Delete a ${clean_name} record by ID (soft delete)
// @Tags         ${snake_case_name}s
// @Produce      json
// @Param        id   path      int  true  "${clean_name} ID"
// @Success      200  {object}  response.Response
// @Security     BearerAuth
// @Router       /${snake_case_name}s/{id} [delete]
func (h *${clean_name}Handler) Delete(c *gin.Context) {
    ${id_parse_snippet}

    ctx := c.Request.Context()
    if err := h.service.Delete(ctx, id); err != nil {
        appErr := errors.FromError(err)
        response.Error(c, appErr.HTTPStatus(), appErr.Error(), appErr.Metadata())
        return
    }

    response.Success(c, http.StatusOK, "Successfully deleted ${clean_name}", nil)
}
EOF
    log_success "Generated handler file: ${snake_case_name}_handler.go"
}

# Generate gRPC proto file
generate_proto_file() {
    local table_name="$1"
    local clean_name=$(to_pascal_case "$table_name") # e.g., Page
    local package_name=$(extract_package_name "$CUSTOM_DIR") # e.g., master

    # Create proto directory structure
    local proto_dir="${INTERNAL_DIR}/infrastructure/transport/grpc/proto/${CUSTOM_DIR}/v1"
    log_info "Generating proto file in: $proto_dir"
    mkdir -p "$proto_dir"

    local proto_file="${proto_dir}/${package_name}.proto"

    if [ -f "$proto_file" ]; then
        log_warning "File already exists, skipping: $proto_file"
        return
    fi

    log_info "Generating file: $proto_file"
    # Start writing the proto file
    cat > "$proto_file" << EOF
syntax = "proto3";

package ${package_name}.v1;

// Path should be relative to the module root.
// The package name is specified after the semicolon.
option go_package = "internal/infrastructure/transport/grpc/gen/${CUSTOM_DIR}/v1;${package_name}V1";

import "google/protobuf/wrappers.proto";
import "google/protobuf/timestamp.proto";
import "google/protobuf/empty.proto";

// Service definition for ${clean_name}.
service ${clean_name}Service {
  // Get a single ${clean_name} by its ID.
  rpc Get${clean_name}(Get${clean_name}Request) returns (${clean_name}Response);

  // Get a list of ${clean_name}s with pagination and filtering.
  rpc List${clean_name}s(List${clean_name}sRequest) returns (List${clean_name}sResponse);

  // Create a new ${clean_name}.
  rpc Create${clean_name}(Create${clean_name}Request) returns (${clean_name}Response);

  // Update an existing ${clean_name}.
  rpc Update${clean_name}(Update${clean_name}Request) returns (${clean_name}Response);

  // Delete a ${clean_name} by its ID.
  rpc Delete${clean_name}(Delete${clean_name}Request) returns (google.protobuf.Empty);
}

// The main message representing a ${clean_name}.
message ${clean_name} {
EOF

    # Add fields from parsed SQL
    local field_index=1
    for col in "${PARSED_COLUMNS[@]}"; do
        IFS='|' read -r col_name col_type nullable <<< "$col"

        # Convert column name to snake_case for proto convention
        local proto_field_name=$(to_snake_case "$col_name")
        local proto_type=$(sql_to_proto_type "$col_type" "$nullable")

        echo "  ${proto_type} ${proto_field_name} = ${field_index};" >> "$proto_file"
        ((field_index++))
    done

    cat >> "$proto_file" << EOF
}

// --- Request/Response Messages ---

message Get${clean_name}Request {
  ${PROTO_PK_TYPE} id = 1;
}

message ${clean_name}Response {
  ${clean_name} data = 1;
}

message List${clean_name}sRequest {
  int32 page = 1;
  int32 page_size = 2;
  // Add filter fields here if needed
}

message List${clean_name}sResponse {
  repeated ${clean_name} data = 1;
  int64 total_items = 2;
}

message Create${clean_name}Request {
EOF

    # Add fields for Create request
    local create_field_index=1
    for col in "${PARSED_COLUMNS[@]}"; do
        IFS='|' read -r col_name col_type nullable <<< "$col"
        # Skip auto-managed fields
        if [[ "$col_name" == "Id" || "$col_name" == "CreatedAt" || "$col_name" == "UpdatedAt" || "$col_name" == "DeletedAt" ]]; then
            continue
        fi
        local proto_field_name=$(to_snake_case "$col_name")
        # For create, we don't use optional, we just use the base type
        local proto_type=$(sql_to_proto_type "$col_type" "false") # Treat as non-nullable for create
        echo "  ${proto_type} ${proto_field_name} = ${create_field_index};" >> "$proto_file"
        ((create_field_index++))
    done

    cat >> "$proto_file" << EOF
}

message Update${clean_name}Request {
  int64 id = 1;
EOF

    # Add fields for Update request, all should be optional
    local update_field_index=2
    for col in "${PARSED_COLUMNS[@]}"; do
        IFS='|' read -r col_name col_type nullable <<< "$col"
        if is_managed_field "$col_name"; then
            continue
        fi
        local proto_field_name=$(to_snake_case "$col_name")
        # For update, all fields are optional
        local proto_type=$(sql_to_proto_type "$col_type" "true")
        echo "  ${proto_type} ${proto_field_name} = ${update_field_index};" >> "$proto_file"
        ((update_field_index++))
    done

    cat >> "$proto_file" << EOF
}

message Delete${clean_name}Request {
  ${PROTO_PK_TYPE} id = 1;
}
EOF

    log_success "Generated proto file: ${package_name}.proto"
    
    log_info "Running proto.sh to generate Go code from this new .proto file..."
    local rel_proto_dir="internal/infrastructure/transport/grpc/proto/${CUSTOM_DIR}/v1"
    if [ -x "${SCRIPT_DIR}/proto.sh" ]; then
        "${SCRIPT_DIR}/proto.sh" "$rel_proto_dir"
    elif [ -f "${SCRIPT_DIR}/proto.sh" ]; then
        bash "${SCRIPT_DIR}/proto.sh" "$rel_proto_dir"
    else
        log_warning "proto.sh not found. Please run it manually."
    fi
}

# Generate gRPC handler and mapper file
generate_grpc_handler() {
    local table_name="$1"
    local clean_name=$(to_pascal_case "$table_name") # e.g., Page
    local package_name=$(extract_package_name "$CUSTOM_DIR") # e.g., master

    local target_dir="${INTERNAL_DIR}/infrastructure/transport/grpc/handlers/${CUSTOM_DIR}"
    log_info "Generating gRPC handler in: $target_dir"
    mkdir -p "$target_dir"

    local handler_file_name="${target_dir}/${package_name}_grpc_handler.go"

    if [ -f "$handler_file_name" ]; then
        log_warning "File already exists, skipping: $handler_file_name"
    else
        log_info "Generating file: $handler_file_name"
        cat > "$handler_file_name" << EOF
package handlers

import (
    "context"
    "fmt"

    gen${clean_name} "service/internal/infrastructure/transport/grpc/gen/${CUSTOM_DIR}/v1"
    ${package_name}Service "service/internal/${CUSTOM_DIR}"
    "service/pkg/errors"

    "google.golang.org/grpc/codes"
    "google.golang.org/grpc/status"
    "google.golang.org/protobuf/types/known/emptypb"
)

type ${clean_name}Handler struct {
    gen${clean_name}.Unimplemented${clean_name}ServiceServer
    service ${package_name}Service.Service
}

func New${clean_name}Handler(service ${package_name}Service.Service) *${clean_name}Handler {
    return &${clean_name}Handler{service: service}
}

func (h *${clean_name}Handler) Get${clean_name}(ctx context.Context, req *gen${clean_name}.Get${clean_name}Request) (*gen${clean_name}.${clean_name}Response, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
    }

    data, err := h.service.GetDetail(ctx, req.GetId())
    if err != nil {
        if errors.Is(err, errors.ErrNotFound) {
            return nil, status.Error(codes.NotFound, "${clean_name} not found")
        }
        return nil, status.Error(codes.Internal, fmt.Sprintf("failed to get ${clean_name}: %v", err))
    }

    return &gen${clean_name}.${clean_name}Response{
        Data: Map${clean_name}ResponseToProto(data),
    }, nil
}

func (h *${clean_name}Handler) List${clean_name}s(ctx context.Context, req *gen${clean_name}.List${clean_name}sRequest) (*gen${clean_name}.List${clean_name}sResponse, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
    }

    page := int(req.GetPage())
    pageSize := int(req.GetPageSize())

    result, err := h.service.GetList(ctx, page, pageSize, nil)
    if err != nil {
        return nil, status.Error(codes.Internal, fmt.Sprintf("failed to list ${clean_name}s: %v", err))
    }

    data := result["data"].([]*${package_name}Service.${clean_name}Response)
    total := result["total"].(int64)

    protoData := make([]*gen${clean_name}.${clean_name}, len(data))
    for i, e := range data {
        protoData[i] = Map${clean_name}ResponseToProto(e)
    }

    return &gen${clean_name}.List${clean_name}sResponse{
        Data:       protoData,
        TotalItems: total,
    }, nil
}

func (h *${clean_name}Handler) Create${clean_name}(ctx context.Context, req *gen${clean_name}.Create${clean_name}Request) (*gen${clean_name}.${clean_name}Response, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
    }

    createReq := MapProtoTo${clean_name}Request(req)
    created, err := h.service.Create(ctx, createReq)
    if err != nil {
        return nil, status.Error(codes.Internal, fmt.Sprintf("failed to create ${clean_name}: %v", err))
    }

    return &gen${clean_name}.${clean_name}Response{
        Data: Map${clean_name}ResponseToProto(created),
    }, nil
}

func (h *${clean_name}Handler) Update${clean_name}(ctx context.Context, req *gen${clean_name}.Update${clean_name}Request) (*gen${clean_name}.${clean_name}Response, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
    }

    updateReq := MapProtoTo${clean_name}UpdateRequest(req)
    updated, err := h.service.Update(ctx, req.GetId(), updateReq)
    if err != nil {
        return nil, status.Error(codes.Internal, fmt.Sprintf("failed to update ${clean_name}: %v", err))
    }

    return &gen${clean_name}.${clean_name}Response{
        Data: Map${clean_name}ResponseToProto(updated),
    }, nil
}

func (h *${clean_name}Handler) Delete${clean_name}(ctx context.Context, req *gen${clean_name}.Delete${clean_name}Request) (*emptypb.Empty, error) {
    if req == nil {
        return nil, status.Error(codes.InvalidArgument, "request cannot be nil")
    }

    if err := h.service.Delete(ctx, req.GetId()); err != nil {
        return nil, status.Error(codes.Internal, fmt.Sprintf("failed to delete ${clean_name}: %v", err))
    }

    return &emptypb.Empty{}, nil
}
EOF
        log_success "Generated gRPC handler file: ${package_name}_grpc_handler.go"
    fi

    # Generate Mapper
    local mapper_file_name="${target_dir}/${package_name}_grpc_mapper.go"
    if [ -f "$mapper_file_name" ]; then
        log_warning "File already exists, skipping: $mapper_file_name"
    else
        log_info "Generating file: $mapper_file_name"
        cat > "$mapper_file_name" << EOF
package handlers

import (
    gen${clean_name} "service/internal/infrastructure/transport/grpc/gen/${CUSTOM_DIR}/v1"
    ${package_name}Service "service/internal/${CUSTOM_DIR}"
)

func Map${clean_name}ResponseToProto(e *${package_name}Service.${clean_name}Response) *gen${clean_name}.${clean_name} {
    if e == nil {
        return nil
    }
    return &gen${clean_name}.${clean_name}{
EOF
        for col in "${PARSED_COLUMNS[@]}"; do
            IFS='|' read -r col_name _ _ <<< "$col"
            go_field_name=$(to_pascal_case "$col_name")
            echo "        ${go_field_name}: e.${go_field_name}," >> "$mapper_file_name"
        done
        cat >> "$mapper_file_name" << EOF
    }
}

func MapProtoTo${clean_name}Request(req *gen${clean_name}.Create${clean_name}Request) ${package_name}Service.${clean_name}Request {
    return ${package_name}Service.${clean_name}Request{
EOF
        for col in "${PARSED_COLUMNS[@]}"; do
            IFS='|' read -r col_name _ _ <<< "$col"
            if is_managed_field "$col_name"; then
                continue
            fi
            go_field_name=$(to_pascal_case "$col_name")
            echo "        ${go_field_name}: req.${go_field_name}," >> "$mapper_file_name"
        done
        cat >> "$mapper_file_name" << EOF
    }
}

func MapProtoTo${clean_name}UpdateRequest(req *gen${clean_name}.Update${clean_name}Request) ${package_name}Service.${clean_name}Request {
    return ${package_name}Service.${clean_name}Request{
EOF
        for col in "${PARSED_COLUMNS[@]}"; do
            IFS='|' read -r col_name _ _ <<< "$col"
            if is_managed_field "$col_name"; then
                continue
            fi
            go_field_name=$(to_pascal_case "$col_name")
            echo "        ${go_field_name}: req.${go_field_name}," >> "$mapper_file_name"
        done
        cat >> "$mapper_file_name" << EOF
    }
}
EOF
        log_success "Generated gRPC mapper file: ${package_name}_grpc_mapper.go"
    fi
}

# --- Main Execution ---

# Helper to create output directory for docs
setup_output_dir() {
    if [[ ! -d "$OUTPUT_DIR" ]]; then
        log_verbose "Creating output directory: $OUTPUT_DIR"
        mkdir -p "$OUTPUT_DIR"
    fi
}

# Main execution function
main() {
    log_header "Advanced Context Generator"
    parse_args "$@"
    setup_output_dir

    # Parse SQL file to get table info
    declare -a PARSED_COLUMNS
    declare PARSED_TABLE_NAME
    declare PARSED_PRIMARY_KEY

    if [[ -n "$SQL_FILE" ]]; then
        if ! parse_sql_table "$SQL_FILE"; then exit 1; fi
    elif [[ -n "$JSON_FILE" ]]; then
        if ! parse_json_payload "$JSON_FILE"; then exit 1; fi
    fi

    log_info "Generating for table: $PARSED_TABLE_NAME"
    log_verbose "Target directory: ${INTERNAL_DIR}/${CUSTOM_DIR}"
    log_verbose "Generate type: $GENERATE_TYPE"

    case "$GENERATE_TYPE" in
        "domain")
            generate_domain_files "$PARSED_TABLE_NAME"
            ;;
        "handler")
            generate_handler_file "$PARSED_TABLE_NAME"
            ;;
        "proto")
            generate_proto_file "$PARSED_TABLE_NAME"
            ;;
        "grpc")
            generate_grpc_handler "$PARSED_TABLE_NAME"
            ;;
        "all")
            generate_domain_files "$PARSED_TABLE_NAME"
            generate_handler_file "$PARSED_TABLE_NAME"
            generate_proto_file "$PARSED_TABLE_NAME"
            generate_grpc_handler "$PARSED_TABLE_NAME"
            ;;
    esac

    local package_name=$(extract_package_name "$CUSTOM_DIR")
    local clean_name=$(to_pascal_case "$PARSED_TABLE_NAME")

    echo -e "\n${PURPLE}================================================================${NC}"
    echo -e "${GREEN}✨ MODULE [${clean_name}] GENERATED SUCCESSFULLY! ✨${NC}"
    echo -e "${PURPLE}================================================================${NC}\n"

    echo -e "${YELLOW}🚀 NEXT STEPS TO ACTIVATE YOUR MODULE:${NC}\n"

    echo -e "${CYAN}1. Inisialisasi Repository & Service (misal di cmd/api/main.go):${NC}"
    echo -e "   Pastikan Anda meng-import package: ${WHITE}service/internal/${CUSTOM_DIR}${NC}"
    echo -e "   ${WHITE}${package_name}CmdRepo := ${package_name}.NewCommandRepository(dbService, \"default\")${NC}"
    echo -e "   ${WHITE}${package_name}QueryRepo := ${package_name}.NewQueryRepository(dbService, \"default\")${NC}"
    echo -e "   ${WHITE}${package_name}Svc := ${package_name}.NewService(${package_name}CmdRepo, ${package_name}QueryRepo, cacheManager)${NC}\n"

    if [[ "$GENERATE_TYPE" == "all" || "$GENERATE_TYPE" == "handler" ]]; then
        echo -e "${CYAN}2. Registrasi REST API Handler (ke Router Gin):${NC}"
        echo -e "   Pastikan Anda meng-import package: ${WHITE}service/internal/infrastructure/transport/http/handlers/${CUSTOM_DIR}${NC}"
        echo -e "   ${WHITE}${package_name}Handler := handlers.New${clean_name}Handler(${package_name}Svc)${NC}"
        echo -e "   ${WHITE}${package_name}Handler.RegisterRoutes(apiGroup)${NC}\n"
    fi

    if [[ "$GENERATE_TYPE" == "all" || "$GENERATE_TYPE" == "grpc" || "$GENERATE_TYPE" == "proto" ]]; then
        echo -e "${CYAN}3. Registrasi gRPC Handler di server.go:${NC}"
        echo -e "   Pastikan Anda meng-import package: ${WHITE}service/internal/infrastructure/transport/grpc/handlers/${CUSTOM_DIR}${NC}"
        echo -e "   ${WHITE}grpc${clean_name}Handler := grpcHandlers.New${clean_name}Handler(${package_name}Svc)${NC}"
        echo -e "   ${WHITE}gen${clean_name}.Register${clean_name}ServiceServer(srv, grpc${clean_name}Handler)${NC}\n"
    fi

    echo -e "${CYAN}4. Rapihkan Dependencies & Tinjau Kode:${NC}"
    echo -e "   ${WHITE}go mod tidy${NC}\n"
}

# Run main function with all arguments
main "$@"