package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"service/internal/infrastructure/database"
	"service/pkg/utils/query"

	"github.com/jmoiron/sqlx"
	"gorm.io/gorm"
)

type CommandRepository interface {
	CreateUser(ctx context.Context, user *User) error
	UpdateUser(ctx context.Context, user *User) error
	ChangePasswordInTx(ctx context.Context, userID int64, newPassword string) error
	RevokeAndSaveTokens(ctx context.Context, oldToken string, newRefreshToken string) error
}

type QueryRepository interface {
	FindUserByEmail(ctx context.Context, email string) (*User, error)
	FindUserByID(ctx context.Context, id int64) (*User, error)
	FindRefreshToken(ctx context.Context, token string) (*StoredToken, error)
}

type repository struct {
	dbManager database.Service
	dbName    string
	qb        query.QueryBuilder
	dbType    query.DBType
}

func newRepository(dbManager database.Service, dbName string) *repository {
	dbType := query.DBTypePostgreSQL
	allowedColumns := []string{
		"Id",
		"Email",
		"Password",
		"RoleID",
		"Active",
		"CreatedAt",
		"UpdatedAt",
	}
	qb := query.NewSQLQueryBuilder(dbType).
		SetSecurityOptions(true, 1000).
		SetQueryLogging(true).
		SetQueryTimeout(30).
		SetAllowedColumns(allowedColumns)
	return &repository{dbManager: dbManager, dbName: dbName, qb: qb, dbType: dbType}
}

func NewCommandRepository(dbManager database.Service, dbName string) CommandRepository {
	return newRepository(dbManager, dbName)
}

func NewQueryRepository(dbManager database.Service, dbName string) QueryRepository {
	return newRepository(dbManager, dbName)
}

func (r *repository) getWriteGormDB() (*gorm.DB, error) {
	return r.dbManager.GetGormDB(r.dbName)
}

func (r *repository) getReadSQLXDB() (*sqlx.DB, error) {
	sqlDB, err := r.dbManager.GetReadDB(r.dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get read db: %w", err)
	}
	return sqlx.NewDb(sqlDB, "pgx"), nil
}

func (r *repository) getWriteSQLXDB() (*sqlx.DB, error) {
	// Mengambil koneksi database utama (tulis)
	sqlDB, err := r.dbManager.GetDB(r.dbName)
	if err != nil {
		return nil, fmt.Errorf("failed to get write db: %w", err)
	}
	return sqlx.NewDb(sqlDB, "pgx"), nil
}

func (r *repository) CreateUser(ctx context.Context, user *User) error {
	db, err := r.getWriteGormDB()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Create(user).Error
}

func (r *repository) UpdateUser(ctx context.Context, user *User) error {
	db, err := r.getWriteGormDB()
	if err != nil {
		return err
	}
	return db.WithContext(ctx).Save(user).Error
}

func (r *repository) FindUserByEmail(ctx context.Context, email string) (*User, error) {
	sqlxDB, err := r.getReadSQLXDB()
	if err != nil {
		return nil, err
	}

	var user User
	q := query.DynamicQuery{
		From: "users", // GORM default table name for User struct
		Filters: []query.FilterGroup{{
			Filters: []query.DynamicFilter{
				query.CreateEqualFilter("Email", email),
			},
		}},
		Limit: 1,
	}

	if err := r.qb.ExecuteQueryRow(ctx, sqlxDB, q, &user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found is not an error, service layer will handle it
		}
		return nil, fmt.Errorf("failed to find user by email: %w", err)
	}
	return &user, nil
}

func (r *repository) ChangePasswordInTx(ctx context.Context, userID int64, newPassword string) (err error) {
	// 1. Dapatkan koneksi database tulis
	writeDB, err := r.getWriteSQLXDB()
	if err != nil {
		return err
	}

	// 2. Mulai transaksi
	tx, err := writeDB.Beginx()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// 3. Defer Rollback. Ini akan dieksekusi jika fungsi return error atau panic.
	defer func() {
		if p := recover(); p != nil {
			tx.Rollback()
			panic(p) // Lanjutkan panic setelah rollback
		} else if err != nil {
			if rbErr := tx.Rollback(); rbErr != nil {
				err = fmt.Errorf("transaction error: %v, rollback error: %v", err, rbErr)
			}
		}
	}()

	// 4. Operasi 1: Update password pengguna
	updateData := query.UpdateData{
		Columns: []string{"Password", "UpdatedAt"},
		Values:  []interface{}{newPassword, time.Now()},
	}
	filters := []query.FilterGroup{{
		Filters: []query.DynamicFilter{query.CreateEqualFilter("Id", userID)},
	}}

	// Berikan objek transaksi 'tx' ke Query Builder
	if _, err = r.qb.ExecuteUpdate(ctx, tx, "users", updateData, filters); err != nil {
		return fmt.Errorf("failed to update password in tx: %w", err)
	}

	// 5. Operasi 2: Catat aktivitas
	logData := query.InsertData{
		Columns: []string{"user_id", "activity", "created_at"},
		Values:  []interface{}{userID, "password changed via transaction", time.Now()},
	}

	// Gunakan objek transaksi 'tx' yang sama
	if _, err = r.qb.ExecuteInsert(ctx, tx, "user_activity_logs", logData); err != nil {
		// Asumsikan tabel 'user_activity_logs' ada
		return fmt.Errorf("failed to log activity in tx: %w", err)
	}

	// 6. Jika semua berhasil, commit transaksi
	return tx.Commit()
}

func (r *repository) FindUserByID(ctx context.Context, id int64) (*User, error) {
	sqlxDB, err := r.getReadSQLXDB()
	if err != nil {
		return nil, err
	}

	var user User
	q := query.DynamicQuery{
		From: "users", // GORM default table name for User struct
		Filters: []query.FilterGroup{{
			Filters: []query.DynamicFilter{
				query.CreateEqualFilter("Id", id),
			},
		}},
		Limit: 1,
	}

	if err := r.qb.ExecuteQueryRow(ctx, sqlxDB, q, &user); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // Not found is not an error, service layer will handle it
		}
		return nil, fmt.Errorf("failed to find user by id: %w", err)
	}
	return &user, nil
}

func (r *repository) FindRefreshToken(ctx context.Context, token string) (*StoredToken, error) {
	sqlxDB, err := r.getReadSQLXDB()
	if err != nil {
		return nil, err
	}

	var stored StoredToken
	q := query.DynamicQuery{
		From: "refresh_tokens",
		Filters: []query.FilterGroup{{
			Filters: []query.DynamicFilter{query.CreateEqualFilter("Token", token)},
		}},
		Limit: 1,
	}

	if err := r.qb.ExecuteQueryRow(ctx, sqlxDB, q, &stored); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, fmt.Errorf("failed to find refresh token: %w", err)
	}
	return &stored, nil
}

func (r *repository) RevokeAndSaveTokens(ctx context.Context, oldToken string, newRefreshToken string) error {
	// Implementasi sederhana rotasi token
	writeDB, err := r.getWriteGormDB()
	if err != nil {
		return err
	}

	// TODO: Sesuaikan dengan struktur tabel Refresh Token yang Anda miliki di database
	// Ini adalah stub untuk memenuhi kontrak interface
	return writeDB.WithContext(ctx).Exec("UPDATE refresh_tokens SET is_revoked = ? WHERE token = ?", true, oldToken).Error
}
