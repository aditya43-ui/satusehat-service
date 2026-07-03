package servers

import (
	"service/internal/infrastructure/cache"
	"service/internal/infrastructure/config"
	"service/internal/infrastructure/database"

	"gorm.io/gorm"

	"service/internal/auth"
	roleMaster "service/internal/master/role/master"
	rolePages "service/internal/master/role/pages"
	rolePermission "service/internal/master/role/permission"
	satuSehatAuth "service/internal/satusehat/reference/auth"
	satuSehatKfa "service/internal/satusehat/reference/kfa"
	"service/internal/satusehat/reference/kyc"
	satuSehatLocation "service/internal/satusehat/reference/location"
	satuSehatOrganization "service/internal/satusehat/reference/organization"
	satuSehatPatient "service/internal/satusehat/reference/patient"
	satuSehatPractitioner "service/internal/satusehat/reference/practitioner"
	"service/internal/satusehat/usecase/allergyintolerance"
	"service/internal/satusehat/usecase/careplan"
	clinicalimpression "service/internal/satusehat/usecase/clinicalImpression"
	"service/internal/satusehat/usecase/composition"
	"service/internal/satusehat/usecase/condition"
	"service/internal/satusehat/usecase/diagnosticreport"
	"service/internal/satusehat/usecase/encounter"
	"service/internal/satusehat/usecase/episodeofcare"
	"service/internal/satusehat/usecase/imagingstudy"
	"service/internal/satusehat/usecase/immunization"
	"service/internal/satusehat/usecase/medication"
	"service/internal/satusehat/usecase/medicationdispense"
	"service/internal/satusehat/usecase/medicationrequest"
	"service/internal/satusehat/usecase/medicationstatement"
	"service/internal/satusehat/usecase/claim"
	"service/internal/satusehat/usecase/claimresponse"
	"service/internal/satusehat/usecase/observation"
	"service/internal/satusehat/usecase/purificationdecision"
	"service/internal/satusehat/usecase/procedure"
	"service/internal/satusehat/usecase/questionnaireresponse"
	"service/internal/satusehat/usecase/servicerequest"
	"service/internal/satusehat/usecase/specimen"
	"service/internal/satusehat/usecase/studies"
)

// MasterServices menampung kumpulan service untuk domain Master Data
type MasterServices struct {
	RolePages      rolePages.Service
	RolePermission rolePermission.Service
	RoleMaster     roleMaster.Service
}

// SatuSehatServices menampung kumpulan service untuk integrasi Kemenkes Satu Sehat
type SatuSehatServices struct {
	Auth         satuSehatAuth.Service
	Patient      satuSehatPatient.Service
	Practitioner satuSehatPractitioner.Service
	Organization satuSehatOrganization.Service
	Location     satuSehatLocation.Service
	KFA          satuSehatKfa.Service
	KYC          kyc.Service

	// Usecase Services
	AllergyIntolerance    allergyintolerance.Service
	CarePlan              careplan.Service
	ClinicalImpression    clinicalimpression.Service
	Composition           composition.Service
	Condition             condition.Service
	DiagnosticReport      diagnosticreport.Service
	Encounter             encounter.Service
	EpisodeOfCare         episodeofcare.Service
	ImagingStudy          imagingstudy.Service
	Immunization          immunization.Service
	Medication            medication.Service
	MedicationDispense    medicationdispense.Service
	MedicationRequest     medicationrequest.Service
	MedicationStatement   medicationstatement.Service
	Observation           observation.Service
	Procedure             procedure.Service
	Claim                 claim.Service
	ClaimResponse         claimresponse.Service
	PurificationDecision  purificationdecision.Service
	QuestionnaireResponse questionnaireresponse.Service
	ServiceRequest        servicerequest.Service
	Specimen              specimen.Service
	DicomStudies          studies.Service
}

// ServiceRegistry berfungsi sebagai Dependency Injection Container untuk transport HTTP.
// Struktur ini mencegah membengkaknya parameter pada saat inisialisasi API.
type ServiceRegistry struct {
	Config       *config.Config
	DBManager    database.Service
	PrimaryDB    *gorm.DB
	CacheManager *cache.Manager

	// Modul Aplikasi (Pisahkan berdasarkan Bounded Context)
	AuthService auth.Service
	Master      *MasterServices
	SatuSehat   *SatuSehatServices
}
