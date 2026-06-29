package patient

// CreatePatientRequest adalah DTO internal yang lebih sederhana untuk diisi oleh Frontend.
type CreatePatientRequest struct {
	NIK       string `json:"nik" validate:"required"`
	Name      string `json:"name" validate:"required"`
	Gender    string `json:"gender" validate:"required"`     // male, female, other, unknown
	BirthDate string `json:"birth_date" validate:"required"` // Format: YYYY-MM-DD
	Phone     string `json:"phone"`                          // Opsional
	Address   string `json:"address"`                        // Opsional
}

// PatientSearchParams menampung berbagai kriteria pencarian pasien Satu Sehat.
type PatientSearchParams struct {
	Name      string `form:"name"`
	BirthDate string `form:"birthdate"` // Format: YYYY-MM-DD
	Gender    string `form:"gender"`    // male, female, other, unknown
	NIK       string `form:"nik"`
	NIKIbu    string `form:"nik_ibu"`
}
