package bpjs

import (
	"bytes"
	"compress/gzip"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"strings"
	"unicode"
	"unicode/utf8"

	lzstring "github.com/daku10/go-lz-string"
)

// DecompressPayload mengupayakan ekstraksi JSON menggunakan berbagai algoritma (GZIP, LZString, dll).
func DecompressPayload(data []byte) (string, error) {
	log.Printf("DecompressPayload: Attempting decompression, data length: %d", len(data))

	// Log hex dump for better debugging
	hexDump := make([]string, min(32, len(data)))
	for i := 0; i < len(hexDump); i++ {
		hexDump[i] = fmt.Sprintf("%02x", data[i])
	}
	log.Printf("DecompressPayload: Hex dump (first 32 bytes): %s", strings.Join(hexDump, " "))

	// Method 1: Try LZ-string first (most common for BPJS)
	if result, err := tryLZStringMethods(data); err == nil {
		log.Println("DecompressPayload: LZ-string decompression successful")
		return result, nil
	}

	// Method 2: Try gzip
	if result, err := tryGzipDecompression(data); err == nil && isValidDecompressedResult(result) {
		log.Println("DecompressPayload: Gzip decompression successful")
		return result, nil
	}

	// Method 3: Try as plain text
	if isValidUTF8AndPrintable(string(data)) {
		result := string(data)
		if isValidDecompressedResult(result) {
			log.Println("DecompressPayload: Data is already valid text")
			return result, nil
		}
	}

	// Method 4: Try base64 decode then decompress
	if result, err := tryBase64ThenDecompress(data); err == nil {
		log.Println("DecompressPayload: Base64 then decompress successful")
		return result, nil
	}

	log.Printf("DecompressPayload: All methods failed")
	return "", errors.New("all decompression methods failed")
}

// tryLZStringMethods attempts LZ-string decompression
func tryLZStringMethods(data []byte) (string, error) {
	dataStr := string(data)
	log.Printf("tryLZStringMethods: Raw data length: %d", len(dataStr))

	// Method 1: Clean corrupt prefix and find LZ-string pattern
	cleanedData := extractCleanLZString(dataStr)
	if cleanedData != "" {
		log.Printf("tryLZStringMethods: Found clean LZ-string: %s", cleanedData[:min(50, len(cleanedData))])

		// Decompress according to BPJS standards
		if result, err := lzstring.DecompressFromEncodedURIComponent(cleanedData); err == nil && len(result) > 0 {
			if isValidDecompressedResult(result) {
				log.Printf("LZ-string decompression successful, length: %d", len(result))
				return result, nil
			}
		}
	}

	// Method 2: Fallback direct decompression
	if result, err := lzstring.DecompressFromEncodedURIComponent(dataStr); err == nil && len(result) > 0 {
		if isValidDecompressedResult(result) {
			return result, nil
		}
	}

	// Method 3: Try base64 LZ-string
	if result, err := lzstring.DecompressFromBase64(dataStr); err == nil && len(result) > 0 {
		if isValidDecompressedResult(result) {
			return result, nil
		}
	}

	return "", errors.New("LZ-string decompression failed")
}

// extractCleanLZString extracts clean LZ-string from corrupt data
func extractCleanLZString(data string) string {
	// Common LZ-string patterns from BPJS documentation
	patterns := []string{"EAuUA", "N4Ig", "BwIw", "CwIw", "DwIw", "EwIw", "FwIw", "GwIw", "HwIw"}

	for _, pattern := range patterns {
		if idx := strings.Index(data, pattern); idx >= 0 {
			// Extract from pattern to end
			candidate := data[idx:]
			log.Printf("extractCleanLZString: Found pattern '%s' at position %d", pattern, idx)

			// Clean only valid base64 characters
			cleaned := extractBase64Only(candidate)
			if len(cleaned) > 100 { // Minimum length for valid data
				return cleaned
			}
		}
	}

	return ""
}

// extractBase64Only extracts only base64 valid characters
func extractBase64Only(s string) string {
	base64Chars := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	var result strings.Builder

	for _, char := range s {
		if strings.ContainsRune(base64Chars, char) {
			result.WriteRune(char)
		} else {
			// Stop at non-base64 character if already long enough
			if result.Len() > 100 {
				break
			}
		}
	}

	return result.String()
}

// tryGzipDecompression attempts gzip decompression
func tryGzipDecompression(data []byte) (string, error) {
	reader, err := gzip.NewReader(bytes.NewReader(data))
	if err != nil {
		return "", err
	}
	defer reader.Close()

	decompressed, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}

	return string(decompressed), nil
}

// tryBase64ThenDecompress attempts base64 decode then decompress
func tryBase64ThenDecompress(data []byte) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(string(data))
	if err != nil {
		return "", err
	}

	return tryLZStringMethods(decoded)
}

