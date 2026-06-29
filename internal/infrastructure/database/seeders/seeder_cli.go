package seeders

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"strings"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// SeederCLI adalah command-line interface untuk seeder
type SeederCLI struct {
	db       *gorm.DB
	registry *SeederRegistry
}

// NewSeederCLI membuat CLI baru
func NewSeederCLI(db *gorm.DB) *SeederCLI {
	return &SeederCLI{
		db:       db,
		registry: DefaultRegistry(),
	}
}

// Run menjalankan CLI seeder
func (c *SeederCLI) Run(args []string) error {
	if len(args) == 0 {
		return c.showHelp()
	}

	command := args[0]
	switch command {
	case "list":
		return c.listTables()
	case "seed":
		if len(args) < 2 {
			return fmt.Errorf("usage: seeder seed <table-name|all> [options]")
		}
		return c.seedTable(args[1], args[2:])
	case "dry-run":
		if len(args) < 2 {
			return fmt.Errorf("usage: seeder dry-run <table-name|all> [options]")
		}
		return c.dryRunTable(args[1], args[2:])
	case "validate":
		if len(args) < 2 {
			return fmt.Errorf("usage: seeder validate <table-name|all>")
		}
		return c.validateTable(args[1])
	default:
		return fmt.Errorf("unknown command: %s", command)
	}
}

// showHelp menampilkan bantuan
func (c *SeederCLI) showHelp() error {
	help := `
Seeder CLI - Flexible Database Seeder

Usage:
  seeder <command> [arguments]

Commands:
  list                    Show all available tables
  seed <table|all>        Seed data for specific table or all tables
  dry-run <table|all>     Preview seeding without actual insertion
  validate <table|all>    Validate CSV files and configuration

Examples:
  seeder list
  seeder seed province
  seeder seed all
  seeder dry-run ethnic
  seeder validate regency

Options:
  -batch-size=N          Set batch size for insertion (default: from config)
  -delete-before         Delete existing data before seeding
  -no-delete             Don't delete existing data (default)
  -csv-path=PATH         Override CSV file path
  -table-name=NAME       Override table name

Notes:
  - CSV files should be in internal/infrastructure/database/csv/
  - First row of CSV should be header with column names
  - Column names in CSV will be mapped to struct fields
`
	fmt.Println(help)
	return nil
}

// listTables menampilkan daftar tabel yang tersedia
func (c *SeederCLI) listTables() error {
	tables := c.registry.List()

	fmt.Println("Available tables for seeding:")
	fmt.Println(strings.Repeat("-", 50))

	for _, table := range tables {
		config, _ := c.registry.Get(table)
		fmt.Printf("%-20s %s\n", table, config.CSVFile)
	}

	fmt.Printf("\nTotal: %d tables\n", len(tables))
	return nil
}

// seedTable melakukan seeding untuk tabel tertentu
func (c *SeederCLI) seedTable(tableName string, options []string) error {
	if tableName == "all" {
		return c.seedAllTables(options)
	}

	config, exists := c.registry.Get(tableName)
	if !exists {
		return fmt.Errorf("table '%s' not found in registry", tableName)
	}

	// Parse options
	opts := c.parseOptions(options)

	// Override config dengan options
	if opts.CSVPath != "" {
		config.CSVFile = opts.CSVPath
	}
	if opts.TableName != "" {
		config.TableName = opts.TableName
	}
	if opts.BatchSize > 0 {
		config.BatchSize = opts.BatchSize
	}
	if opts.DeleteBefore {
		config.DeleteBefore = true
	}

	// Validasi config
	if err := config.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get entity
	entity := GetEntityByTableName(config.TableName)
	if entity == nil {
		return fmt.Errorf("entity not found for table '%s'", config.TableName)
	}

	// Create seeder
	seederConfig := SeederConfig{
		CSVPath:      config.CSVFile,
		TableName:    config.TableName,
		ColumnMap:    config.ColumnMap,
		SkipHeader:   true,
		BatchSize:    config.BatchSize,
		DeleteBefore: config.DeleteBefore,
		DryRun:       false,
	}

	seeder := NewMasterSeeder(c.db, seederConfig)

	log.Printf("Starting seed for table: %s", config.TableName)
	log.Printf("CSV file: %s", config.CSVFile)
	log.Printf("Batch size: %d", config.BatchSize)
	log.Printf("Delete before: %v", config.DeleteBefore)

	if err := seeder.SeedFromCSV(entity); err != nil {
		return fmt.Errorf("seeding failed: %w", err)
	}

	log.Printf("Successfully seeded table: %s", config.TableName)
	return nil
}

