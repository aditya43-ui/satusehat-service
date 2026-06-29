package logger

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/sirupsen/logrus"
)

// Logger interface mendefinisikan kontrak untuk logging
type Logger interface {
	Debug(message string, fields ...Field)
	Info(message string, fields ...Field)
	Warn(message string, fields ...Field)
	Error(message string, fields ...Field)
	Fatal(message string, fields ...Field)
	WithContext(ctx context.Context) Logger
	WithFields(fields ...Field) Logger
	WithError(err error) Logger
}

// Field mendefinisikan struktur data untuk logging fields
type Field struct {
	Key   string
	Value interface{}
}

// Helper functions untuk membuat Field
func String(key, value string) Field {
	return Field{Key: key, Value: value}
}

func Int(key string, value int) Field {
	return Field{Key: key, Value: value}
}

func Int64(key string, value int64) Field {
	return Field{Key: key, Value: value}
}

func Float64(key string, value float64) Field {
	return Field{Key: key, Value: value}
}

func Bool(key string, value bool) Field {
	return Field{Key: key, Value: value}
}

func ErrorField(err error) Field {
	return Field{Key: "error", Value: err}
}

func Duration(key string, value time.Duration) Field {
	return Field{Key: key, Value: value}
}

func Time(key string, value time.Time) Field {
	return Field{Key: key, Value: value}
}

func Any(key string, value interface{}) Field {
	return Field{Key: key, Value: value}
}

// Config menyimpan konfigurasi untuk logger
type Config struct {
	Level            string
	Format           string
	Output           string
	FilePath         string
	MaxSize          int
	MaxBackups       int
	MaxAge           int
	Compress         bool
	EnableCaller     bool
	EnableStackTrace bool
	ServiceName      string
	Environment      string
}

// DefaultConfig mengembalikan konfigurasi default
func DefaultConfig() Config {
	return Config{
		Level:       "info",
		Format:      "json", // Format JSON menjadi default
		Output:      "both", // Default langsung di set ke 'both' agar console & file aktif
		ServiceName: "service-general",
		Environment: "development",
		MaxSize:     100, // MB
		MaxBackups:  3,
		MaxAge:      7, // days
		Compress:    true,
	}
}

type loggerImpl struct {
	entry        *logrus.Entry
	enableCaller bool
	mu           sync.RWMutex
}

var (
	defaultLogger Logger
	once          sync.Once
)

// Default mengembalikan logger default
func Default() Logger {
	if defaultLogger == nil {
		Init(DefaultConfig())
	}
	return defaultLogger
}

// New membuat instance logger baru yang independen (Ideal untuk Domain spesifik / CQRS)
func New(cfg Config) Logger {
	logrusLogger := logrus.New()

	// Set log level
	level, err := logrus.ParseLevel(cfg.Level)
	if err != nil {
		level = logrus.InfoLevel
	}
	logrusLogger.SetLevel(level)

	// Set formatter
	switch cfg.Format {
	case "json":
		logrusLogger.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat:   time.RFC3339,
			DisableHTMLEscape: true,
			PrettyPrint:       false,
		})
	case "text":
		logrusLogger.SetFormatter(&customFormatter{})
	default:
		logrusLogger.SetFormatter(&customFormatter{})
	}

	// Set output
	var output io.Writer
	switch cfg.Output {
	case "file":
		output = newDailyFileWriter("logs")
	case "both":
		fileWriter := newDailyFileWriter("logs")
		output = io.MultiWriter(os.Stdout, fileWriter)
	default:
		output = os.Stdout
	}
	logrusLogger.SetOutput(output)

	// Set default fields
	entry := logrusLogger.WithFields(logrus.Fields{
		"service":     cfg.ServiceName,
		"environment": cfg.Environment,
	})

	return &loggerImpl{
		entry:        entry,
		enableCaller: cfg.EnableCaller,
	}
}

// Init menginisialisasi logger global (default)
func Init(cfg Config) Logger {
	once.Do(func() {
		defaultLogger = New(cfg)
	})
	return defaultLogger
}

