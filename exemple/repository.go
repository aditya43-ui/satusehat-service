package report

import (
	"context"
	"service/internal/infrastructure/database"
)

type Repository interface {
	GetEmployeesFromPostgres(ctx context.Context) ([]Employee, error)
	GetAttendancesFromMySQL(ctx context.Context, employeeIDs []int64) ([]Attendance, error)
}

type repository struct {
	dbManager database.Service
	pgDBName  string // Nama koneksi PostgreSQL di config.yaml
	mySQLName string // Nama koneksi MySQL di config.yaml
}

func NewRepository(dbManager database.Service, pgDBName, mySQLName string) Repository {
	return &repository{
		dbManager: dbManager,
		pgDBName:  pgDBName,
		mySQLName: mySQLName,
	}
}

// Mengakses PostgreSQL
func (r *repository) GetEmployeesFromPostgres(ctx context.Context) ([]Employee, error) {
	db, err := r.dbManager.GetGormDB(r.pgDBName)
	if err != nil {
		return nil, err
	}

	var employees []Employee
	err = db.WithContext(ctx).Table("employees").Find(&employees).Error
	return employees, err
}

// Mengakses MySQL
func (r *repository) GetAttendancesFromMySQL(ctx context.Context, employeeIDs []int64) ([]Attendance, error) {
	db, err := r.dbManager.GetGormDB(r.mySQLName)
	if err != nil {
		return nil, err
	}

	var attendances []Attendance
	// Hanya ambil attendance untuk employee yang ditemukan di Postgres
	err = db.WithContext(ctx).Table("attendances").Where("employee_id IN ?", employeeIDs).Find(&attendances).Error
	return attendances, err
}
