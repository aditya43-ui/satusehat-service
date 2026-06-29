package errors

import (
	"encoding/json"
	"fmt"
	"strings"
	"sync"
)

// MessageStore stores localized messages
type MessageStore struct {
	messages map[string]map[string]string // language -> code -> message
	mu       sync.RWMutex
}

var (
	messageStore = &MessageStore{
		messages: make(map[string]map[string]string),
	}
)

func init() {
	// Initialize default messages
	initializeDefaultMessages()
}

// initializeDefaultMessages sets up default error messages
func initializeDefaultMessages() {
	// English messages
	englishMessages := map[string]string{
		ErrCodeValidationFailed:    "Validation failed",
		ErrCodeInvalidInput:        "Invalid input provided",
		ErrCodeMissingField:        "Required field is missing",
		ErrCodeInvalidFormat:       "Invalid format",
		ErrCodeInvalidLength:       "Invalid length",
		ErrCodeInvalidRange:        "Value out of range",
		ErrCodeNotFound:            "Resource not found",
		ErrCodeUserNotFound:        "User not found",
		ErrCodeResourceNotFound:    "Resource not found",
		ErrCodeDataNotFound:        "Data not found",
		ErrCodeUnauthorized:        "Unauthorized access",
		ErrCodeInvalidToken:        "Invalid token provided",
		ErrCodeTokenExpired:        "Token has expired",
		ErrCodeInvalidCredentials:  "Invalid credentials",
		ErrCodeForbidden:           "Access forbidden",
		ErrCodeInsufficientRights:  "Insufficient permissions",
		ErrCodeAccessDenied:        "Access denied",
		ErrCodeConflict:            "Conflict occurred",
		ErrCodeDuplicateEntry:      "Duplicate entry",
		ErrCodeResourceLocked:      "Resource is locked",
		ErrCodeConcurrentUpdate:    "Concurrent update detected",
		ErrCodeRateLimitExceeded:   "Rate limit exceeded",
		ErrCodeTooManyRequests:     "Too many requests",
		ErrCodeQuotaExceeded:       "Quota exceeded",
		ErrCodeInternalError:       "Internal server error",
		ErrCodeUnexpectedError:     "An unexpected error occurred",
		ErrCodeServiceUnavailable:  "Service is unavailable",
		ErrCodeConfigurationError:  "Configuration error",
		ErrCodeExternalError:       "External service error",
		ErrCodeThirdPartyError:     "Third party service error",
		ErrCodeAPIError:            "API error occurred",
		ErrCodeBusinessRule:        "Business rule violation",
		ErrCodeInsufficientBalance: "Insufficient balance",
		ErrCodeAccountSuspended:    "Account suspended",
		ErrCodeTimeout:             "Request timeout",
		ErrCodeRequestTimeout:      "Request timeout",
		ErrCodeConnectionTimeout:   "Connection timeout",
		ErrCodeNetworkError:        "Network error",
		ErrCodeConnectionFailed:    "Connection failed",
		ErrCodeConnectionLost:      "Connection lost",
		ErrCodeDatabaseError:       "Database error",
		ErrCodeQueryFailed:         "Query execution failed",
		ErrCodeTransactionFailed:   "Transaction failed",
	}

	// Indonesian messages
	indonesianMessages := map[string]string{
		ErrCodeValidationFailed:    "Validasi gagal",
		ErrCodeInvalidInput:        "Input tidak valid",
		ErrCodeMissingField:        "Field wajib tidak ada",
		ErrCodeInvalidFormat:       "Format tidak valid",
		ErrCodeInvalidLength:       "Panjang tidak valid",
		ErrCodeInvalidRange:        "Nilai di luar jangkauan",
		ErrCodeNotFound:            "Sumber daya tidak ditemukan",
		ErrCodeUserNotFound:        "Pengguna tidak ditemukan",
		ErrCodeResourceNotFound:    "Sumber daya tidak ditemukan",
		ErrCodeDataNotFound:        "Data tidak ditemukan",
		ErrCodeUnauthorized:        "Akses tidak sah",
		ErrCodeInvalidToken:        "Token tidak valid",
		ErrCodeTokenExpired:        "Token telah kadaluarsa",
		ErrCodeInvalidCredentials:  "Kredensial tidak valid",
		ErrCodeForbidden:           "Akses dilarang",
		ErrCodeInsufficientRights:  "Izin tidak mencukupi",
		ErrCodeAccessDenied:        "Akses ditolak",
		ErrCodeConflict:            "Terjadi konflik",
		ErrCodeDuplicateEntry:      "Entri duplikat",
		ErrCodeResourceLocked:      "Sumber daya terkunci",
		ErrCodeConcurrentUpdate:    "Pembaruan bersamaan terdeteksi",
		ErrCodeRateLimitExceeded:   "Batas laju terlampaui",
		ErrCodeTooManyRequests:     "Terlalu banyak permintaan",
		ErrCodeQuotaExceeded:       "Kuota terlampaui",
		ErrCodeInternalError:       "Kesalahan server internal",
		ErrCodeUnexpectedError:     "Terjadi kesalahan tak terduga",
		ErrCodeServiceUnavailable:  "Layanan tidak tersedia",
		ErrCodeConfigurationError:  "Kesalahan konfigurasi",
		ErrCodeExternalError:       "Kesalahan layanan eksternal",
		ErrCodeThirdPartyError:     "Kesalahan layanan pihak ketiga",
		ErrCodeAPIError:            "Terjadi kesalahan API",
		ErrCodeBusinessRule:        "Pelanggaran aturan bisnis",
		ErrCodeInsufficientBalance: "Saldo tidak mencukupi",
		ErrCodeAccountSuspended:    "Akun ditangguhkan",
		ErrCodeTimeout:             "Permintaan timeout",
		ErrCodeRequestTimeout:      "Permintaan timeout",
		ErrCodeConnectionTimeout:   "Koneksi timeout",
		ErrCodeNetworkError:        "Kesalahan jaringan",
		ErrCodeConnectionFailed:    "Koneksi gagal",
		ErrCodeConnectionLost:      "Koneksi terputus",
		ErrCodeDatabaseError:       "Kesalahan database",
		ErrCodeQueryFailed:         "Eksekusi query gagal",
		ErrCodeTransactionFailed:   "Transaksi gagal",
	}

	// Add messages to store
	SetMessages("en", englishMessages)
	SetMessages("id", indonesianMessages)
}

