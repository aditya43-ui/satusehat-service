// internal/infrastructure/transport/http/servers/server.go
package servers

import (
	"context" // PERBAIKAN: Tambahkan import untuk graceful shutdown
	"fmt"
	"log"
	"net/http"
	"time"

	"service/internal/infrastructure/cache" // PERBAIKAN: Tambahkan import cache
	"service/internal/infrastructure/config"
	"service/internal/infrastructure/transport/http/middleware"
	"service/internal/infrastructure/transport/http/routes"

	authHttp "service/internal/infrastructure/transport/http/handlers/main/auth"
	healthHttp "service/internal/infrastructure/transport/http/handlers/main/health"
	roleHttp "service/internal/infrastructure/transport/http/handlers/main/master/roles"
	satuSehatRefHttp "service/internal/infrastructure/transport/http/handlers/satusehat/reference"
	satuSehatUsecaseHttp "service/internal/infrastructure/transport/http/handlers/satusehat/usecase"

	// Inisialisasi pkg
	"service/pkg/errors"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

type Server struct {
	engine      *gin.Engine
	config      *config.ServerRESTConfig
	cacheConfig cache.CacheConfig // PERBAIKAN: Tambahkan cacheConfig
	httpServer  *http.Server      // PERBAIKAN: Simpan instance http.Server untuk graceful shutdown
}

// NewHTTPServer membuat instance server HTTP baru
func NewHTTPServer(
	restConfig *config.ServerRESTConfig,
	registry *ServiceRegistry,
) *Server {
	if !restConfig.Enabled {
		log.Println("HTTP server is disabled in configuration.")
		return &Server{
			config: restConfig,
		}
	}

	// Set mode Gin berdasarkan konfigurasi global
	if registry.Config.Server.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()

	// Middleware untuk recovery dan penanganan error konsisten.
	// Diletakkan paling awal untuk menangkap panic dari middleware/handler lain.
	engine.Use(errors.CombinedMiddleware())

	// Middleware lainnya
	engine.Use(middleware.LoggingMiddleware())
	engine.Use(middleware.SecurityMiddleware())
	engine.Use(middleware.CORSMiddleware())
	// PERBAIKAN: Berikan cacheManager ke RateLimitMiddleware
	engine.Use(middleware.RateLimitMiddleware(registry.CacheManager))

	// Inisialisasi ModuleHandlers secara dinamis berdasarkan fitur yang aktif
	appHandlers := &routes.ModuleHandlers{}

	if registry.AuthService != nil {
		appHandlers.Auth = authHttp.NewAuthHandler(registry.AuthService)
	}

	// Inisialisasi Handlers Modul Master
	if registry.Master != nil {
		if registry.Master.RolePages != nil {
			appHandlers.RolePages = roleHttp.NewRolPagesHandler(registry.Master.RolePages)
		}
		if registry.Master.RolePermission != nil {
			appHandlers.RolePermission = roleHttp.NewRolPermissionHandler(registry.Master.RolePermission)
		}
		if registry.Master.RoleMaster != nil {
			appHandlers.RoleMaster = roleHttp.NewRoleMasterHandler(registry.Master.RoleMaster)
		}
	}

	// Inisialisasi Handlers Modul Satu Sehat
	if registry.SatuSehat != nil {
		if registry.SatuSehat.Auth != nil {
			appHandlers.SatuSehatAuth = satuSehatRefHttp.NewAuthHandler(registry.SatuSehat.Auth)
		}
		if registry.SatuSehat.Patient != nil {
			appHandlers.SatuSehatPatient = satuSehatRefHttp.NewPatientHandler(registry.SatuSehat.Patient)
		}
		if registry.SatuSehat.Practitioner != nil {
			appHandlers.SatuSehatPractitioner = satuSehatRefHttp.NewPractitionerHandler(registry.SatuSehat.Practitioner)
		}
		if registry.SatuSehat.Organization != nil {
			appHandlers.SatuSehatOrganization = satuSehatRefHttp.NewOrganizationHandler(registry.SatuSehat.Organization)
		}
		if registry.SatuSehat.Location != nil {
			appHandlers.SatuSehatLocation = satuSehatRefHttp.NewLocationHandler(registry.SatuSehat.Location)
		}
		if registry.SatuSehat.KFA != nil {
			appHandlers.SatuSehatKFA = satuSehatRefHttp.NewKFAHandler(registry.SatuSehat.KFA)
		}
		if registry.SatuSehat.KYC != nil {
			appHandlers.SatuSehatKYC = satuSehatRefHttp.NewKYCHandler(registry.SatuSehat.KYC)
		}

		// Inisialisasi Handlers Modul Satu Sehat Usecase
		if registry.SatuSehat.AllergyIntolerance != nil {
			appHandlers.SSAllergyIntolerance = satuSehatUsecaseHttp.NewAllergyIntoleranceHandler(registry.SatuSehat.AllergyIntolerance)
		}
		if registry.SatuSehat.CarePlan != nil {
			appHandlers.SSCarePlan = satuSehatUsecaseHttp.NewCarePlanHandler(registry.SatuSehat.CarePlan)
		}
		if registry.SatuSehat.ClinicalImpression != nil {
			appHandlers.SSClinicalImpression = satuSehatUsecaseHttp.NewClinicalImpressionHandler(registry.SatuSehat.ClinicalImpression)
		}
		if registry.SatuSehat.Composition != nil {
			appHandlers.SSComposition = satuSehatUsecaseHttp.NewCompositionHandler(registry.SatuSehat.Composition)
		}
		if registry.SatuSehat.Condition != nil {
			appHandlers.SSCondition = satuSehatUsecaseHttp.NewConditionHandler(registry.SatuSehat.Condition)
		}
		if registry.SatuSehat.DiagnosticReport != nil {
			appHandlers.SSDiagnosticReport = satuSehatUsecaseHttp.NewDiagnosticReportHandler(registry.SatuSehat.DiagnosticReport)
		}
		if registry.SatuSehat.Encounter != nil {
			appHandlers.SSEncounter = satuSehatUsecaseHttp.NewEncounterHandler(registry.SatuSehat.Encounter)
		}
		if registry.SatuSehat.EpisodeOfCare != nil {
			appHandlers.SSEpisodeOfCare = satuSehatUsecaseHttp.NewEpisodeOfCareHandler(registry.SatuSehat.EpisodeOfCare)
		}
		if registry.SatuSehat.ImagingStudy != nil {
			appHandlers.SSImagingStudy = satuSehatUsecaseHttp.NewImagingStudyHandler(registry.SatuSehat.ImagingStudy)
		}
		if registry.SatuSehat.Immunization != nil {
			appHandlers.SSImmunization = satuSehatUsecaseHttp.NewImmunizationHandler(registry.SatuSehat.Immunization)
		}
		if registry.SatuSehat.Medication != nil {
			appHandlers.SSMedication = satuSehatUsecaseHttp.NewMedicationHandler(registry.SatuSehat.Medication)
		}
		if registry.SatuSehat.MedicationDispense != nil {
			appHandlers.SSMedicationDispense = satuSehatUsecaseHttp.NewMedicationDispenseHandler(registry.SatuSehat.MedicationDispense)
		}
		if registry.SatuSehat.MedicationRequest != nil {
			appHandlers.SSMedicationRequest = satuSehatUsecaseHttp.NewMedicationRequestHandler(registry.SatuSehat.MedicationRequest)
		}
		if registry.SatuSehat.MedicationStatement != nil {
			appHandlers.SSMedicationStatement = satuSehatUsecaseHttp.NewMedicationStatementHandler(registry.SatuSehat.MedicationStatement)
		}
		if registry.SatuSehat.Observation != nil {
			appHandlers.SSObservation = satuSehatUsecaseHttp.NewObservationHandler(registry.SatuSehat.Observation)
		}
		if registry.SatuSehat.Procedure != nil {
			appHandlers.SSProcedure = satuSehatUsecaseHttp.NewProcedureHandler(registry.SatuSehat.Procedure)
		}
		if registry.SatuSehat.Claim != nil {
			appHandlers.SSClaim = satuSehatUsecaseHttp.NewClaimHandler(registry.SatuSehat.Claim)
		}
		if registry.SatuSehat.ClaimResponse != nil {
			appHandlers.SSClaimResponse = satuSehatUsecaseHttp.NewClaimResponseHandler(registry.SatuSehat.ClaimResponse)
		}
		if registry.SatuSehat.PurificationDecision != nil {
			appHandlers.SSPurificationDecision = satuSehatUsecaseHttp.NewPurificationDecisionHandler(registry.SatuSehat.PurificationDecision)
		}
		if registry.SatuSehat.QuestionnaireResponse != nil {
			appHandlers.SSQuestionnaireResponse = satuSehatUsecaseHttp.NewQuestionnaireResponseHandler(registry.SatuSehat.QuestionnaireResponse)
		}
		if registry.SatuSehat.ServiceRequest != nil {
			appHandlers.SSServiceRequest = satuSehatUsecaseHttp.NewServiceRequestHandler(registry.SatuSehat.ServiceRequest)
		}
		if registry.SatuSehat.Specimen != nil {
			appHandlers.SSSpecimen = satuSehatUsecaseHttp.NewSpecimenHandler(registry.SatuSehat.Specimen)
		}
		if registry.SatuSehat.DicomStudies != nil {
			appHandlers.SSDicomStudies = satuSehatUsecaseHttp.NewDicomStudiesHandler(registry.SatuSehat.DicomStudies)
		}
	}

	// Inisialisasi Health Handler dengan dependensi yang benar
	// Gunakan constructor baru yang menerima cache manager
	healthHandler := healthHttp.NewHealthHandlerWithCache(registry.PrimaryDB, registry.CacheManager, registry.Config, registry.DBManager)

	// Setup routes menggunakan handler yang sudah diinisialisasi
	routes.SetupRoutes(
		engine,
		registry.Config,
		registry.CacheManager,
		healthHandler,
		appHandlers,
	)

	return &Server{
		engine: engine,
		config: restConfig,
	}
}

// RegisterSwagger mendaftarkan dokumentasi Swagger UI
func (s *Server) RegisterSwagger() {
	if s.engine == nil {
		log.Println("Cannot register Swagger: engine is nil (server disabled)")
		return
	}
	s.engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))
}

// Start menjalankan server HTTP
func (s *Server) Start(globalConfig *config.ServerConfig) error {
	if !s.config.Enabled {
		log.Println("HTTP server is disabled, not starting.")
		return nil
	}

	addr := fmt.Sprintf(":%d", s.config.Port)
	log.Printf("Starting HTTP server on %s", addr)

	// PERBAIKAN: Simpan instance server ke struct
	s.httpServer = &http.Server{
		Addr:         addr,
		Handler:      s.engine,
		ReadTimeout:  time.Duration(globalConfig.ReadTimeout) * time.Second,
		WriteTimeout: time.Duration(globalConfig.WriteTimeout) * time.Second,
		IdleTimeout:  60 * time.Second, // Good practice to have an idle timeout
	}

	return s.httpServer.ListenAndServe()
}

// PERBAIKAN: Tambahkan method untuk graceful shutdown
// Shutdown memberhentikan server HTTP dengan graceful shutdown
func (s *Server) Shutdown(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	log.Println("Shutting down HTTP server...")
	return s.httpServer.Shutdown(ctx)
}

// GetEngine mengembalikan instance Gin engine (terutama untuk testing)
func (s *Server) GetEngine() *gin.Engine {
	return s.engine
}
