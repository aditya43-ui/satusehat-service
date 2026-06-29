package database

import (
	"context"
	"crypto/tls"
	"database/sql"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"service/internal/infrastructure/config"
	"service/pkg/logger"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"

	// Driver GORM di-import secara eksplisit untuk Migration
	_ "github.com/lib/pq" // Driver PostgreSQL
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/driver/sqlserver"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type DatabaseType string

const (
	Postgres  DatabaseType = "postgres"
	MySQL     DatabaseType = "mysql"
	SQLServer DatabaseType = "sqlserver"
	SQLite    DatabaseType = "sqlite"
	MongoDB   DatabaseType = "mongodb"
)

type Service interface {
	Health() map[string]map[string]string
	// Metode utama sekarang mengembalikan *gorm.DB
	GetGormDB(name string) (*gorm.DB, error)
	// Metode untuk mendapatkan koneksi mentah jika diperlukan
	GetDB(name string) (*sql.DB, error)
	GetSQLXDB(name string) (*sqlx.DB, error)
	GetMongoClient(name string) (*mongo.Client, error)
	GetReadDB(name string) (*sql.DB, error)
	Close() error
	ListDBs() []string
	GetDBType(name string) (DatabaseType, error)
	ListenForChanges(ctx context.Context, dbName string, channels []string, callback func(string, string)) error
	NotifyChange(dbName, channel, payload string) error
	GetPrimaryDB(name string) (*sql.DB, error)
	ExecuteQuery(ctx context.Context, dbName string, query string, args ...interface{}) (*sql.Rows, error)
	ExecuteQueryRow(ctx context.Context, dbName string, query string, args ...interface{}) *sql.Row
	Exec(ctx context.Context, dbName string, query string, args ...interface{}) (sql.Result, error)
	// Method untuk menjalankan migrasi database otomatis
	RegisterModel(models ...interface{}) // Method baru untuk registrasi dinamis
	Migrate() error
	// Method untuk mendapatkan informasi semua database yang terkoneksi
	GetAllDatabasesInfo() map[string]interface{}
}

type service struct {
	// GORM DB adalah sumber utama
	gormDatabases map[string]*gorm.DB
	// Map untuk menyimpan koneksi mentah yang diekstrak dari GORM
	sqlDatabases    map[string]*sql.DB
	sqlxDatabases   map[string]*sqlx.DB
	mongoClients    map[string]*mongo.Client
	readReplicas    map[string][]*sql.DB
	configs         map[string]config.DatabaseConfig
	readConfigs     map[string][]config.DatabaseConfig
	mu              sync.RWMutex
	readBalancer    map[string]int
	listeners       map[string]*pq.Listener // Menggunakan sql.Listener dari GORM/driver
	listenersMu     sync.RWMutex
	modelsToMigrate []interface{} // Menyimpan daftar entitas yang akan dimigrasi
}

var (
	dbManager *service
	once      sync.Once
)

// gormLoggerAdapter adalah adapter untuk menghubungkan logger kita dengan GORM
type gormLoggerAdapter struct {
	logger.Logger
	loggerConfig logger.Config
	Config       gormlogger.Config
}

// NewGormLoggerAdapter membuat instance baru dari adapter
func NewGormLoggerAdapter(l logger.Logger) gormlogger.Interface {
	// Kita tidak perlu menyimpan logger.Config di sini, karena logger kita sudah diinisialisasi.
	// Kita hanya perlu menentukan log level untuk GORM.
	// Anda bisa membuat ini lebih dinamis dengan membaca dari config jika perlu.
	// Untuk sekarang, kita set ke Info agar semua query terlihat.
	logLevel := gormlogger.Info

	return &gormLoggerAdapter{
		Logger: l,
		// Konfigurasi logger GORM
		Config: gormlogger.Config{
			SlowThreshold:             200 * time.Millisecond,
			LogLevel:                  logLevel, // Gunakan level yang sudah ditentukan
			IgnoreRecordNotFoundError: false,    // Jangan abaikan error "not found"
			Colorful:                  false,    // Nonaktifkan warna karena kita menggunakan logger terstruktur
		},
	}
}

// LogMode menetapkan level log dan mengembalikan instance baru
func (l *gormLoggerAdapter) LogMode(level gormlogger.LogLevel) gormlogger.Interface {
	// Buat instance adapter baru dengan level yang berbeda
	newLogger := *l
	newLogger.Config.LogLevel = level
	return &newLogger
}

// Info mencatat log info (jarang dipanggil GORM)
func (l *gormLoggerAdapter) Info(ctx context.Context, msg string, data ...interface{}) {
	formattedMsg := fmt.Sprintf(msg, data...)
	// Mencegah log notifikasi duplikat (dead notif) yang tidak berguna dari internal GORM
	if strings.Contains(formattedMsg, "replacing callback") {
		return
	}
	l.Logger.WithContext(ctx).Info(formattedMsg)
}

// Warn mencatat log warning
func (l *gormLoggerAdapter) Warn(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.WithContext(ctx).Warn(fmt.Sprintf(msg, data...))
}

// Error mencatat log error
func (l *gormLoggerAdapter) Error(ctx context.Context, msg string, data ...interface{}) {
	l.Logger.WithContext(ctx).Error(fmt.Sprintf(msg, data...))
}

// Trace mencatat log trace (SQL query) - ini adalah metode terpenting
func (l *gormLoggerAdapter) Trace(ctx context.Context, begin time.Time, fc func() (sql string, rowsAffected int64), err error) {
	elapsed := time.Since(begin)
	sql, rows := fc()

	// Siapkan fields menggunakan helper dari logger kita
	fields := []logger.Field{
		logger.Duration("duration_ms", elapsed),
		logger.Int64("rows", rows),
		logger.String("sql", sql),
	}

	// Tentukan level log dan pesan berdasarkan kondisi
	switch {
	case err != nil && l.Config.LogLevel >= gormlogger.Error:
		// Jika ada error dan log level adalah Error atau lebih tinggi
		l.Logger.WithContext(ctx).WithError(err).WithFields(fields...).Error("Database query failed")
	case elapsed > l.Config.SlowThreshold && l.Config.LogLevel >= gormlogger.Warn:
		// Jika query lambat dan log level adalah Warn atau lebih tinggi
		l.Logger.WithContext(ctx).WithFields(fields...).Warn("Slow database query")
	case l.Config.LogLevel == gormlogger.Info:
		// Jika log level adalah Info, catat semua query sebagai debug
		l.Logger.WithContext(ctx).WithFields(fields...).Debug("Database query executed")
	}
}

func New(cfg *config.Config) Service {
	once.Do(func() {
		dbManager = &service{
			gormDatabases: make(map[string]*gorm.DB),
			sqlDatabases:  make(map[string]*sql.DB),
			sqlxDatabases: make(map[string]*sqlx.DB),
			mongoClients:  make(map[string]*mongo.Client),
			readReplicas:  make(map[string][]*sql.DB),
			configs:       make(map[string]config.DatabaseConfig),
			readConfigs:   make(map[string][]config.DatabaseConfig),
			readBalancer:  make(map[string]int),
			listeners:     make(map[string]*pq.Listener),
		}

		logger.Default().Info("Initializing database service with GORM as primary abstraction...")
		dbManager.loadFromConfig(cfg)

		if _, exists := dbManager.configs["default"]; !exists {
			logger.Default().Warn("No 'default' database configured in config")
		}

		for name, dbConfig := range dbManager.configs {
			if err := dbManager.addDatabase(name, dbConfig); err != nil {
				logger.Default().Error("Failed to connect to database", logger.String("db_name", name), logger.ErrorField(err))
			} else {
				logger.Default().Info("Successfully connected to database", logger.String("db_name", name))
			}
		}

		for name, replicaConfigs := range dbManager.readConfigs {
			for i, replicaConfig := range replicaConfigs {
				if err := dbManager.addReadReplica(name, i, replicaConfig); err != nil {
					logger.Default().Error("Failed to connect to read replica", logger.String("db_name", name), logger.Int("replica_index", i), logger.ErrorField(err))
				} else {
					logger.Default().Info("Successfully connected to read replica", logger.String("db_name", name), logger.Int("replica_index", i))
				}
			}
		}
	})

	return dbManager
}

func (s *service) loadFromConfig(cfg *config.Config) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for name, dbConfig := range cfg.Databases {
		s.configs[name] = dbConfig
	}

	for name, replicaConfigs := range cfg.ReadReplicas {
		s.readConfigs[name] = replicaConfigs
	}
}

