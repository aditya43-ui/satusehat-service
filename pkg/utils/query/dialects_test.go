package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPostgreSQLDialect_EscapeIdentifier(t *testing.T) {
	dialect := NewPostgreSQLDialect()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple lowercase",
			input:    "users",
			expected: `"users"`,
		},
		{
			name:     "PascalCase with underscore",
			input:    "Nama_device",
			expected: `"Nama_device"`,
		},
		{
			name:     "PascalCase",
			input:    "Tahun",
			expected: `"Tahun"`,
		},
		{
			name:     "Schema and table",
			input:    "role_access.rol_permission",
			expected: `"role_access"."rol_permission"`,
		},
		{
			name:     "Wildcard should not be escaped",
			input:    "*",
			expected: `*`,
		},
		{
			name:     "Identifier with quotes inside",
			input:    `user"s`,
			expected: `"user""s"`,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := dialect.EscapeIdentifier(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestMySQLDialect_EscapeIdentifier(t *testing.T) {
	dialect := NewMySQLDialect()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple lowercase",
			input:    "users",
			expected: "`users`",
		},
		{
			name:     "PascalCase with underscore",
			input:    "Nama_device",
			expected: "`Nama_device`",
		},
		{
			name:     "PascalCase",
			input:    "Tahun",
			expected: "`Tahun`",
		},
		{
			name:     "Schema and table",
			input:    "role_access.rol_permission",
			expected: "`role_access`.`rol_permission`",
		},
		{
			name:     "Wildcard should not be escaped",
			input:    "*",
			expected: `*`,
		},
		{
			name:     "Identifier with backticks inside",
			input:    "user`s",
			expected: "`user``s`",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := dialect.EscapeIdentifier(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSQLServerDialect_EscapeIdentifier(t *testing.T) {
	dialect := NewSQLServerDialect()

	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Simple lowercase",
			input:    "users",
			expected: "[users]",
		},
		{
			name:     "PascalCase with underscore",
			input:    "Nama_device",
			expected: "[Nama_device]",
		},
		{
			name:     "PascalCase",
			input:    "Tahun",
			expected: "[Tahun]",
		},
		{
			name:     "Schema and table",
			input:    "role_access.rol_permission",
			expected: "[role_access].[rol_permission]",
		},
		{
			name:     "Wildcard should not be escaped",
			input:    "*",
			expected: `*`,
		},
		{
			name:     "Identifier with brackets inside",
			input:    "user]s",
			expected: "[user]]s]",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := dialect.EscapeIdentifier(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestSQLiteDialect_EscapeIdentifier(t *testing.T) {
	dialect := NewSQLiteDialect()

	// SQLite uses the same quoting as PostgreSQL
	t.Run("PascalCase with underscore", func(t *testing.T) {
		result := dialect.EscapeIdentifier("Nama_device")
		assert.Equal(t, `"Nama_device"`, result)
	})

	t.Run("Schema and table", func(t *testing.T) {
		result := dialect.EscapeIdentifier("role_access.rol_permission")
		assert.Equal(t, `"role_access"."rol_permission"`, result)
	})
}
