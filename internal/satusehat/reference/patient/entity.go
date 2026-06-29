package patient

// PatientEntity mewakili konstanta dan aturan domain (business rules) untuk Pasien Satu Sehat.
const (
	// Standar Identifier System Kemenkes
	IdentifierSystemNIK    = "https://fhir.kemkes.go.id/id/nik"
	IdentifierSystemNIKIbu = "https://fhir.kemkes.go.id/id/nik-ibu"
	IdentifierSystemIHS    = "https://fhir.kemkes.go.id/id/ihs-number"

	GenderMale    = "male"
	GenderFemale  = "female"
	GenderOther   = "other"
	GenderUnknown = "unknown"
)
