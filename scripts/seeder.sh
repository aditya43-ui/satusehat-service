#!/bin/bash

# Script untuk generate seeder dari SQL migration files
# Usage: ./scripts/generate_seeder.sh [OPTIONS]
# Examples:
#   ./scripts/generate_seeder.sh --sql=internal/infrastructure/database/sql/20260127064613_create_ethnic_table.sql
#   ./scripts/generate_seeder.sh --sql-dir=internal/infrastructure/database/sql/
#   ./scripts/generate_seeder.sh --all

set -e  # Exit immediately if a command exits with a non-zero status

echo "🔧 Seeder Generator Script"

# Warna untuk output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
SQL_FILE=""
SQL_DIR=""
GENERATE_ALL=false
OUTPUT_DIR="internal/infrastructure/database/seeders"
CSV_DIR="internal/infrastructure/database/csv"
BATCH_SIZE=100
GENERATE_CSV=true
VERBOSE=false
DRY_RUN=false

# Fungsi untuk menampilkan usage
show_usage() {
    echo "Usage: $0 [OPTIONS]"
    echo ""
    echo "Generate seeder from SQL migration files"
    echo ""
    echo "Options:"
    echo "  --sql=FILE          Path to specific SQL file"
    echo "  --sql-dir=DIR       Process all SQL files in directory"
    echo "  --all               Process all SQL files in default migrations directory"
    echo "  --output=DIR        Output directory for generated seeders (default: $OUTPUT_DIR)"
    echo "  --csv-dir=DIR       CSV output directory (default: $CSV_DIR)"
    echo "  --batch-size=NUM    Batch size for seeding (default: $BATCH_SIZE)"
    echo "  --no-csv            Don't generate CSV templates"
    echo "  --verbose, -v       Verbose output"
    echo "  --dry-run           Show what would be generated without creating files"
    echo "  --help, -h          Show this help message"
    echo ""
    echo "Examples:"
    echo "  $0 --sql=internal/infrastructure/database/sql/20260127064613_create_ethnic_table.sql"
    echo "  $0 --sql-dir=internal/infrastructure/database/sql/"
    echo "  $0 --all"
    echo "  $0 --sql=sql/create_table.sql --output=custom/seeders --batch-size=50"
    echo "  $0 --all --dry-run --verbose"
}

# Fungsi untuk logging
log_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

log_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

log_warning() {
    echo -e "${YELLOW}⚠️  $1${NC}"
}

log_error() {
    echo -e "${RED}❌ $1${NC}"
}

log_verbose() {
    if [ "$VERBOSE" = true ]; then
        echo -e "${BLUE}🔍 $1${NC}"
    fi
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        --sql=*)
            SQL_FILE="${1#*=}"
            shift
            ;;
        --sql-dir=*)
            SQL_DIR="${1#*=}"
            shift
            ;;
        --all)
            GENERATE_ALL=true
            shift
            ;;
        --output=*)
            OUTPUT_DIR="${1#*=}"
            shift
            ;;
        --csv-dir=*)
            CSV_DIR="${1#*=}"
            shift
            ;;
        --batch-size=*)
            BATCH_SIZE="${1#*=}"
            shift
            ;;
        --no-csv)
            GENERATE_CSV=false
            shift
            ;;
        --verbose|-v)
            VERBOSE=true
            shift
            ;;
        --dry-run)
            DRY_RUN=true
            shift
            ;;
        --help|-h)
            show_usage
            exit 0
            ;;
        *)
            log_error "Unknown option: $1"
            show_usage
            exit 1
            ;;
    esac
done

# Validasi input
if [ "$GENERATE_ALL" = true ] && [ -n "$SQL_FILE" ]; then
    log_error "Cannot use --all with --sql option"
    exit 1
fi

if [ "$GENERATE_ALL" = true ] && [ -n "$SQL_DIR" ]; then
    log_error "Cannot use --all with --sql-dir option"
    exit 1
fi

if [ -n "$SQL_FILE" ] && [ -n "$SQL_DIR" ]; then
    log_error "Cannot use --sql with --sql-dir option"
    exit 1
fi

if [ "$GENERATE_ALL" = false ] && [ -z "$SQL_FILE" ] && [ -z "$SQL_DIR" ]; then
    log_error "Must specify either --sql, --sql-dir, or --all"
    show_usage
    exit 1
fi

# Set default SQL directory jika menggunakan --all
if [ "$GENERATE_ALL" = true ]; then
    SQL_DIR="internal/infrastructure/database/sql"
fi

# Fungsi untuk cek apakah file SQL valid
validate_sql_file() {
    local file="$1"
    
    if [ ! -f "$file" ]; then
        log_error "SQL file not found: $file"
        return 1
    fi
    
    if ! grep -q "CREATE TABLE" "$file"; then
        log_warning "File does not contain CREATE TABLE statement: $file"
        return 1
    fi
    
    return 0
}

# Fungsi untuk extract table name dari SQL file
extract_table_name() {
    local sql_file="$1"
    local filename=$(basename "$sql_file" .sql)
    
    # Pattern: YYYYMMDDHHMMSS_create_tablename.sql
    if [[ "$filename" =~ ^[0-9]{14}_create_(.+)$ ]]; then
        echo "${BASH_REMATCH[1]}"
    else
        # Fallback: extract dari CREATE TABLE statement
        grep -i "CREATE TABLE" "$sql_file" | head -1 | sed -n 's/.*CREATE TABLE[[:space:]]*["\`]\?\([^"\`[:space:]]*\)["\`]\?.*/\1/ip' | tr '[:upper:]' '[:lower:]'
    fi
}

