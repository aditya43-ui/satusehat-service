# Flexible Database Seeder

A flexible and extensible database seeder for Go applications with GORM support.

## Features

- 🎯 **Flexible Configuration**: Support for any table structure
- 📊 **CSV Import**: Import data from CSV files with customizable mapping
- 🔄 **Batch Processing**: Efficient bulk inserts with configurable batch sizes
- 🧪 **Dry Run Mode**: Preview seeding without actual database changes
- ✅ **Validation**: Validate CSV files and configurations before seeding
- 📝 **Column Mapping**: Map CSV columns to struct fields flexibly
- 🗂️ **Multiple Tables**: Support for all master data tables
- 🔍 **Progress Tracking**: Detailed logging and error reporting

## Quick Start

### Build the Seeder CLI

```bash
make seeder-build
```

### List Available Tables

```bash
make seeder-list
# or
./bin/seeder list
```

### Seed a Specific Table

```bash
make seeder-seed TABLE=province
# or
./bin/seeder seed province
```

### Seed All Tables

```bash
make seeder-seed-all
# or
./bin/seeder seed all
```

### Dry Run (Preview)

```bash
make seeder-dry-run TABLE=ethnic
# or
./bin/seeder dry-run ethnic
```

### Validate Configuration

```bash
make seeder-validate
# or
./bin/seeder validate all
```

## Advanced Usage

### Command Line Options

```bash
# Seed with custom batch size and delete existing data
./bin/seeder seed province -batch-size=200 -delete-before

# Seed with custom CSV file path
./bin/seeder seed province -csv-path=/path/to/custom/provinces.csv

# Seed with custom table name
./bin/seeder seed province -table-name=provinces_backup

# Dry run with specific options
./bin/seeder dry-run ethnic -batch-size=10
```

### Makefile Targets

```bash
# Advanced seeding with options
make seeder-advanced TABLE=province BATCH_SIZE=200 DELETE_BEFORE=1

# Build seeder
make seeder-build

# Clean seeder binary
make clean-seeder
```

## Configuration

### Table Configuration

Each table has a configuration that includes:

- **TableName**: Database table name
- **CSVFile**: Path to CSV file
- **ColumnMap**: Mapping between CSV headers and struct fields
- **DeleteBefore**: Whether to delete existing data before seeding
- **BatchSize**: Number of records to insert per batch

### Default Registry

The seeder comes with pre-configured tables:

| Table | CSV File | Description |
|-------|----------|-------------|
| province | provinces.csv | Province data |
| regency | regencies.csv | Regency/City data |
| district | districts.csv | District data |
| village | villages.csv | Village data |
| ethnic | ethnics.csv | Ethnic groups |
| language | languages.csv | Languages |
| installation | installations.csv | Hospital installations |
| unit | units.csv | Medical units |
| specialist | specialists.csv | Medical specialists |
| subspecialist | subspecialists.csv | Medical subspecialists |

### CSV File Format

CSV files should follow these conventions:

1. **Header Row**: First row must contain column names
2. **Column Names**: Use snake_case or PascalCase
3. **Data Types**: Automatic type conversion is supported
4. **Empty Values**: Empty cells are handled gracefully

Example CSV format:
```csv
Code,Name,Status
001,General Medicine,true
002,Cardiology,true
003,Neurology,true
```

## Extending the Seeder

### Adding New Tables

1. **Create Entity Struct** (if not exists):
```go
type MyEntity struct {
    Id        int64     `gorm:"primaryKey;autoIncrement"`
    Code      string    `gorm:"uniqueIndex;not null"`
    Name      string    `gorm:"not null"`
    Status    bool      `gorm:"default:true"`
    CreatedAt time.Time
    UpdatedAt *time.Time
    DeletedAt *time.Time
}
```

2. **Register in DefaultRegistry**:
```go
registry.Register("mytable", TableConfig{
    TableName: "MyTable",
    CSVFile:   filepath.Join(basePath, "mytable.csv"),
    ColumnMap: map[string]string{
        "Code":   "Code",
        "Name":   "Name",
        "Status": "Status",
    },
    DeleteBefore: true,
    BatchSize:    50,
})
```

3. **Add to GetEntityByTableName**:
```go
case "mytable":
    return &MyEntity{}
```

### Custom Column Mapping

The seeder supports flexible column mapping:

```go
ColumnMap: map[string]string{
    "CSV_Column_Name": "StructFieldName",
    "province_code":   "ProvinceCode",
    "full_name":       "Name",
}
```

### Custom Type Conversion

For complex types, you can extend the `setFieldValue` function in `MasterSeeder`.

## Error Handling

The seeder provides detailed error messages for:

- **File Not Found**: CSV file doesn't exist
- **Invalid CSV Format**: Malformed CSV files
- **Database Errors**: Connection issues, constraint violations
- **Type Conversion**: Invalid data types
- **Validation**: Missing required fields

## Logging

The seeder provides comprehensive logging:

```
2024-01-01 12:00:00 [INFO] Starting seed for table: province
2024-01-01 12:00:00 [INFO] CSV file: internal/infrastructure/database/csv/provinces.csv
2024-01-01 12:00:00 [INFO] Batch size: 50
2024-01-01 12:00:00 [INFO] Delete before: true
2024-01-01 12:00:00 [INFO] Seeded province: 11 - ACEH
2024-01-01 12:00:00 [INFO] Seeding completed. 38 records processed from Province
```

## Performance Tips

1. **Batch Size**: Use larger batch sizes (100-500) for better performance
2. **Indexing**: Consider dropping indexes during seeding for large datasets
3. **Transactions**: Each batch is processed in a transaction
4. **Dry Run**: Always use dry-run first to validate data
5. **Delete Strategy**: Use `delete-before` for clean slate seeding

## Environment Variables

```bash
# Database connection (optional, defaults to localhost)
export DATABASE_URL="host=localhost user=postgres password=postgres dbname=health port=5432 sslmode=disable"
```

## Troubleshooting

### Common Issues

1. **CSV File Not Found**
   ```
   Error: CSV file not found: internal/infrastructure/database/csv/provinces.csv
   ```
   Solution: Ensure CSV files are in the correct directory

2. **Database Connection Failed**
   ```
   Error: Failed to connect to database: ...
   ```
   Solution: Check DATABASE_URL environment variable

3. **Invalid Column Mapping**
   ```
   Error: Failed to set field X: field not found
   ```
   Solution: Check ColumnMap configuration

4. **Type Conversion Error**
   ```
   Error: Failed to parse int from 'ABC': ...
   ```
   Solution: Check CSV data types match expected format

### Debug Mode

Use dry-run mode to preview data without inserting:
```bash
./bin/seeder dry-run province
```

## Contributing

1. Fork the repository
2. Create a feature branch
3. Add tests for new functionality
4. Submit a pull request

## License

This seeder is part of the service-general project and follows the same license terms.