func (l *loggerImpl) WithContext(ctx context.Context) Logger {
	if ctx == nil {
		return l
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	fields := make(logrus.Fields)

	// Extract standard identifiers automatically
	keys := []string{"trace_id", "request_id", "correlation_id", "user_id"}
	for _, key := range keys {
		if val := ctx.Value(key); val != nil {
			fields[key] = val
		}
	}

	return &loggerImpl{
		entry:        l.entry.WithFields(fields),
		enableCaller: l.enableCaller,
	}
}

func (l *loggerImpl) WithFields(fields ...Field) Logger {
	l.mu.RLock()
	defer l.mu.RUnlock()

	logrusFields := make(logrus.Fields)
	for _, field := range fields {
		logrusFields[field.Key] = field.Value
	}

	return &loggerImpl{
		entry:        l.entry.WithFields(logrusFields),
		enableCaller: l.enableCaller,
	}
}

func (l *loggerImpl) WithError(err error) Logger {
	if err == nil {
		return l // Mencegah Panic bila log dipanggil dengan err == nil
	}

	l.mu.RLock()
	defer l.mu.RUnlock()

	fields := logrus.Fields{
		"error": err.Error(),
	}

	// Add stack trace if available
	if stackErr, ok := err.(interface{ StackTrace() []string }); ok {
		fields["stack_trace"] = stackErr.StackTrace()
	}

	return &loggerImpl{
		entry:        l.entry.WithFields(fields),
		enableCaller: l.enableCaller,
	}
}

func (l *loggerImpl) log(level logrus.Level, message string, fields ...Field) {
	l.mu.Lock()
	defer l.mu.Unlock()

	entry := l.entry

	// Add caller information (hanya jika EnableCaller dikonfigurasi aktif)
	if l.enableCaller {
		if pc, file, line, ok := runtime.Caller(2); ok {
			funcName := runtime.FuncForPC(pc).Name()
			entry = entry.WithFields(logrus.Fields{
				"caller":   fmt.Sprintf("%s:%d", filepath.Base(file), line),
				"function": funcName,
			})
		}
	}

	// Convert fields to logrus fields
	if len(fields) > 0 {
		logrusFields := make(logrus.Fields)
		for _, field := range fields {
			logrusFields[field.Key] = field.Value
		}
		entry = entry.WithFields(logrusFields)
	}

	switch level {
	case logrus.DebugLevel:
		entry.Debug(message)
	case logrus.InfoLevel:
		entry.Info(message)
	case logrus.WarnLevel:
		entry.Warn(message)
	case logrus.ErrorLevel:
		entry.Error(message)
	case logrus.FatalLevel:
		entry.Fatal(message)
	}
}

func (l *loggerImpl) Debug(message string, fields ...Field) {
	l.log(logrus.DebugLevel, message, fields...)
}

func (l *loggerImpl) Info(message string, fields ...Field) {
	l.log(logrus.InfoLevel, message, fields...)
}

func (l *loggerImpl) Warn(message string, fields ...Field) {
	l.log(logrus.WarnLevel, message, fields...)
}

func (l *loggerImpl) Error(message string, fields ...Field) {
	l.log(logrus.ErrorLevel, message, fields...)
}

func (l *loggerImpl) Fatal(message string, fields ...Field) {
	l.log(logrus.FatalLevel, message, fields...)
}

// Helper functions untuk backward compatibility
func fieldsToMap(fields ...Field) map[string]interface{} {
	result := make(map[string]interface{})
	for _, field := range fields {
		result[field.Key] = field.Value
	}
	return result
}

// Global functions untuk backward compatibility
func Debug(message string, fields map[string]interface{}) {
	logger := Default()
	var fieldSlice []Field
	for k, v := range fields {
		fieldSlice = append(fieldSlice, Any(k, v))
	}
	logger.Debug(message, fieldSlice...)
}

func Info(message string, fields map[string]interface{}) {
	logger := Default()
	var fieldSlice []Field
	for k, v := range fields {
		fieldSlice = append(fieldSlice, Any(k, v))
	}
	logger.Info(message, fieldSlice...)
}

func Warn(message string, fields map[string]interface{}) {
	logger := Default()
	var fieldSlice []Field
	for k, v := range fields {
		fieldSlice = append(fieldSlice, Any(k, v))
	}
	logger.Warn(message, fieldSlice...)
}

func Error(message string, fields map[string]interface{}) {
	logger := Default()
	var fieldSlice []Field
	for k, v := range fields {
		fieldSlice = append(fieldSlice, Any(k, v))
	}
	logger.Error(message, fieldSlice...)
}

func Fatal(message string, fields map[string]interface{}) {
	logger := Default()
	var fieldSlice []Field
	for k, v := range fields {
		fieldSlice = append(fieldSlice, Any(k, v))
	}
	logger.Fatal(message, fieldSlice...)
}

// ============================================================================
// CUSTOM DAILY DIRECTORY WRITER
// ============================================================================

// dailyFileWriter adalah custom io.Writer untuk menulis log ke dalam struktur folder bulanan (logs/YYYY/MM/YYYY-MM-DD.log)
type dailyFileWriter struct {
	mu       sync.Mutex
	basePath string
	currDate string
	file     *os.File
}

func newDailyFileWriter(basePath string) *dailyFileWriter {
	if basePath == "" {
		basePath = "logs"
	}
	return &dailyFileWriter{basePath: basePath}
}

func (w *dailyFileWriter) Write(p []byte) (n int, err error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	now := time.Now()
	dateStr := now.Format("2006-01-02") // Format harian (YYYY-MM-DD)

	// Cek jika hari berganti atau file belum terbuka
	if w.currDate != dateStr || w.file == nil {
		if w.file != nil {
			w.file.Close()
		}

		year := now.Format("2006")
		month := now.Format("01")

		// Buat struktur direktori logs/YYYY/MM
		dir := filepath.Join(w.basePath, year, month)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return 0, err
		}

		// Buka / Buat file logs/YYYY/MM/YYYY-MM-DD.log
		filename := filepath.Join(dir, dateStr+".log")
		file, err := os.OpenFile(filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			return 0, err
		}

		w.file = file
		w.currDate = dateStr
	}

	return w.file.Write(p)
}

func (w *dailyFileWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	if w.file != nil {
		return w.file.Close()
	}
	return nil
}

// ============================================================================
// CUSTOM LOG FORMATTER (READABLE TEXT FORMAT)
// ============================================================================

type customFormatter struct{}

func (f *customFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b *bytes.Buffer
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}

	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	level := strings.ToUpper(entry.Level.String())
	if len(level) > 5 {
		level = level[:5]
	}

	// 1. Base Format: [TIMESTAMP] [LEVEL] MESSAGE (Lebar di-fix agar sejajar)
	fmt.Fprintf(b, "[%s] [%-5s] %-55s", timestamp, level, entry.Message)

	// 2. Extract and Sort Keys agar log selalu konsisten urutannya
	keys := make([]string, 0, len(entry.Data))
	for k := range entry.Data {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	// 3. Append metadata dengan pembatas " | "
	for _, k := range keys {
		v := entry.Data[k]
		switch k {
		case "error":
			fmt.Fprintf(b, " | ❌ ERROR: %v", v)
		case "caller":
			fmt.Fprintf(b, " | 📍 %v", v)
		case "function":
			fmt.Fprintf(b, " | ⚙️ %v", v)
		default:
			fmt.Fprintf(b, " | %s: %v", k, v)
		}
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}
