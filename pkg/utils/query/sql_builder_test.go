package query

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSQLQueryBuilder_BuildWhereClause(t *testing.T) {
	// Menggunakan PostgreSQL sebagai dialek pengujian
	builder := NewSQLQueryBuilder(DBTypePostgreSQL)
	allowedColumns := []string{
		"age",
		"status",
		"category",
		"price",
		"is_featured",
		"id",
		"deleted_at",
		"updated_at",
		"created_at", // Diperlukan untuk perbandingan kolom-ke-kolom
		"name",       // Untuk test LIKE
		"rating",     // Untuk test BETWEEN
	}
	// Daftarkan semua kolom yang digunakan dalam tes sebagai kolom yang diizinkan.
	// Ini penting karena security check (validasi kolom) diaktifkan secara default.
	builder.SetAllowedColumns(allowedColumns)

	t.Run("should build single group with AND logic", func(t *testing.T) {
		filters := []FilterGroup{
			{
				Filters: []DynamicFilter{
					{Column: "age", Operator: OpGreaterThan, Value: 30},
					{Column: "status", Operator: OpEqual, Value: "active"},
				},
				LogicOp: "AND",
			},
		}

		sql, args, err := builder.BuildWhereClause(filters)

		assert.NoError(t, err)
		assert.Equal(t, `("age" > ? AND "status" = ?)`, sql)
		assert.Equal(t, []interface{}{30, "active"}, args)
	})

	t.Run("should build multiple groups with OR logic", func(t *testing.T) {
		// Skenario ini menguji: (category = 'electronics' AND price < 1000) OR (is_featured = true)
		filters := []FilterGroup{
			{ // Grup 1
				Filters: []DynamicFilter{
					{Column: "category", Operator: OpEqual, Value: "electronics"},
					{Column: "price", Operator: OpLessThan, Value: 1000},
				},
				LogicOp: "AND",
			},
			{ // Grup 2
				Filters: []DynamicFilter{
					{Column: "is_featured", Operator: OpEqual, Value: true},
				},
			},
		}

		sql, args, err := builder.BuildWhereClause(filters)

		assert.NoError(t, err)
		// Memastikan logika OR antar grup berjalan dengan benar
		assert.Equal(t, `("category" = ? AND "price" < ?) OR ("is_featured" = ?)`, sql)
		assert.Equal(t, []interface{}{"electronics", 1000, true}, args)
	})

	t.Run("should handle IN operator with slice", func(t *testing.T) {
		filters := []FilterGroup{
			{
				Filters: []DynamicFilter{
					{Column: "id", Operator: OpIn, Value: []interface{}{1, 2, 3}},
				},
			},
		}

		sql, args, err := builder.BuildWhereClause(filters)

		assert.NoError(t, err)
		assert.Equal(t, `("id" IN (?,?,?))`, sql)
		assert.Equal(t, []interface{}{1, 2, 3}, args)
	})

	t.Run("should handle IS NULL operator", func(t *testing.T) {
		filters := []FilterGroup{
			{
				Filters: []DynamicFilter{
					{Column: "deleted_at", Operator: OpNull, Value: nil},
				},
			},
		}

		sql, args, err := builder.BuildWhereClause(filters)

		assert.NoError(t, err)
		assert.Equal(t, `("deleted_at" IS NULL)`, sql)
		assert.Nil(t, args)
	})

	t.Run("should handle column-to-column comparison", func(t *testing.T) {
		filters := []FilterGroup{
			{
				Filters: []DynamicFilter{
					// Membandingkan kolom updated_at dengan kolom created_at
					{Column: "updated_at", Operator: OpGreaterThan, Value: "users.created_at"},
				},
			},
		}

		sql, args, err := builder.BuildWhereClause(filters)

		assert.NoError(t, err)
		assert.Equal(t, `("updated_at" > "users"."created_at")`, sql)
		assert.Nil(t, args, "Column-to-column comparison should not produce arguments")
	})

	t.Run("should handle LIKE and ILIKE operators", func(t *testing.T) {
		filters := []FilterGroup{
			{
				Filters: []DynamicFilter{
					{Column: "name", Operator: OpLike, Value: "%John%"},
					{Column: "category", Operator: OpILike, Value: "ELECTRONICS"},
				},
				LogicOp: "AND",
			},
		}

		// Test with PostgreSQL which supports ILIKE natively
		pgBuilder := NewSQLQueryBuilder(DBTypePostgreSQL).SetAllowedColumns(allowedColumns)
		sql, args, err := pgBuilder.BuildWhereClause(filters)
		assert.NoError(t, err)
		assert.Equal(t, `("name" LIKE ? AND "category" ILIKE ?)`, sql)
		assert.Equal(t, []interface{}{"%John%", "ELECTRONICS"}, args)

		// Test with MySQL which simulates ILIKE with LOWER()
		mysqlBuilder := NewSQLQueryBuilder(DBTypeMySQL).SetAllowedColumns(allowedColumns)
		sql, args, err = mysqlBuilder.BuildWhereClause(filters)
		assert.NoError(t, err)
		assert.Equal(t, "(`name` LIKE ? AND LOWER(`category`) LIKE LOWER(?))", sql)
		assert.Equal(t, []interface{}{"%John%", "ELECTRONICS"}, args)
	})

	t.Run("should handle BETWEEN operator", func(t *testing.T) {
		filters := []FilterGroup{
			{
				Filters: []DynamicFilter{
					{Column: "rating", Operator: OpBetween, Value: []interface{}{4.0, 5.0}},
				},
			},
		}
		// This test uses the original `builder` which is configured for PostgreSQL
		sql, args, err := builder.BuildWhereClause(filters)
		assert.NoError(t, err)
		assert.Equal(t, `("rating" BETWEEN ? AND ?)`, sql)
		assert.Equal(t, []interface{}{4.0, 5.0}, args)
	})
}