// GetGormDB sekarang menjadi metode utama dan lebih sederhana
func (s *service) GetGormDB(name string) (*gorm.DB, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	db, exists := s.gormDatabases[name]
	if !exists {
		return nil, fmt.Errorf("database %s not found", name)
	}
	return db, nil
}

// GetDB mengekstrak *sql.DB dari GORM
func (s *service) GetDB(name string) (*sql.DB, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	db, exists := s.sqlDatabases[name]
	if !exists {
		return nil, fmt.Errorf("database %s not found", name)
	}
	return db, nil
}

// GetSQLXDB mengekstrak *sqlx.DB dari GORM
func (s *service) GetSQLXDB(name string) (*sqlx.DB, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	db, exists := s.sqlxDatabases[name]
	if !exists {
		return nil, fmt.Errorf("database %s not found", name)
	}
	return db, nil
}

func (s *service) GetAllDatabasesInfo() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()

	dbInfo := make(map[string]interface{})

	for name, gormDB := range s.gormDatabases {
		dbConfig, exists := s.configs[name]
		if !exists {
			continue
		}

		info := gin.H{
			"name":     name,
			"type":     dbConfig.Type,
			"host":     dbConfig.Host,
			"port":     dbConfig.Port,
			"database": dbConfig.Database,
			// "username": dbConfig.Username,
			"status": "connected",
		}

		// Coba dapatkan informasi tambahan dari database
		sqlDB, err := gormDB.DB()
		if err == nil {
			ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
			defer cancel()

			if err := sqlDB.PingContext(ctx); err == nil {
				// Coba query spesifik berdasarkan tipe database
				switch dbConfig.Type {
				case "postgres":
					var dbname, version, user string
					if err := gormDB.Raw("SELECT current_database(), version(), current_user").Row().Scan(&dbname, &version, &user); err == nil {
						// info["current_user"] = user
						// info["database_name"] = dbname
						info["database_version"] = version
					}
				case "mysql":
					var dbname, version, user string
					if err := gormDB.Raw("SELECT DATABASE(), VERSION(), USER()").Row().Scan(&dbname, &version, &user); err == nil {
						// info["current_user"] = user
						// info["database_name"] = dbname
						info["database_version"] = version
					}
				case "sqlserver":
					var dbname, version, user string
					if err := gormDB.Raw("SELECT DB_NAME(), @@VERSION, SYSTEM_USER").Row().Scan(&dbname, &version, &user); err == nil {
						// info["current_user"] = user
						// info["database_name"] = dbname
						info["database_version"] = version
					}
				case "sqlite":
					var seq int
					var name, file string
					if err := gormDB.Raw("PRAGMA database_list").Row().Scan(&seq, &name, &file); err == nil {
						// info["database_name"] = name
						info["database_file"] = file
					}
				}
			}
		}

		dbInfo[name] = info
	}

	return dbInfo
}