// SetMessages sets messages for a language
func SetMessages(language string, messages map[string]string) {
	messageStore.mu.Lock()
	defer messageStore.mu.Unlock()

	if messageStore.messages[language] == nil {
		messageStore.messages[language] = make(map[string]string)
	}

	for code, message := range messages {
		messageStore.messages[language][code] = message
	}
}

// SetMessage sets a single message for a language
func SetMessage(language, code, message string) {
	messageStore.mu.Lock()
	defer messageStore.mu.Unlock()

	if messageStore.messages[language] == nil {
		messageStore.messages[language] = make(map[string]string)
	}

	messageStore.messages[language][code] = message
}

// GetLocalizedMessage returns localized message for error code
func GetLocalizedMessage(code, language, defaultMessage string) string {
	messageStore.mu.RLock()
	defer messageStore.mu.RUnlock()

	// Try specific language
	if langMessages, exists := messageStore.messages[language]; exists {
		if message, exists := langMessages[code]; exists {
			return message
		}
	}

	// Try English as fallback
	if langMessages, exists := messageStore.messages["en"]; exists {
		if message, exists := langMessages[code]; exists {
			return message
		}
	}

	// Return default message
	return defaultMessage
}

// GetSupportedLanguages returns all supported languages
func GetSupportedLanguages() []string {
	messageStore.mu.RLock()
	defer messageStore.mu.RUnlock()

	languages := make([]string, 0, len(messageStore.messages))
	for lang := range messageStore.messages {
		languages = append(languages, lang)
	}
	return languages
}

