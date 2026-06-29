package seeders

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// TableConfig berisi konfigurasi untuk setiap tabel
type TableConfig struct {
	TableName    string
	CSVFile      string
	Entity       interface{}
	ColumnMap    map[string]string
	DeleteBefore bool
	BatchSize    int
}

// SeederRegistry berisi registry untuk semua tabel yang bisa di-seed
type SeederRegistry struct {
	tables map[string]TableConfig
}

// NewSeederRegistry membuat registry baru
func NewSeederRegistry() *SeederRegistry {
	return &SeederRegistry{
		tables: make(map[string]TableConfig),
	}
}

// Register mendaftarkan tabel ke registry
func (r *SeederRegistry) Register(name string, config TableConfig) {
	r.tables[name] = config
}

// Get mendapatkan konfigurasi tabel berdasarkan nama
func (r *SeederRegistry) Get(name string) (TableConfig, bool) {
	config, exists := r.tables[name]
	return config, exists
}

// List mengembalikan daftar nama tabel yang terdaftar
func (r *SeederRegistry) List() []string {
	names := make([]string, 0, len(r.tables))
	for name := range r.tables {
		names = append(names, name)
	}
	return names
}

// DefaultRegistry berisi konfigurasi default untuk semua tabel
func DefaultRegistry() *SeederRegistry {
	registry := NewSeederRegistry()

	basePath := "internal/infrastructure/database/csv"

	// Register Province
	registry.Register("province", TableConfig{
		TableName: "Province",
		CSVFile:   filepath.Join(basePath, "provinces.csv"),
		ColumnMap: map[string]string{
			"Code": "Code",
			"Name": "Name",
		},
		DeleteBefore: true,
		BatchSize:    50,
	})

	// Register Regency
	registry.Register("regency", TableConfig{
		TableName: "Regency",
		CSVFile:   filepath.Join(basePath, "regencies.csv"),
		ColumnMap: map[string]string{
			"Code":          "Code",
			"Province_Code": "ProvinceCode",
			"Name":          "Name",
		},
		DeleteBefore: true,
		BatchSize:    100,
	})

	// Register District
	registry.Register("district", TableConfig{
		TableName: "District",
		CSVFile:   filepath.Join(basePath, "districts.csv"),
		ColumnMap: map[string]string{
			"Code":         "Code",
			"Regency_Code": "RegencyCode",
			"Name":         "Name",
		},
		DeleteBefore: true,
		BatchSize:    200,
	})

	// Register Village
	registry.Register("village", TableConfig{
		TableName: "Village",
		CSVFile:   filepath.Join(basePath, "villages.csv"),
		ColumnMap: map[string]string{
			"Code":          "Code",
			"District_Code": "DistrictCode",
			"Name":          "Name",
		},
		DeleteBefore: true,
		BatchSize:    500,
	})

	// Register Ethnic
	registry.Register("ethnic", TableConfig{
		TableName: "Ethnic",
		CSVFile:   filepath.Join(basePath, "ethnics.csv"),
		ColumnMap: map[string]string{
			"Code":   "Code",
			"Name":   "Name",
			"Status": "Status",
		},
		DeleteBefore: true,
		BatchSize:    50,
	})

	// Register Language
	registry.Register("language", TableConfig{
		TableName: "Language",
		CSVFile:   filepath.Join(basePath, "languages.csv"),
		ColumnMap: map[string]string{
			"Code":   "Code",
			"Name":   "Name",
			"Status": "Status",
		},
		DeleteBefore: true,
		BatchSize:    50,
	})

	// Register Installation
	registry.Register("installation", TableConfig{
		TableName: "Installation",
		CSVFile:   filepath.Join(basePath, "installations.csv"),
		ColumnMap: map[string]string{
			"Code":   "Code",
			"Name":   "Name",
			"Status": "Status",
		},
		DeleteBefore: true,
		BatchSize:    50,
	})

	// Register Unit
	registry.Register("unit", TableConfig{
		TableName: "Unit",
		CSVFile:   filepath.Join(basePath, "units.csv"),
		ColumnMap: map[string]string{
			"Code":   "Code",
			"Name":   "Name",
			"Status": "Status",
		},
		DeleteBefore: true,
		BatchSize:    50,
	})

	// Register Specialist
	registry.Register("specialist", TableConfig{
		TableName: "Specialist",
		CSVFile:   filepath.Join(basePath, "specialists.csv"),
		ColumnMap: map[string]string{
			"Code":   "Code",
			"Name":   "Name",
			"Status": "Status",
		},
		DeleteBefore: true,
		BatchSize:    50,
	})

	// Register SubSpecialist
	registry.Register("subspecialist", TableConfig{
		TableName: "SubSpecialist",
		CSVFile:   filepath.Join(basePath, "subspecialists.csv"),
		ColumnMap: map[string]string{
			"Code":            "Code",
			"Specialist_Code": "SpecialistCode",
			"Name":            "Name",
			"Status":          "Status",
		},
		DeleteBefore: true,
		BatchSize:    50,
	})

	// Register RolPages (Menu Structure)
	registry.Register("rol_pages", TableConfig{
		TableName:    "rol_pages",
		Entity:       &RolPages{},
		DeleteBefore: true,
		BatchSize:    20,
	})

	return registry
}