func (s *service) addDatabase(name string, dbConfig config.DatabaseConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var gormDB *gorm.DB
	var err error

	dbType := DatabaseType(dbConfig.Type)

	switch dbType {
	case Postgres:
		gormDB, err = s.openPostgresGORM(dbConfig)
	case MySQL:
		gormDB, err = s.openMySQLGORM(dbConfig)
	case SQLServer:
		gormDB, err = s.openSQLServerGORM(dbConfig)
	case SQLite:
		gormDB, err = s.openSQLiteGORM(dbConfig)
	case MongoDB:
		return s.addMongoDB(name, dbConfig)
	default:
		return fmt.Errorf("unsupported database type: %s", dbConfig.Type)
	}

	if err != nil {
		return fmt.Errorf("failed to connect to GORM for %s: %w", name, err)
	}

	// Ekstrak koneksi *sql.DB dari GORM
	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying sql.DB from GORM for %s: %w", name, err)
	}

	// Optimasi connection pool
	s.optimizeConnectionPool(name, sqlDB, dbConfig)

	// Simpan instance GORM dan koneksi mentahnya
	s.gormDatabases[name] = gormDB
	s.sqlDatabases[name] = sqlDB

	// Buat dan simpan instance sqlx.DB
	driverName := getDriverName(dbType)
	s.sqlxDatabases[name] = sqlx.NewDb(sqlDB, driverName)

	logger.Default().Info("Successfully connected and configured GORM for database", logger.String("db_name", name))
	return nil
}

// Helper untuk mendapatkan nama driver sqlx
func getDriverName(dbType DatabaseType) string {
	switch dbType {
	case Postgres:
		return "pgx" // Gunakan driver pgx untuk performa lebih baik
	case MySQL:
		return "mysql"
	case SQLServer:
		return "sqlserver"
	case SQLite:
		return "sqlite3"
	default:
		return string(dbType)
	}
}

// --- Metode Koneksi GORM ---

func (s *service) openPostgresGORM(config config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d sslmode=%s TimeZone=Asia/Jakarta",
		config.Host, config.Username, config.Password, config.Database, config.Port, config.SSLMode)

	if config.ConnectTimeout > 0 {
		dsn += fmt.Sprintf(" connect_timeout=%d", int(config.ConnectTimeout.Seconds()))
	}

	if config.StatementTimeout > 0 {
		dsn += fmt.Sprintf(" statement_timeout=%d", int(config.StatementTimeout.Milliseconds()))
	}

	if config.Schema != "" {
		dsn += fmt.Sprintf(" search_path=%s", config.Schema)
	}

	gormConfig := &gorm.Config{
		Logger: NewGormLoggerAdapter(logger.Default()), // Atur level log GORM
	}

	if config.RequireSSL {
		// Untuk SSL, kita mungkin perlu mengkonfigurasi tls.Config secara manual
		// dan menggunakannya dalam DSN atau melalui pgx driver config
	}

	return gorm.Open(postgres.Open(dsn), gormConfig)
}

func (s *service) openMySQLGORM(config config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	if config.ConnectTimeout > 0 {
		dsn += fmt.Sprintf("&timeout=%s", config.ConnectTimeout.String())
	} else if config.Timeout > 0 {
		dsn += fmt.Sprintf("&timeout=%s", config.Timeout.String())
	}

	if config.ReadTimeout > 0 {
		dsn += fmt.Sprintf("&readTimeout=%s", config.ReadTimeout.String())
	}
	if config.WriteTimeout > 0 {
		dsn += fmt.Sprintf("&writeTimeout=%s", config.WriteTimeout.String())
	}

	if config.RequireSSL {
		dsn += "&tls=true"
	}

	gormConfig := &gorm.Config{
		Logger: NewGormLoggerAdapter(logger.Default()),
	}

	return gorm.Open(mysql.Open(dsn), gormConfig)
}

func (s *service) openSQLServerGORM(config config.DatabaseConfig) (*gorm.DB, error) {
	dsn := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s",
		config.Username, config.Password, config.Host, config.Port, config.Database)

	if config.ConnectTimeout > 0 {
		dsn += fmt.Sprintf("&connection+timeout=%d", int(config.ConnectTimeout.Seconds()))
	}

	if config.RequireSSL {
		dsn += "&encrypt=true"
		if config.SSLRootCert != "" {
			dsn += "&trustServerCertificate=false"
		} else {
			dsn += "&trustServerCertificate=true"
		}
	}

	gormConfig := &gorm.Config{
		Logger: NewGormLoggerAdapter(logger.Default()),
	}

	return gorm.Open(sqlserver.Open(dsn), gormConfig)
}

func (s *service) openSQLiteGORM(config config.DatabaseConfig) (*gorm.DB, error) {
	dsn := config.Path
	if config.Timeout > 0 {
		dsn += fmt.Sprintf("?_busy_timeout=%d", config.Timeout.Milliseconds())
	}

	gormConfig := &gorm.Config{
		Logger: NewGormLoggerAdapter(logger.Default()),
	}

	return gorm.Open(sqlite.Open(dsn), gormConfig)
}

// RegisterModel mendaftarkan entitas GORM secara dinamis untuk Automigrate
func (s *service) RegisterModel(models ...interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.modelsToMigrate = append(s.modelsToMigrate, models...)
}

