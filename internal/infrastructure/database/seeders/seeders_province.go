package seeders

import (
	"encoding/csv"
	"fmt"
	"log"
	"os"
	"reflect"
	"strings"
	"time"

	"gorm.io/gorm"
)

// SeederConfig berisi konfigurasi untuk seeding
type SeederConfig struct {
	CSVPath      string
	TableName    string
	ColumnMap    map[string]string // mapping CSV header ke field struct
	SkipHeader   bool
	BatchSize    int
	DryRun       bool
	DeleteBefore bool
}

// MasterSeeder adalah struct utama untuk seeding fleksibel
type MasterSeeder struct {
	db     *gorm.DB
	config SeederConfig
}

// NewMasterSeeder membuat instance seeder baru
func NewMasterSeeder(db *gorm.DB, config SeederConfig) *MasterSeeder {
	if config.BatchSize == 0 {
		config.BatchSize = 100
	}
	return &MasterSeeder{
		db:     db,
		config: config,
	}
}

// SeedFromCSV melakukan seeding dari file CSV
func (s *MasterSeeder) SeedFromCSV(model interface{}) error {
	if s.config.CSVPath == "" {
		return fmt.Errorf("CSV path is required")
	}

	file, err := os.Open(s.config.CSVPath)
	if err != nil {
		return fmt.Errorf("failed to open CSV file: %w", err)
	}
	defer file.Close()

	reader := csv.NewReader(file)
	reader.Comma = ','

	// Baca header
	headers, err := reader.Read()
	if err != nil {
		return fmt.Errorf("failed to read CSV header: %w", err)
	}

	log.Printf("CSV Headers: %v", headers)

	// Hapus data lama jika diminta
	if s.config.DeleteBefore {
		if err := s.db.Where("1=1").Delete(model).Error; err != nil {
			return fmt.Errorf("failed to delete existing data: %w", err)
		}
		log.Printf("Deleted existing data from %s", s.config.TableName)
	}

	count := 0
	batch := []interface{}{}

	for {
		record, err := reader.Read()
		if err != nil {
			if err.Error() == "EOF" {
				break
			}
			log.Printf("Error reading row %d: %v", count+1, err)
			continue
		}

		if len(record) < len(headers) {
			log.Printf("Skipping incomplete row %d", count+1)
			continue
		}

		// Buat instance model baru
		instance := reflect.New(reflect.TypeOf(model).Elem()).Interface()

		// Mapping data dari CSV ke struct
		if err := s.mapCSVToStruct(headers, record, instance); err != nil {
			log.Printf("Error mapping row %d: %v", count+1, err)
			continue
		}

		// Set audit fields
		s.setAuditFields(instance)

		if s.config.DryRun {
			log.Printf("Dry run - Would insert: %+v", instance)
		} else {
			batch = append(batch, instance)

			// Insert batch jika sudah penuh
			if len(batch) >= s.config.BatchSize {
				if err := s.insertBatch(batch); err != nil {
					return fmt.Errorf("failed to insert batch: %w", err)
				}
				count += len(batch)
				batch = []interface{}{}
			}
		}
	}

	// Insert sisa batch
	if len(batch) > 0 && !s.config.DryRun {
		if err := s.insertBatch(batch); err != nil {
			return fmt.Errorf("failed to insert final batch: %w", err)
		}
		count += len(batch)
	}

	log.Printf("Seeding completed. %d records processed from %s", count, s.config.TableName)
	return nil
}

// mapCSVToStruct mapping data CSV ke struct berdasarkan column map atau reflection
func (s *MasterSeeder) mapCSVToStruct(headers []string, record []string, instance interface{}) error {
	v := reflect.ValueOf(instance).Elem()
	t := v.Type()

	for i, header := range headers {
		if i >= len(record) {
			continue
		}

		value := strings.TrimSpace(record[i])
		if value == "" {
			continue
		}

		// Cari field berdasarkan column map atau nama header
		fieldName := s.getFieldName(header)

		field, found := t.FieldByName(fieldName)
		if !found {
			// Coba dengan nama field yang berbeda case
			for j := 0; j < t.NumField(); j++ {
				f := t.Field(j)
				if strings.EqualFold(f.Name, fieldName) {
					field = f
					found = true
					break
				}
			}
		}

		if found {
			if err := s.setFieldValue(v.FieldByName(field.Name), value); err != nil {
				return fmt.Errorf("failed to set field %s: %w", field.Name, err)
			}
		}
	}

	return nil
}

// getFieldName mendapatkan nama field dari header CSV
func (s *MasterSeeder) getFieldName(header string) string {
	// Cek di column map dulu
	if fieldName, exists := s.config.ColumnMap[header]; exists {
		return fieldName
	}

	// Konversi header ke PascalCase
	words := strings.Split(header, "_")
	for i, word := range words {
		words[i] = strings.Title(strings.ToLower(word))
	}
	return strings.Join(words, "")
}