// dryRunTable melakukan dry-run untuk tabel tertentu
func (c *SeederCLI) dryRunTable(tableName string, options []string) error {
	if tableName == "all" {
		return c.dryRunAllTables(options)
	}

	config, exists := c.registry.Get(tableName)
	if !exists {
		return fmt.Errorf("table '%s' not found in registry", tableName)
	}

	// Parse options
	opts := c.parseOptions(options)

	// Override config dengan options
	if opts.CSVPath != "" {
		config.CSVFile = opts.CSVPath
	}
	if opts.TableName != "" {
		config.TableName = opts.TableName
	}

	// Validasi config
	if err := config.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get entity
	entity := GetEntityByTableName(config.TableName)
	if entity == nil {
		return fmt.Errorf("entity not found for table '%s'", config.TableName)
	}

	// Create seeder dengan dry-run mode
	seederConfig := SeederConfig{
		CSVPath:      config.CSVFile,
		TableName:    config.TableName,
		ColumnMap:    config.ColumnMap,
		SkipHeader:   true,
		BatchSize:    10, // Smaller batch for dry-run
		DeleteBefore: false,
		DryRun:       true,
	}

	seeder := NewMasterSeeder(c.db, seederConfig)

	log.Printf("Starting dry-run for table: %s", config.TableName)
	log.Printf("CSV file: %s", config.CSVFile)

	if err := seeder.SeedFromCSV(entity); err != nil {
		return fmt.Errorf("dry-run failed: %w", err)
	}

	log.Printf("Dry-run completed for table: %s", config.TableName)
	return nil
}

// validateTable memvalidasi konfigurasi dan CSV file
func (c *SeederCLI) validateTable(tableName string) error {
	if tableName == "all" {
		return c.validateAllTables()
	}

	config, exists := c.registry.Get(tableName)
	if !exists {
		return fmt.Errorf("table '%s' not found in registry", tableName)
	}

	// Validasi config
	if err := config.Validate(); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	// Get entity
	entity := GetEntityByTableName(config.TableName)
	if entity == nil {
		return fmt.Errorf("entity not found for table '%s'", config.TableName)
	}

	// Baca CSV file dan validasi
	file, err := os.Open(config.CSVFile)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	// Hitung jumlah baris
	rowCount := 0
	for {
		_, err := reader.Read()
		if err != nil {
			break
		}
		rowCount++
	}

	fmt.Printf("Validation passed for table: %s\n", tableName)
	fmt.Printf("CSV file: %s\n", config.CSVFile)
	fmt.Printf("Headers: %v\n", headers)
	fmt.Printf("Expected rows: %d\n", rowCount)
	fmt.Printf("Column mapping: %v\n", config.ColumnMap)
	fmt.Printf("Entity type: %T\n", entity)

	return nil
}

// seedAllTables melakukan seeding untuk semua tabel
func (c *SeederCLI) seedAllTables(options []string) error {
	tables := c.registry.List()

	log.Printf("Starting seed for all tables (%d tables)", len(tables))

	for _, table := range tables {
		if err := c.seedTable(table, options); err != nil {
			log.Printf("Error seeding table %s: %v", table, err)
			continue
		}
	}

	log.Printf("Completed seeding all tables")
	return nil
}

// dryRunAllTables melakukan dry-run untuk semua tabel
func (c *SeederCLI) dryRunAllTables(options []string) error {
	tables := c.registry.List()

	log.Printf("Starting dry-run for all tables (%d tables)", len(tables))

	for _, table := range tables {
		if err := c.dryRunTable(table, options); err != nil {
			log.Printf("Error in dry-run for table %s: %v", table, err)
			continue
		}
	}

	log.Printf("Completed dry-run for all tables")
	return nil
}

// validateAllTables memvalidasi semua tabel
func (c *SeederCLI) validateAllTables() error {
	tables := c.registry.List()

	fmt.Printf("Validating all tables (%d tables):\n", len(tables))
	fmt.Println(strings.Repeat("=", 60))

	for _, table := range tables {
		fmt.Printf("\n[%s]\n", table)
		if err := c.validateTable(table); err != nil {
			fmt.Printf("❌ Validation failed: %v\n", err)
		} else {
			fmt.Printf("✅ Validation passed\n")
		}
	}

	return nil
}

// SeederOptions berisi parsed options
type SeederOptions struct {
	CSVPath      string
	TableName    string
	BatchSize    int
	DeleteBefore bool
	DryRun       bool
}

// parseOptions mem-parsing command line options
func (c *SeederCLI) parseOptions(options []string) SeederOptions {
	var opts SeederOptions

	for _, option := range options {
		switch {
		case strings.HasPrefix(option, "-batch-size="):
			fmt.Sscanf(option, "-batch-size=%d", &opts.BatchSize)
		case option == "-delete-before":
			opts.DeleteBefore = true
		case strings.HasPrefix(option, "-csv-path="):
			opts.CSVPath = strings.TrimPrefix(option, "-csv-path=")
		case strings.HasPrefix(option, "-table-name="):
			opts.TableName = strings.TrimPrefix(option, "-table-name=")
		}
	}

	return opts
}

// MainCLI adalah entry point utama untuk CLI
func MainCLI() {
	// Setup database connection
	db, err := setupDatabase()
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}

	cli := NewSeederCLI(db)

	if err := cli.Run(os.Args[1:]); err != nil {
		log.Fatalf("Error: %v", err)
	}
}

// setupDatabase setup koneksi database
func setupDatabase() (*gorm.DB, error) {
	// Bisa diambil dari config atau environment variable
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		dsn = "host=localhost user=postgres password=postgres dbname=health port=5432 sslmode=disable"
	}

	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, err
	}

	return db, nil
}