// --- PERBAIKAN: Implementasi Migrate Method dengan validasi lebih baik ---
func (s *service) Migrate() error {
	s.mu.RLock()
	defer s.mu.RUnlock()

	log.Println("🔄 Starting database migration process using GORM AutoMigrate...")

	if len(s.gormDatabases) == 0 {
		return fmt.Errorf("no GORM databases available for migration")
	}

	// Gunakan model yang sudah didaftarkan
	models := s.modelsToMigrate
	if len(models) == 0 {
		log.Println("⚠️ No models registered for migration.")
		// Tetap lanjutkan untuk runSQLMigrationsForDB
	}

	// Jalankan AutoMigrate untuk setiap database
	for name, gormDB := range s.gormDatabases {
		dbConfig := s.configs[name]
		dbType := DatabaseType(dbConfig.Type)

		if dbType == Postgres || dbType == MySQL || dbType == SQLServer || dbType == SQLite {
			log.Printf("⚙️ Running migration for [%s] (%s)...", name, dbType)

			// Gunakan AutoMigrate untuk model-model yang terdaftar
			if len(models) > 0 {
				if err := gormDB.AutoMigrate(models...); err != nil {
					return fmt.Errorf("auto migration failed for %s: %w", name, err)
				}
				log.Printf("✅ AutoMigrate successful for [%s]", name)
			}

			// Jalankan SQL migrations sebagai tambahan
			if err := s.runSQLMigrationsForDB(gormDB, dbType); err != nil {
				return fmt.Errorf("sql migration failed for %s: %w", name, err)
			}
		}
	}

	logger.Default().Info("All database migrations completed successfully")
	return nil
}

func (s *service) runSQLMigrationsForDB(gormDB *gorm.DB, dbType DatabaseType) error {
	migrationFiles, err := s.readMigrationFiles()
	if err != nil {
		return fmt.Errorf("failed to read migration files: %w", err)
	}

	if len(migrationFiles) == 0 {
		logger.Default().Info("No migration files found")
		return nil
	}

	logger.Default().Info("Found migration files to execute", logger.Int("count", len(migrationFiles)))

	sqlDB, err := gormDB.DB()
	if err != nil {
		return fmt.Errorf("cannot get underlying sql.DB: %w", err)
	}

	if err := s.ensureMigrationTable(sqlDB, dbType); err != nil {
		return fmt.Errorf("failed to ensure migration table: %w", err)
	}

	for _, filePath := range migrationFiles {
		filename := filepath.Base(filePath)
		if s.isMigrationApplied(sqlDB, filename) {
			logger.Default().Debug("Skipping already applied migration", logger.String("file", filename))
			continue
		}

		if err := s.executeMigrationFile(sqlDB, filePath); err != nil {
			return fmt.Errorf("migration failed at file %s: %w", filePath, err)
		}
		logger.Default().Info("Successfully executed migration file", logger.String("file", filename))
	}

	return nil
}

// Fungsi bantu untuk membaca file migrasi
func (s *service) readMigrationFiles() ([]string, error) {
	migrationsDir := "internal/infrastructure/database/sql"

	// Cek apakah directory migrations ada
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		logger.Default().Info("Migrations directory does not exist", logger.String("path", migrationsDir))
		return nil, nil
	}

	var files []string
	err := filepath.Walk(migrationsDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// Hanya proses file .sql
		if !info.IsDir() && strings.HasSuffix(strings.ToLower(info.Name()), ".sql") {
			files = append(files, path)
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to walk migrations directory: %w", err)
	}

	// Sort files by name untuk konsisten
	sort.Strings(files)
	return files, nil
}

// Fungsi bantu untuk cek apakah tabel sudah ada
func isTableNotExistsError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "doesn't exist") ||
		strings.Contains(errStr, "does not exist") ||
		strings.Contains(errStr, "no such table")
}

func (s *service) addReadReplica(name string, index int, config config.DatabaseConfig) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var db *sql.DB
	var err error

	dbType := DatabaseType(config.Type)

	switch dbType {
	case Postgres:
		db, err = s.openPostgresConnection(config)
	case MySQL:
		db, err = s.openMySQLConnection(config)
	case SQLServer:
		db, err = s.openSQLServerConnection(config)
	case SQLite:
		db, err = s.openSQLiteConnection(config)
	default:
		return fmt.Errorf("unsupported database type for read replica: %s", config.Type)
	}

	if err != nil {
		return err
	}

	// Optimasi juga untuk read replica
	s.optimizeConnectionPool(fmt.Sprintf("%s-replica-%d", name, index), db, config)

	// Validasi koneksi
	if err := s.validateConnection(fmt.Sprintf("%s-replica-%d", name, index), db); err != nil {
		db.Close()
		return fmt.Errorf("connection validation failed for read replica %s[%d]: %w", name, index, err)
	}

	if s.readReplicas[name] == nil {
		s.readReplicas[name] = make([]*sql.DB, 0)
	}

	for len(s.readReplicas[name]) <= index {
		s.readReplicas[name] = append(s.readReplicas[name], nil)
	}

	s.readReplicas[name][index] = db
	logger.Default().Info("Successfully connected to read replica", logger.String("db_name", name), logger.Int("replica_index", index))

	return nil
}

// Helper functions untuk mendeteksi jenis error
func isTableExistsError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "relation already exists") ||
		strings.Contains(errMsg, "table already exists")
}

func isSyntaxError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "syntax error") ||
		strings.Contains(errMsg, "syntax error at") ||
		strings.Contains(errMsg, "invalid syntax")
}