// LoadMessagesFromJSON loads messages from JSON file
func LoadMessagesFromJSON(language, jsonData string) error {
	var messages map[string]string
	if err := json.Unmarshal([]byte(jsonData), &messages); err != nil {
		return fmt.Errorf("failed to parse JSON messages: %w", err)
	}

	SetMessages(language, messages)
	return nil
}

// FormatMessage formats message with parameters
func FormatMessage(template string, params map[string]interface{}) string {
	result := template
	for key, value := range params {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
	}
	return result
}

// GetLocalizedMessageWithParams returns localized message with parameters
func GetLocalizedMessageWithParams(code, language, defaultMessage string, params map[string]interface{}) string {
	message := GetLocalizedMessage(code, language, defaultMessage)
	if params != nil {
		message = FormatMessage(message, params)
	}
	return message
}

// MessageBuilder builds localized messages
type MessageBuilder struct {
	language string
	code     string
	params   map[string]interface{}
}

// NewMessageBuilder creates a new message builder.
// Parameter 'lang' dianjurkan untuk dikirim berdasarkan context HTTP request
// misalnya dari header Accept-Language dari klien per request.
func NewMessageBuilder(lang string) *MessageBuilder {
	if lang == "" {
		lang = "en" // Fallback language default
	}
	return &MessageBuilder{
		language: lang,
		params:   make(map[string]interface{}),
	}
}

// Language sets language
func (mb *MessageBuilder) Language(lang string) *MessageBuilder {
	mb.language = lang
	return mb
}

// Code sets error code
func (mb *MessageBuilder) Code(code string) *MessageBuilder {
	mb.code = code
	return mb
}

// Param adds parameter
func (mb *MessageBuilder) Param(key string, value interface{}) *MessageBuilder {
	mb.params[key] = value
	return mb
}

// Params adds multiple parameters
func (mb *MessageBuilder) Params(params map[string]interface{}) *MessageBuilder {
	for k, v := range params {
		mb.params[k] = v
	}
	return mb
}

// Build builds the final message
func (mb *MessageBuilder) Build(defaultMessage string) string {
	return GetLocalizedMessageWithParams(mb.code, mb.language, defaultMessage, mb.params)
}

// ValidationMessages provides common validation messages
type ValidationMessages struct {
	Required     string
	Invalid      string
	TooShort     string
	TooLong      string
	InvalidEmail string
	InvalidPhone string
}

// GetValidationMessages returns validation messages for language
func GetValidationMessages(language string) ValidationMessages {
	return ValidationMessages{
		Required:     GetLocalizedMessage("VALIDATION_REQUIRED", language, "This field is required"),
		Invalid:      GetLocalizedMessage("VALIDATION_INVALID", language, "Invalid value"),
		TooShort:     GetLocalizedMessage("VALIDATION_TOO_SHORT", language, "Value is too short"),
		TooLong:      GetLocalizedMessage("VALIDATION_TOO_LONG", language, "Value is too long"),
		InvalidEmail: GetLocalizedMessage("VALIDATION_INVALID_EMAIL", language, "Invalid email format"),
		InvalidPhone: GetLocalizedMessage("VALIDATION_INVALID_PHONE", language, "Invalid phone format"),
	}
}

// AddValidationMessages adds validation messages for language
func AddValidationMessages(language string, messages ValidationMessages) {
	SetMessage(language, "VALIDATION_REQUIRED", messages.Required)
	SetMessage(language, "VALIDATION_INVALID", messages.Invalid)
	SetMessage(language, "VALIDATION_TOO_SHORT", messages.TooShort)
	SetMessage(language, "VALIDATION_TOO_LONG", messages.TooLong)
	SetMessage(language, "VALIDATION_INVALID_EMAIL", messages.InvalidEmail)
	SetMessage(language, "VALIDATION_INVALID_PHONE", messages.InvalidPhone)
}
