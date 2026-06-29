package report

import (
	"context"
	"service/pkg/errors"
)

type Service interface {
	GenerateCrossDBReport(ctx context.Context) ([]CombinedReport, error)
}

type service struct {
	repo Repository
}

func NewService(repo Repository) Service {
	return &service{repo: repo}
}

func (s *service) GenerateCrossDBReport(ctx context.Context) ([]CombinedReport, error) {
	// 1. Ambil data master karyawan dari PostgreSQL
	employees, err := s.repo.GetEmployeesFromPostgres(ctx)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to fetch employees from Postgres").Cause(err).Build()
	}

	if len(employees) == 0 {
		return []CombinedReport{}, nil
	}

	// 2. Ekstrak ID karyawan untuk klausa IN (?) di MySQL
	var empIDs []int64
	for _, emp := range employees {
		empIDs = append(empIDs, emp.ID)
	}

	// 3. Ambil data transaksi kehadiran dari MySQL menggunakan list ID di atas
	attendances, err := s.repo.GetAttendancesFromMySQL(ctx, empIDs)
	if err != nil {
		return nil, errors.InternalError().Message("Failed to fetch attendances from MySQL").Cause(err).Build()
	}

	// 4. Proses In-Memory Join (Mapping)
	// Kelompokkan attendance berdasarkan EmployeeID agar pencarian cepat O(1)
	attendanceMap := make(map[int64][]Attendance)
	for _, att := range attendances {
		attendanceMap[att.EmployeeID] = append(attendanceMap[att.EmployeeID], att)
	}

	// Gabungkan data
	var reports []CombinedReport
	for _, emp := range employees {
		reports = append(reports, CombinedReport{
			EmployeeName: emp.Name,
			Department:   emp.Department,
			Attendances:  attendanceMap[emp.ID], // Ambil relasinya dari memory map
		})
	}

	return reports, nil
}