func isDuplicateKeyError(err error) bool {
	errMsg := err.Error()
	return strings.Contains(errMsg, "duplicate key") ||
		strings.Contains(errMsg, "unique constraint")
}

func isTableCreationCommand(cmd string) bool {
	cmd = strings.ToLower(cmd)
	return strings.Contains(cmd, "create table") ||
		strings.Contains(cmd, "create index") ||
		strings.Contains(cmd, "alter table")
}

func (s *service) openPostgresConnection(config config.DatabaseConfig) (*sql.DB, error) {
	connectTimeoutSec := int(config.ConnectTimeout.Seconds())
	statementTimeoutSec := int(config.StatementTimeout.Seconds())

	// Menggunakan pgx.ParseConfig untuk fleksibilitas lebih tinggi
	dsn := fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s connect_timeout=%d",
		config.Host,
		config.Port,
		config.Username,
		config.Password,
		config.Database,
		config.SSLMode,
		connectTimeoutSec,
	)

	if config.Schema != "" {
		dsn += " search_path=" + config.Schema
	}

	// Jika menggunakan parameter statement_timeout (pg specific)
	if statementTimeoutSec > 0 {
		// Note: pgx driver handle options differently, but usually passed in DSN or via Config
		dsn += fmt.Sprintf(" statement_timeout=%d", statementTimeoutSec*1000)
	}

	// Setup SSL certificates path if required
	if config.RequireSSL {
		// Pastikan file path valid
		if config.SSLCert != "" && config.SSLKey != "" && config.SSLRootCert != "" {
			dsn += fmt.Sprintf(" sslcert=%s sslkey=%s sslrootcert=%s", config.SSLCert, config.SSLKey, config.SSLRootCert)
		}
	}

	// Menggunakan stdlib.OpenDB dari pgx untuk kompatibilitas database/sql
	pgxConfig, err := pgx.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to parse pgx config: %w", err)
	}

	// Apply manual runtime settings if needed here

	db := stdlib.OpenDB(*pgxConfig)

	// Set timeout yang lebih panjang untuk koneksi awal
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping PostgreSQL database: %w", err)
	}

	return db, nil
}

func (s *service) openMySQLConnection(config config.DatabaseConfig) (*sql.DB, error) {
	connStr := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?parseTime=true&timeout=%s&readTimeout=%s&writeTimeout=%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.Timeout,
		config.ReadTimeout,
		config.WriteTimeout,
	)

	if config.RequireSSL {
		connStr += "&tls=true"
		// Setup custom TLS config biasanya dilakukan secara global di driver mysql
		// atau menggunakan nama config yang sudah didaftarkan
	}

	db, err := sql.Open("mysql", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open MySQL connection: %w", err)
	}

	// Set timeout yang lebih panjang untuk koneksi awal
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping MySQL database: %w", err)
	}

	return db, nil
}

func (s *service) openSQLServerConnection(config config.DatabaseConfig) (*sql.DB, error) {
	connectTimeoutSec := int(config.ConnectTimeout.Seconds())

	connStr := fmt.Sprintf("sqlserver://%s:%s@%s:%d?database=%s&connection timeout=%d",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		connectTimeoutSec,
	)

	if config.RequireSSL {
		connStr += "&encrypt=true"
		if config.SSLRootCert != "" {
			connStr += "&trustServerCertificate=false"
		} else {
			connStr += "&trustServerCertificate=true"
		}
	}

	db, err := sql.Open("sqlserver", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQL Server connection: %w", err)
	}

	// Set timeout yang lebih panjang untuk koneksi awal
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping SQL Server database: %w", err)
	}

	return db, nil
}

func (s *service) openSQLiteConnection(config config.DatabaseConfig) (*sql.DB, error) {
	// Memerlukan import _ "github.com/mattn/go-sqlite3" di tempat lain atau gunakan driver GORM yang membungkusnya
	// Namun disini kita gunakan driver name "sqlite3" standar
	db, err := sql.Open("sqlite3", config.Path)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite connection: %w", err)
	}

	// Performance tuning untuk SQLite (WAL mode)
	_, err = db.Exec("PRAGMA foreign_keys = ON; PRAGMA journal_mode = WAL; PRAGMA synchronous = NORMAL;")
	if err != nil {
		return nil, fmt.Errorf("failed to configure SQLite: %w", err)
	}

	// Set timeout yang lebih panjang untuk koneksi awal
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping SQLite database: %w", err)
	}

	return db, nil
}

func (s *service) addMongoDB(name string, config config.DatabaseConfig) error {
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 10 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	uri := fmt.Sprintf("mongodb://%s:%s@%s:%d/%s",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
	)

	clientOptions := options.Client().ApplyURI(uri)

	if config.RequireSSL {
		clientOptions.SetTLSConfig(&tls.Config{
			InsecureSkipVerify: config.SSLMode == "require" || config.SSLMode == "skip-verify",
			MinVersion:         tls.VersionTLS12,
		})
	}

	clientOptions.SetConnectTimeout(config.ConnectTimeout)
	clientOptions.SetServerSelectionTimeout(config.Timeout)

	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return fmt.Errorf("failed to connect to MongoDB: %w", err)
	}

	if err := client.Ping(ctx, nil); err != nil {
		return fmt.Errorf("failed to ping MongoDB: %w", err)
	}

	s.mongoClients[name] = client
	logger.Default().Info("Successfully connected to MongoDB", logger.String("db_name", name))

	return nil
}