// isValidDecompressedResult validates decompressed result
func isValidDecompressedResult(result string) bool {
	if len(result) == 0 {
		return false
	}

	// Trim whitespace and check UTF-8
	trimmed := strings.TrimSpace(result)
	if !utf8.ValidString(trimmed) {
		return false
	}

	// Must start with { or [ for JSON
	if len(trimmed) > 0 && (trimmed[0] == '{' || trimmed[0] == '[') {
		// Validate as JSON
		var js json.RawMessage
		if json.Unmarshal([]byte(result), &js) == nil {
			log.Printf("Decompressed result is valid JSON, length: %d", len(result))
			return true
		}
	}

	// If not JSON, reject
	log.Printf("Decompressed result is not valid JSON")
	return false
}

// isValidUTF8AndPrintable checks if string is valid UTF-8 and printable
func isValidUTF8AndPrintable(s string) bool {
	if !utf8.ValidString(s) {
		log.Printf("isValidUTF8AndPrintable: String is not valid UTF-8")
		return false
	}

	// Count valid characters
	validChars := 0
	totalChars := 0

	for _, r := range s {
		totalChars++

		if r >= 32 && r <= 126 { // Printable ASCII
			validChars++
		} else if r == '\n' || r == '\r' || r == '\t' { // Allowed control chars
			validChars++
		} else if r >= 160 { // Unicode characters
			validChars++
		}
	}

	validRatio := float64(validChars) / float64(totalChars)
	log.Printf("isValidUTF8AndPrintable: Valid chars ratio: %.2f (%d/%d)", validRatio, validChars, totalChars)

	// At least 70% should be valid characters
	return validRatio >= 0.7
}

// hasLZStringPattern detects LZ-string patterns
func hasLZStringPattern(s string) bool {
	if len(s) < 10 {
		return false
	}

	// Common LZ-string compressed data patterns
	commonLZPatterns := []string{
		"N4Ig", "BwIw", "CwIw", "DwIw", "EwIw", "FwIw", "GwIw", "HwIw",
		"IwIw", "JwIw", "KwIw", "LwIw", "MwIw", "NwIw", "OwIw", "PwIw",
	}

	for _, pattern := range commonLZPatterns {
		if strings.HasPrefix(s, pattern) {
			return true
		}
	}

	// Check if string contains only base64 characters without spaces or newlines
	base64Pattern := "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789+/="
	if len(s) > 50 { // Only check long strings
		invalidChars := 0
		for _, char := range s {
			if !strings.ContainsRune(base64Pattern, char) {
				invalidChars++
			}
		}
		// If less than 5% invalid characters, likely LZ-string
		if float64(invalidChars)/float64(len(s)) < 0.05 {
			return true
		}
	}

	return false
}

// cleanResponse cleans response string
func cleanResponse(s string) string {
	// Remove UTF-8 BOM and other BOM variations
	s = strings.TrimPrefix(s, "\xef\xbb\xbf") // UTF-8 BOM
	s = strings.TrimPrefix(s, "\ufeff")       // Unicode BOM
	s = strings.TrimPrefix(s, "\ufffe")       // Unicode BOM (reverse)
	s = strings.TrimPrefix(s, "\xff\xfe")     // UTF-16 LE BOM
	s = strings.TrimPrefix(s, "\xfe\xff")     // UTF-16 BE BOM

	// Remove control and non-printable characters
	var result strings.Builder
	for _, r := range s {
		if r >= 32 && r <= 126 || r == '\n' || r == '\r' || r == '\t' {
			result.WriteRune(r)
		} else if r > 126 && unicode.IsPrint(r) {
			// Allow Unicode printable characters
			result.WriteRune(r)
		}
		// Skip all other characters (including BOM fragments)
	}

	cleaned := result.String()
	cleaned = strings.TrimSpace(cleaned)

	// Find and extract valid JSON
	if idx := strings.Index(cleaned, "{"); idx >= 0 {
		cleaned = cleaned[idx:]
		// Find matching closing brace
		if endIdx := findMatchingBrace(cleaned); endIdx > 0 {
			cleaned = cleaned[:endIdx+1]
		}
	}

	log.Printf("cleanResponse: Final cleaned length: %d", len(cleaned))
	log.Printf("cleanResponse: Final result preview: %s", cleaned[:min(200, len(cleaned))])
	return cleaned
}

// findMatchingBrace finds matching closing brace
func findMatchingBrace(s string) int {
	if len(s) == 0 || s[0] != '{' {
		return -1
	}

	braceCount := 0
	inString := false
	escaped := false

	for i, char := range s {
		if escaped {
			escaped = false
			continue
		}

		if char == '\\' {
			escaped = true
			continue
		}

		if char == '"' && !escaped {
			inString = !inString
			continue
		}

		if !inString {
			if char == '{' {
				braceCount++
			} else if char == '}' {
				braceCount--
				if braceCount == 0 {
					return i
				}
			}
		}
	}

	return -1
}