// ValidateConfig memvalidasi konfigurasi
func (c *TableConfig) Validate() error {
	if c.TableName == "" {
		return fmt.Errorf("table name is required")
	}
	if c.CSVFile == "" {
		return fmt.Errorf("CSV file path is required")
	}

	// Cek apakah file CSV ada
	if _, err := os.Stat(c.CSVFile); os.IsNotExist(err) {
		return fmt.Errorf("CSV file not found: %s", c.CSVFile)
	}

	if c.BatchSize <= 0 {
		c.BatchSize = 100
	}

	return nil
}

// GetEntityByTableName mendapatkan entity berdasarkan nama tabel
func GetEntityByTableName(tableName string) interface{} {
	switch strings.ToLower(tableName) {
	case "province":
		return &Province{}
	case "regency":
		return &Regency{}
	case "district":
		return &District{}
	case "village":
		return &Village{}
	// case "ethnic":
	// 	return &Ethnic{}
	case "language":
		return &Language{}
	case "installation":
		return &Installation{}
	case "unit":
		return &Unit{}
	case "specialist":
		return &Specialist{}
	case "subspecialist":
		return &SubSpecialist{}
	case "rol_pages":
		return &RolPages{}
	default:
		return nil
	}
}

// Entity definitions

type Province struct {
	Id        int64  `gorm:"primaryKey;autoIncrement"`
	Code      string `gorm:"uniqueIndex;not null"`
	Name      string `gorm:"not null"`
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

type Regency struct {
	Id           int64  `gorm:"primaryKey;autoIncrement"`
	Code         string `gorm:"uniqueIndex;not null"`
	ProvinceCode string `gorm:"index;not null"`
	Name         string `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	DeletedAt    *time.Time
}

type District struct {
	Id          int64  `gorm:"primaryKey;autoIncrement"`
	Code        string `gorm:"uniqueIndex;not null"`
	RegencyCode string `gorm:"index;not null"`
	Name        string `gorm:"not null"`
	CreatedAt   time.Time
	UpdatedAt   *time.Time
	DeletedAt   *time.Time
}

type Village struct {
	Id           int64  `gorm:"primaryKey;autoIncrement"`
	Code         string `gorm:"uniqueIndex;not null"`
	DistrictCode string `gorm:"index;not null"`
	Name         string `gorm:"not null"`
	CreatedAt    time.Time
	UpdatedAt    *time.Time
	DeletedAt    *time.Time
}

type Language struct {
	Id        int64       `gorm:"primaryKey;autoIncrement"`
	Code      interface{} `gorm:"uniqueIndex;not null"`
	Name      interface{} `gorm:"not null"`
	Status    interface{} `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

type Installation struct {
	Id        int64       `gorm:"primaryKey;autoIncrement"`
	Code      interface{} `gorm:"uniqueIndex;not null"`
	Name      interface{} `gorm:"not null"`
	Status    interface{} `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

type Unit struct {
	Id        int64       `gorm:"primaryKey;autoIncrement"`
	Code      interface{} `gorm:"uniqueIndex;not null"`
	Name      interface{} `gorm:"not null"`
	Status    interface{} `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

type Specialist struct {
	Id        int64       `gorm:"primaryKey;autoIncrement"`
	Code      interface{} `gorm:"uniqueIndex;not null"`
	Name      interface{} `gorm:"not null"`
	Status    interface{} `gorm:"default:true"`
	CreatedAt time.Time
	UpdatedAt *time.Time
	DeletedAt *time.Time
}

type SubSpecialist struct {
	Id             int64       `gorm:"primaryKey;autoIncrement"`
	Code           interface{} `gorm:"uniqueIndex;not null"`
	SpecialistCode interface{} `gorm:"index;not null"`
	Name           interface{} `gorm:"not null"`
	Status         interface{} `gorm:"default:true"`
	CreatedAt      time.Time
	UpdatedAt      *time.Time
	DeletedAt      *time.Time
}

// RolPages struct untuk menu hierarchy
type RolPages struct {
	Id        int64     `gorm:"primaryKey;autoIncrement"`
	Name      string    `gorm:"not null;size:20"`
	Icon      string    `gorm:"size:100"`
	URL       string    `gorm:"not null"`
	Level     int16     `gorm:"not null;default:0"`
	Sort      int16     `gorm:"not null;default:0"`
	Parent    *int64    `gorm:"index"`
	Active    bool      `gorm:"not null;default:true"`
	CreatedAt time.Time `gorm:"not null;default:CURRENT_TIMESTAMP"`
	UpdatedAt *time.Time
	DeletedAt *time.Time `gorm:"index"`
}