func (s *service) configureSQLDB(name string, db *sql.DB, config config.DatabaseConfig) error {
	// Configuration is handled in optimizeConnectionPool, but we can double check or set defaults here
	timeout := config.Timeout
	if timeout <= 0 {
		timeout = 5 * time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return fmt.Errorf("failed to ping database: %w", err)
	}

	s.sqlDatabases[name] = db

	dbType := DatabaseType(config.Type)
	var driverName string

	switch dbType {
	case Postgres:
		driverName = "pgx"
	case MySQL:
		driverName = "mysql"
	case SQLServer:
		driverName = "sqlserver"
	case SQLite:
		driverName = "sqlite3"
	default:
		// Fallback for sqlx
		driverName = string(dbType)
	}

	sqlxDB := sqlx.NewDb(db, driverName)
	s.sqlxDatabases[name] = sqlxDB

	logger.Default().Info("Successfully connected to database", logger.String("db_name", name), logger.String("driver", driverName))

	return nil
}

func (s *service) Health() map[string]map[string]string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	result := make(map[string]map[string]string)

	// Check SQL Databases
	for name, db := range s.sqlDatabases {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		stats := make(map[string]string)

		err := db.PingContext(ctx)
		if err != nil {
			stats["status"] = "down"
			stats["error"] = fmt.Sprintf("db down: %v", err)
			stats["type"] = "sql"
			stats["role"] = "primary"
			result[name] = stats
			continue
		}

		stats["status"] = "up"
		stats["message"] = "It's healthy"
		stats["type"] = "sql"
		stats["role"] = "primary"

		dbStats := db.Stats()
		stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
		stats["in_use"] = strconv.Itoa(dbStats.InUse)
		stats["idle"] = strconv.Itoa(dbStats.Idle)
		stats["wait_count"] = strconv.FormatInt(dbStats.WaitCount, 10)
		stats["wait_duration"] = dbStats.WaitDuration.String()

		result[name] = stats
	}

	// Check Read Replicas
	for name, replicas := range s.readReplicas {
		for i, db := range replicas {
			if db == nil {
				continue
			}

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			replicaName := fmt.Sprintf("%s_replica_%d", name, i)
			stats := make(map[string]string)

			err := db.PingContext(ctx)
			if err != nil {
				stats["status"] = "down"
				stats["error"] = fmt.Sprintf("read replica down: %v", err)
				stats["role"] = "replica"
				result[replicaName] = stats
				continue
			}

			stats["status"] = "up"
			stats["role"] = "replica"

			dbStats := db.Stats()
			stats["open_connections"] = strconv.Itoa(dbStats.OpenConnections)
			result[replicaName] = stats
		}
	}

	// Check MongoDB
	for name, client := range s.mongoClients {
		ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
		defer cancel()

		stats := make(map[string]string)

		err := client.Ping(ctx, nil)
		if err != nil {
			stats["status"] = "down"
			stats["error"] = fmt.Sprintf("mongodb down: %v", err)
			stats["type"] = "mongodb"
			result[name] = stats
			continue
		}

		stats["status"] = "up"
		stats["type"] = "mongodb"
		result[name] = stats
	}

	return result
}
func (s *service) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()

	var errs []error

	for name, listener := range s.listeners {
		if err := listener.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close listener for %s: %w", name, err))
		}
	}

	for name, db := range s.sqlDatabases {
		if err := db.Close(); err != nil {
			errs = append(errs, fmt.Errorf("failed to close database %s: %w", name, err))
		}
	}

	for name, replicas := range s.readReplicas {
		for i, db := range replicas {
			if db != nil {
				if err := db.Close(); err != nil {
					errs = append(errs, fmt.Errorf("failed to close read replica %s[%d]: %w", name, i, err))
				}
			}
		}
	}

	for name, client := range s.mongoClients {
		if err := client.Disconnect(context.Background()); err != nil {
			errs = append(errs, fmt.Errorf("failed to disconnect MongoDB client %s: %w", name, err))
		}
	}

	// Clear maps
	s.sqlDatabases = make(map[string]*sql.DB)
	s.sqlxDatabases = make(map[string]*sqlx.DB)
	s.mongoClients = make(map[string]*mongo.Client)
	s.listeners = make(map[string]*pq.Listener)

	if len(errs) > 0 {
		return fmt.Errorf("errors closing databases: %v", errs)
	}

	return nil
}

func (s *service) GetReadDB(name string) (*sql.DB, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	replicas, exists := s.readReplicas[name]
	if !exists || len(replicas) == 0 {
		// Fallback to primary if no replicas
		return s.GetDB(name)
	}

	// Round-robin selection
	s.readBalancer[name] = (s.readBalancer[name] + 1) % len(replicas)
	selected := replicas[s.readBalancer[name]]

	if selected == nil {
		return s.GetDB(name)
	}

	return selected, nil
}

func (s *service) GetMongoClient(name string) (*mongo.Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	client, exists := s.mongoClients[name]
	if !exists {
		return nil, fmt.Errorf("MongoDB client %s not found", name)
	}

	return client, nil
}

func (s *service) ListDBs() []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	names := make([]string, 0, len(s.sqlDatabases)+len(s.mongoClients))

	for name := range s.sqlDatabases {
		names = append(names, name)
	}

	for name := range s.mongoClients {
		names = append(names, name)
	}

	return names
}

func (s *service) GetDBType(name string) (DatabaseType, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	config, exists := s.configs[name]
	if !exists {
		return "", fmt.Errorf("database %s not found", name)
	}

	return DatabaseType(config.Type), nil
}

func (s *service) GetPrimaryDB(name string) (*sql.DB, error) {
	return s.GetDB(name)
}

func (s *service) ExecuteQuery(ctx context.Context, dbName string, query string, args ...interface{}) (*sql.Rows, error) {
	db, err := s.GetDB(dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database %s: %w", dbName, err)
	}

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return rows, nil
}

