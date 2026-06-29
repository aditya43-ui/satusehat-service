package response

import (
	"fmt"
	"net/http"
	"service/pkg/errors"

	"github.com/gin-gonic/gin"
)

// Response represents a standard API response
type Response struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
	Error   interface{} `json:"error,omitempty"`
	Meta    interface{} `json:"meta,omitempty"`
}

// getErrorCodeAndCategory memetakan HTTP Status Code ke Error Code terpusat
func getErrorCodeAndCategory(statusCode int) (string, string) {
	switch statusCode {
	case http.StatusBadRequest:
		return errors.ErrCodeInvalidInput, errors.CategoryValidation
	case http.StatusUnauthorized:
		return errors.ErrCodeUnauthorized, errors.CategoryUnauthorized
	case http.StatusForbidden:
		return errors.ErrCodeForbidden, errors.CategoryForbidden
	case http.StatusNotFound:
		return errors.ErrCodeNotFound, errors.CategoryNotFound
	case http.StatusConflict:
		return errors.ErrCodeConflict, errors.CategoryConflict
	case http.StatusTooManyRequests:
		return errors.ErrCodeRateLimitExceeded, errors.CategoryRateLimit
	default:
		return errors.ErrCodeInternalError, errors.CategoryInternal
	}
}

// Success sends a success response
func Success(c *gin.Context, statusCode int, message string, data interface{}) {
	c.JSON(statusCode, Response{
		Status:  "success",
		Message: message,
		Data:    data,
	})
}

// Error sends an error response
func Error(c *gin.Context, statusCode int, message string, err interface{}) {
	var e error
	if err != nil {
		if errVal, ok := err.(error); ok {
			e = errVal
		} else {
			e = fmt.Errorf("%v", err)
		}
	} else {
		e = fmt.Errorf("%s", message)
	}

	// Jika error sudah berupa AppError dari pkg/errors, cetak langsung formatnya
	if errors.IsAppError(e) {
		errors.HandleHTTPError(c, e.(errors.Error))
		return
	}

	// Jika berupa error standard/native biasa, konversi menjadi format terpusat AppError
	code, category := getErrorCodeAndCategory(statusCode)
	wrappedErr := errors.NewBuilder().
		Code(code).
		Category(category).
		HTTPStatus(statusCode).
		Message(message).
		Cause(e).
		Metadata("reason", e.Error()).
		Build()

	errors.HandleHTTPError(c, wrappedErr)
}

// ErrorWithLog mengirimkan respons error HTTP sekaligus menyisipkan error asli
// ke dalam context Gin agar bisa direkam oleh LoggingMiddleware.
func ErrorWithLog(c *gin.Context, err error, statusCode int, message string, details interface{}) {
	if err != nil {
		c.Error(err) // Meneruskan error asli ke Gin Context untuk dicatat logger
	} else {
		err = fmt.Errorf("%s", message)
	}

	// Sama seperti fungsi Error(), kita teruskan ke format sentral
	if errors.IsAppError(err) {
		errors.HandleHTTPError(c, err.(errors.Error))
		return
	}

	code, category := getErrorCodeAndCategory(statusCode)
	builder := errors.NewBuilder().
		Code(code).
		Category(category).
		HTTPStatus(statusCode).
		Message(message).
		Cause(err).
		Metadata("reason", err.Error())

	// Sisipkan details tambahan ke dalam metadata error builder
	if details != nil {
		builder.Metadata("additional_details", details)
	}

	errors.HandleHTTPError(c, builder.Build())
}

// Meta contains pagination metadata
type Meta struct {
	Page       int `json:"page"`
	Limit      int `json:"limit"`
	Total      int `json:"total"`
	TotalPages int `json:"total_pages"`
}

// Paginated sends a paginated response
func Paginated(c *gin.Context, statusCode int, message string, data interface{}, meta Meta) {
	c.JSON(statusCode, Response{
		Status:  "success",
		Message: message,
		Data:    data,
		Meta:    meta,
	})
}

// =========================================================================
// BPJS FORMATTER
// =========================================================================

// BPJSMetaData merepresentasikan metadata standar dari API BPJS
type BPJSMetaData struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

// BPJSResponse merepresentasikan standar response API BPJS
type BPJSResponse struct {
	MetaData BPJSMetaData `json:"metaData"`
	Response interface{}  `json:"response,omitempty"`
}

// BPJS mengirimkan respons dengan format standar API BPJS
func BPJS(c *gin.Context, httpStatusCode int, bpjsCode string, message string, data interface{}) {
	c.JSON(httpStatusCode, BPJSResponse{
		MetaData: BPJSMetaData{
			Code:    bpjsCode, // Contoh: "200" (sukses), "201" (dibuat), atau kode error BPJS lainnya
			Message: message,
		},
		Response: data,
	})
}

// =========================================================================
// SATU SEHAT (HL7 FHIR) FORMATTER
// =========================================================================

// FHIROperationOutcome merepresentasikan struktur error standar HL7 FHIR
type FHIROperationOutcome struct {
	ResourceType string           `json:"resourceType"` // Harus selalu "OperationOutcome"
	Issue        []FHIRErrorIssue `json:"issue"`
}

// FHIRErrorIssue berisi detail issue untuk error Satu Sehat
type FHIRErrorIssue struct {
	Severity    string `json:"severity"`    // fatal | error | warning | information
	Code        string `json:"code"`        // invalid | security | exception | not-found | dll
	Diagnostics string `json:"diagnostics"` // Pesan detail / human-readable
}

// FHIR mengirimkan respons untuk resource FHIR secara langsung (standar Satu Sehat)
// Satu Sehat tidak menggunakan wrapper "data" atau "status", melainkan me-return object Resource langsung
func FHIR(c *gin.Context, statusCode int, resource interface{}) {
	c.JSON(statusCode, resource)
}

// FHIRError mengirimkan pesan error yang comply dengan format OperationOutcome FHIR
func FHIRError(c *gin.Context, statusCode int, severity, code, diagnostics string) {
	c.JSON(statusCode, FHIROperationOutcome{
		ResourceType: "OperationOutcome",
		Issue: []FHIRErrorIssue{
			{
				Severity:    severity,
				Code:        code,
				Diagnostics: diagnostics,
			},
		},
	})
}