# Fungsi untuk generate seeder dari single SQL file
generate_seeder_from_file() {
    local sql_file="$1"
    
    log_info "Processing SQL file: $sql_file"
    
    if ! validate_sql_file "$sql_file"; then
        return 1
    fi
    
    local table_name=$(extract_table_name "$sql_file")
    if [ -z "$table_name" ]; then
        log_error "Could not extract table name from: $sql_file"
        return 1
    fi
    
    log_verbose "Extracted table name: $table_name"
    
    local seeder_filename="seeder_${table_name}.go"
    local seeder_path="$OUTPUT_DIR/$seeder_filename"
    local csv_filename="${table_name}.csv"
    local csv_path="$CSV_DIR/$csv_filename"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "[DRY RUN] Would generate:"
        log_info "  Seeder: $seeder_path"
        if [ "$GENERATE_CSV" = true ]; then
            log_info "  CSV: $csv_path"
        fi
        return 0
    fi
    
    # Create directories if they don't exist
    mkdir -p "$OUTPUT_DIR"
    mkdir -p "$CSV_DIR"
    
    # Generate seeder using Go program
    log_info "Generating seeder for table: $table_name"
    
    local csv_flag=""
    if [ "$GENERATE_CSV" = true ]; then
        csv_flag="-csv-template"
    fi
    
    local verbose_flag=""
    if [ "$VERBOSE" = true ]; then
        verbose_flag="-v"
    fi
    
    if go run cmd/seeder-generator/main.go \
        -sql="$sql_file" \
        -output="$OUTPUT_DIR" \
        -csv-dir="$CSV_DIR" \
        -batch="$BATCH_SIZE" \
        $csv_flag $verbose_flag; then
        
        log_success "Generated seeder: $seeder_path"
        if [ "$GENERATE_CSV" = true ] && [ -f "$csv_path" ]; then
            log_success "Generated CSV: $csv_path"
        fi
    else
        log_error "Failed to generate seeder for: $table_name"
        return 1
    fi
}

# Fungsi untuk process semua SQL files di directory
process_sql_directory() {
    local dir="$1"
    
    if [ ! -d "$dir" ]; then
        log_error "Directory not found: $dir"
        return 1
    fi
    
    log_info "Processing SQL files in directory: $dir"
    
    local sql_files=()
    while IFS= read -r -d $'\0' file; do
        sql_files+=("$file")
    done < <(find "$dir" -name "*.sql" -type f -print0 | sort -z)
    
    if [ ${#sql_files[@]} -eq 0 ]; then
        log_warning "No SQL files found in directory: $dir"
        return 1
    fi
    
    log_info "Found ${#sql_files[@]} SQL files to process"
    
    local success_count=0
    local failed_count=0
    
    for sql_file in "${sql_files[@]}"; do
        if generate_seeder_from_file "$sql_file"; then
            ((success_count++))
        else
            ((failed_count++))
        fi
    done
    
    log_info "Processing complete: $success_count successful, $failed_count failed"
    
    if [ $failed_count -gt 0 ]; then
        return 1
    fi
    
    return 0
}

# Fungsi untuk update seeder registry
update_registry() {
    local registry_file="$OUTPUT_DIR/seeder_registry.go"
    
    log_info "Updating seeder registry: $registry_file"
    
    # TODO: Implement registry update logic
    # For now, just show instructions
    log_info "Don't forget to add generated seeders to your seeder registry!"
}

# Fungsi untuk test database connection
test_db_connection() {
    log_info "Testing database connection..."
    
    if go run scripts/test_db_connection.go; then
        log_success "Database connection successful"
        return 0
    else
        log_error "Database connection failed"
        return 1
    fi
}

# Main execution
main() {
    log_info "Starting seeder generation..."
    
    if [ "$VERBOSE" = true ]; then
        log_verbose "Configuration:"
        log_verbose "  SQL_FILE: $SQL_FILE"
        log_verbose "  SQL_DIR: $SQL_DIR"
        log_verbose "  OUTPUT_DIR: $OUTPUT_DIR"
        log_verbose "  CSV_DIR: $CSV_DIR"
        log_verbose "  BATCH_SIZE: $BATCH_SIZE"
        log_verbose "  GENERATE_CSV: $GENERATE_CSV"
        log_verbose "  DRY_RUN: $DRY_RUN"
    fi
    
    # Test database connection (optional)
    # test_db_connection || true
    
    # Process SQL files
    if [ -n "$SQL_FILE" ]; then
        if ! generate_seeder_from_file "$SQL_FILE"; then
            exit 1
        fi
    elif [ -n "$SQL_DIR" ]; then
        if ! process_sql_directory "$SQL_DIR"; then
            exit 1
        fi
    fi
    
    # Update registry
    update_registry
    
    log_success "Seeder generation completed!"
    
    if [ "$DRY_RUN" = true ]; then
        log_info "This was a dry run. No files were created."
    else
        log_info "Next steps:"
        log_info "1. Review generated seeders in: $OUTPUT_DIR"
        log_info "2. Add seeders to your registry"
        log_info "3. Run: make seeder-seed-all"
    fi
}

# Run main function
main "$@"