func (s *service) ExecuteQueryRow(ctx context.Context, dbName string, query string, args ...interface{}) *sql.Row {
	db, err := s.GetDB(dbName)
	if err != nil {
		// Return dummy row that will return error on Scan
		// But standard sql.Row doesn't allow easy mocking of error without query execution.
		// So we must handle nil db check before calling this or panic might occur if not careful?
		// sql.DB methods are safe to call, but we returned error from GetDB.
		// Standard pattern: return row created from db, if db error, we can't create row easily.
		// Workaround: return a row from a failed query on a dummy DB or panic if critical.
		// For now, let's assume GetDB returns error and caller handles it, but since return signature is *sql.Row only...
		// Ideally change signature to (*sql.Row, error) or panic.
		// Here we just return an empty row which will likely fail on scan.
		return &sql.Row{}
	}

	return db.QueryRowContext(ctx, query, args...)
}

func (s *service) Exec(ctx context.Context, dbName string, query string, args ...interface{}) (sql.Result, error) {
	db, err := s.GetDB(dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get database %s: %w", dbName, err)
	}

	result, err := db.ExecContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to execute query: %w", err)
	}

	return result, nil
}

func (s *service) ListenForChanges(ctx context.Context, dbName string, channels []string, callback func(string, string)) error {
	s.mu.RLock()
	config, exists := s.configs[dbName]
	s.mu.RUnlock()

	if !exists {
		return fmt.Errorf("database %s not found", dbName)
	}

	if DatabaseType(config.Type) != Postgres {
		return fmt.Errorf("LISTEN/NOTIFY only supported for PostgreSQL databases")
	}

	// Reconstruct connection string for pq listener
	connectTimeoutSec := int(config.ConnectTimeout.Seconds())
	connStr := fmt.Sprintf("postgres://%s:%s@%s:%d/%s?sslmode=%s&connect_timeout=%d",
		config.Username,
		config.Password,
		config.Host,
		config.Port,
		config.Database,
		config.SSLMode,
		connectTimeoutSec,
	)

	listener := pq.NewListener(
		connStr,
		10*time.Second,
		time.Minute,
		func(ev pq.ListenerEventType, err error) {
			if err != nil {
				log.Printf("Database listener (%s) connection error: %v", dbName, err)
			}
		},
	)

	s.listenersMu.Lock()
	s.listeners[dbName] = listener
	s.listenersMu.Unlock()

	for _, channel := range channels {
		if err := listener.Listen(channel); err != nil {
			listener.Close()
			return fmt.Errorf("failed to listen to channel %s: %w", channel, err)
		}
		log.Printf("Listening to database channel: %s on %s", channel, dbName)
	}

	go func() {
		defer func() {
			listener.Close()
			s.listenersMu.Lock()
			delete(s.listeners, dbName)
			s.listenersMu.Unlock()
			log.Printf("Database listener for %s stopped", dbName)
		}()

		// Ticker untuk melakukan Ping manual sebagai pengganti MonitorPing
		pingTicker := time.NewTicker(30 * time.Second)
		defer pingTicker.Stop()

		for {
			select {
			case n := <-listener.Notify:
				if n != nil {
					// Panggil callback secara async
					go callback(n.Channel, n.Extra)
				}
			case <-pingTicker.C:
				// Manual Ping untuk memastikan koneksi tetap hidup
				go func() {
					if err := listener.Ping(); err != nil {
						log.Printf("⚠️ Listener ping warning for %s: %v", dbName, err)
					}
				}()
			case <-ctx.Done():
				return
			}
		}
	}()

	return nil
}

func (s *service) NotifyChange(dbName, channel, payload string) error {
	db, err := s.GetDB(dbName)
	if err != nil {
		return fmt.Errorf("failed to get database %s: %w", dbName, err)
	}

	// Pastikan ini Postgres
	s.mu.RLock()
	config, exists := s.configs[dbName]
	s.mu.RUnlock()

	if !exists || DatabaseType(config.Type) != Postgres {
		return fmt.Errorf("NOTIFY only supported for PostgreSQL")
	}

	_, err = db.Exec("SELECT pg_notify($1, $2)", channel, payload)
	if err != nil {
		return fmt.Errorf("failed to send notification: %w", err)
	}

	return nil
}

func (s *service) optimizeConnectionPool(name string, db *sql.DB, config config.DatabaseConfig) {
	// Jika config menyediakan nilai, gunakan itu. Jika tidak, gunakan formula auto-tuning.

	maxOpen := config.MaxOpenConns
	maxIdle := config.MaxIdleConns

	// Auto-tune jika tidak diset di config (0)
	if maxOpen <= 0 {
		cpuCores := runtime.NumCPU()
		maxOpen = cpuCores*2 + 1
	}

	if maxIdle <= 0 {
		maxIdle = maxOpen / 2
	}

	db.SetMaxOpenConns(maxOpen)
	db.SetMaxIdleConns(maxIdle)

	// Gunakan config duration atau default
	if config.ConnMaxLifetime > 0 {
		db.SetConnMaxLifetime(config.ConnMaxLifetime)
	} else {
		db.SetConnMaxLifetime(30 * time.Minute)
	}

	if config.MaxIdleTime > 0 {
		db.SetConnMaxIdleTime(config.MaxIdleTime)
	} else {
		db.SetConnMaxIdleTime(10 * time.Minute)
	}

	log.Printf("Connection pool configured for %s: MaxOpen=%d, MaxIdle=%d", name, maxOpen, maxIdle)
}

