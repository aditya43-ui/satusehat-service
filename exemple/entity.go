package report

// Employee merepresentasikan data dari PostgreSQL
type Employee struct {
	ID         int64  `gorm:"column:id"`
	Name       string `gorm:"column:name"`
	Department string `gorm:"column:department"`
}

// Attendance merepresentasikan data dari MySQL
type Attendance struct {
	EmployeeID int64  `gorm:"column:employee_id"`
	Status     string `gorm:"column:status"`
	Date       string `gorm:"column:date"`
}

// CombinedReport adalah hasil olahan (gabungan) yang akan dikembalikan ke Client
type CombinedReport struct {
	EmployeeName string       `json:"employee_name"`
	Department   string       `json:"department"`
	Attendances  []Attendance `json:"attendances"`
}
