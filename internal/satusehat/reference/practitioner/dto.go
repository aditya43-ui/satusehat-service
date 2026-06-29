package practitioner

// PractitionerSearchParams menampung berbagai kriteria pencarian Tenaga Medis Satu Sehat.
type PractitionerSearchParams struct {
	Name      string `form:"name"`
	NIK       string `form:"nik"`
	Gender    string `form:"gender"`    // male, female
	BirthDate string `form:"birthdate"` // Format: YYYY-MM-DD
}