// setFieldValue mengisi value ke field struct dengan type conversion
func (s *MasterSeeder) setFieldValue(field reflect.Value, value string) error {
	if !field.CanSet() {
		return nil
	}

	switch field.Kind() {
	case reflect.String:
		field.SetString(value)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		// Handle bigserial (int64)
		if value == "" {
			field.SetInt(0)
		} else {
			intVal, err := s.parseInt(value)
			if err != nil {
				return err
			}
			field.SetInt(intVal)
		}
	case reflect.Bool:
		boolVal := strings.ToLower(value) == "true" || value == "1"
		field.SetBool(boolVal)
	case reflect.Interface:
		// Untuk field interface{} (seperti Code, Name, Status pada Ethnic)
		field.Set(reflect.ValueOf(value))
	default:
		// Handle pointer types
		if field.Kind() == reflect.Ptr {
			if value == "" {
				field.Set(reflect.Zero(field.Type()))
			} else {
				// Create new instance of the pointed type
				newValue := reflect.New(field.Type().Elem())
				if err := s.setFieldValue(newValue.Elem(), value); err != nil {
					return err
				}
				field.Set(newValue)
			}
		}
	}
	return nil
}

// parseInt parsing string ke int64
func (s *MasterSeeder) parseInt(value string) (int64, error) {
	var intVal int64
	_, err := fmt.Sscanf(value, "%d", &intVal)
	if err != nil {
		return 0, fmt.Errorf("failed to parse int from '%s': %w", value, err)
	}
	return intVal, nil
}

// setAuditFields mengisi field audit (CreatedAt, UpdatedAt, DeletedAt)
func (s *MasterSeeder) setAuditFields(instance interface{}) {
	v := reflect.ValueOf(instance).Elem()
	now := time.Now()

	// Set CreatedAt jika field ada
	if field := v.FieldByName("CreatedAt"); field.IsValid() && field.CanSet() {
		if field.Kind() == reflect.Struct {
			// Handle time.Time
			field.Set(reflect.ValueOf(now))
		} else if field.Kind() == reflect.Interface {
			field.Set(reflect.ValueOf(now))
		}
	}

	// Set UpdatedAt jika field ada (pointer)
	if field := v.FieldByName("UpdatedAt"); field.IsValid() && field.CanSet() {
		if field.Kind() == reflect.Ptr {
			field.Set(reflect.ValueOf(&now))
		}
	}

	// Set DeletedAt ke nil jika field ada
	if field := v.FieldByName("DeletedAt"); field.IsValid() && field.CanSet() {
		if field.Kind() == reflect.Ptr {
			field.Set(reflect.Zero(field.Type()))
		}
	}
}

// insertBatch insert batch data ke database
func (s *MasterSeeder) insertBatch(batch []interface{}) error {
	return s.db.Transaction(func(tx *gorm.DB) error {
		for _, item := range batch {
			// Gunakan FirstOrCreate untuk menghindari duplicate
			v := reflect.ValueOf(item).Elem()

			// Buat kondisi where berdasarkan field utama (Code atau ID)
			var condition interface{}
			if codeField := v.FieldByName("Code"); codeField.IsValid() {
				condition = map[string]interface{}{"Code": codeField.Interface()}
			} else if idField := v.FieldByName("Id"); idField.IsValid() {
				condition = map[string]interface{}{"Id": idField.Interface()}
			}

			if condition != nil {
				result := tx.Where(condition).FirstOrCreate(item)
				if result.Error != nil {
					return result.Error
				}
			} else {
				// Jika tidak ada kondisi, insert langsung
				if err := tx.Create(item).Error; err != nil {
					return err
				}
			}
		}
		return nil
	})
}

// Helper functions untuk seeding spesifik

// SeedProvinces seed data provinsi dari CSV
func SeedProvinces(db *gorm.DB, csvPath string) error {
	config := SeederConfig{
		CSVPath:      csvPath,
		TableName:    "Province",
		ColumnMap:    map[string]string{"Code": "Code", "Name": "Name"},
		SkipHeader:   true,
		BatchSize:    50,
		DeleteBefore: true,
	}

	seeder := NewMasterSeeder(db, config)

	// Definisikan struct untuk Province
	type Province struct {
		Id        int64  `gorm:"primaryKey;autoIncrement"`
		Code      string `gorm:"uniqueIndex;not null"`
		Name      string `gorm:"not null"`
		CreatedAt time.Time
		UpdatedAt *time.Time
		DeletedAt *time.Time
	}

	return seeder.SeedFromCSV(&Province{})
}

// SeedRegencies seed data kabupaten dari CSV
func SeedRegencies(db *gorm.DB, csvPath string) error {
	config := SeederConfig{
		CSVPath:      csvPath,
		TableName:    "Regency",
		ColumnMap:    map[string]string{"Code": "Code", "Province_Code": "ProvinceCode", "Name": "Name"},
		SkipHeader:   true,
		BatchSize:    100,
		DeleteBefore: true,
	}

	seeder := NewMasterSeeder(db, config)

	type Regency struct {
		Id           int64  `gorm:"primaryKey;autoIncrement"`
		Code         string `gorm:"uniqueIndex;not null"`
		ProvinceCode string `gorm:"index;not null"`
		Name         string `gorm:"not null"`
		CreatedAt    time.Time
		UpdatedAt    *time.Time
		DeletedAt    *time.Time
	}

	return seeder.SeedFromCSV(&Regency{})
}

// SeedEthnics seed data suku bangsa dari CSV
// func SeedEthnics(db *gorm.DB, csvPath string) error {
// 	config := SeederConfig{
// 		CSVPath:      csvPath,
// 		TableName:    "Ethnic",
// 		ColumnMap:    map[string]string{"Code": "Code", "Name": "Name", "Status": "Status"},
// 		SkipHeader:   true,
// 		BatchSize:    50,
// 		DeleteBefore: true,
// 	}

// 	seeder := NewMasterSeeder(db, config)

// 	// Gunakan struct Ethnic yang sudah ada
// 	return seeder.SeedFromCSV(&Ethnic{})
// }