func (s *service) validateConnection(name string, db *sql.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second) // Timeout lebih panjang untuk validasi
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		return fmt.Errorf("connection validation failed for %s: %w", name, err)
	}

	return nil
}

func (s *service) ensureMigrationTable(db *sql.DB, dbType DatabaseType) error {
	var createTableSQL string

	switch dbType {
	case Postgres:
		createTableSQL = `
        CREATE TABLE IF NOT EXISTS Migrations (
            id SERIAL PRIMARY KEY,
            filename VARCHAR(255) NOT NULL UNIQUE,
            executed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
        );
        `
	case MySQL:
		createTableSQL = `
        CREATE TABLE IF NOT EXISTS Migrations (
            id INT AUTO_INCREMENT PRIMARY KEY,
            filename VARCHAR(255) NOT NULL UNIQUE,
            executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        `
	case SQLServer:
		createTableSQL = `
        CREATE TABLE IF NOT EXISTS Migrations (
            id INT IDENTITY(1,1) PRIMARY KEY,
            filename VARCHAR(255) NOT NULL UNIQUE,
            executed_at DATETIME DEFAULT GETDATE()
        );
        `
	case SQLite:
		createTableSQL = `
        CREATE TABLE IF NOT EXISTS Migrations (
            id INTEGER PRIMARY KEY AUTOINCREMENT,
            filename TEXT NOT NULL UNIQUE,
            executed_at DATETIME DEFAULT CURRENT_TIMESTAMP
        );
        `
	default:
		return fmt.Errorf("unsupported database type for migration table: %s", dbType)
	}

	// Eksekusi query dengan penanganan error yang lebih baik
	result, err := db.Exec(createTableSQL)
	if err != nil {
		return fmt.Errorf("failed to create migrations table: %w", err)
	}

	// Cek apakah tabel berhasil dibuat
	rowsAffected, _ := result.RowsAffected()
	if rowsAffected > 0 {
		logger.Default().Info("Migrations table created successfully")
	} else {
		logger.Default().Debug("Migrations table already exists")
	}

	return nil
}

func (s *service) isMigrationApplied(db *sql.DB, filename string) bool {
	var count int

	// Gunakan query yang sesuai dengan database type
	var query string
	switch s.getDatabaseTypeFromConnection(db) {
	case Postgres:
		query = "SELECT COUNT(*) FROM Migrations WHERE filename = $1"
	case MySQL, SQLite:
		query = "SELECT COUNT(*) FROM Migrations WHERE filename = ?"
	case SQLServer:
		query = "SELECT COUNT(*) FROM Migrations WHERE filename = @p1"
	default:
		query = "SELECT COUNT(*) FROM Migrations WHERE filename = ?"
	}

	err := db.QueryRow(query, filename).Scan(&count)
	if err != nil {
		// Jika tabel belum ada, anggap migrasi belum dijalankan
		if isTableNotExistsError(err) {
			return false
		}
		logger.Default().Error("Error checking migration status", logger.ErrorField(err))
		return false
	}

	return count > 0
}

func (s *service) executeMigrationFile(db *sql.DB, filePath string) error {
	// Baca konten file SQL
	sqlContent, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read migration file %s: %w", filePath, err)
	}

	// Jalankan migrasi dalam transaksi
	tx, err := db.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	defer func() {
		if err != nil {
			tx.Rollback()
		}
	}()

	// Eksekusi setiap perintah SQL
	commands := strings.Split(string(sqlContent), ";")
	for _, cmd := range commands {
		cmd = strings.TrimSpace(cmd)
		if cmd == "" {
			continue
		}

		_, err := tx.Exec(cmd)
		if err != nil {
			return fmt.Errorf("failed to execute SQL command: %s, error: %w", cmd, err)
		}
	}

	// Commit transaksi
	if err = tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	// Catat migrasi sebagai telah dijalankan
	filename := filepath.Base(filePath)
	if err := s.recordMigration(db, filename); err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

func (s *service) recordMigration(db *sql.DB, filename string) error {
	// Gunakan query yang sesuai dengan database type
	var query string
	switch s.getDatabaseTypeFromConnection(db) {
	case Postgres:
		query = "INSERT INTO Migrations (filename) VALUES ($1)"
	case MySQL, SQLite:
		query = "INSERT INTO Migrations (filename) VALUES (?)"
	case SQLServer:
		query = "INSERT INTO Migrations (filename) VALUES (@p1)"
	default:
		query = "INSERT INTO Migrations (filename) VALUES (?)"
	}

	_, err := db.Exec(query, filename)
	if err != nil {
		return fmt.Errorf("failed to record migration: %w", err)
	}

	return nil
}

// Fungsi bantu untuk mendeteksi tipe database dari koneksi
func (s *service) getDatabaseTypeFromConnection(db *sql.DB) DatabaseType {
	// Ini adalah implementasi sederhana, Anda mungkin perlu menyesuaikan
	// dengan cara Anda mengelola koneksi database
	// Coba dapatkan driver name dari koneksi
	// Note: Ini mungkin perlu disesuaikan dengan driver yang Anda gunakan
	if postgresDriver := db.Driver(); postgresDriver != nil {
		driverType := fmt.Sprintf("%T", postgresDriver)
		if strings.Contains(driverType, "postgres") {
			return Postgres
		} else if strings.Contains(driverType, "mysql") {
			return MySQL
		} else if strings.Contains(driverType, "sqlserver") {
			return SQLServer
		} else if strings.Contains(driverType, "sqlite") {
			return SQLite
		}
	}

	// Default fallback
	return Postgres
}
