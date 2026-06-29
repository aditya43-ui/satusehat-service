package routes

import (
	"net/http"
	"time"

	"service/internal/infrastructure/cache"
	"service/internal/infrastructure/config"
	"service/internal/infrastructure/transport/http/middleware"

	authHttp "service/internal/infrastructure/transport/http/handlers/main/auth"
	healthHttp "service/internal/infrastructure/transport/http/handlers/main/health"
	roleHttp "service/internal/infrastructure/transport/http/handlers/main/master/roles"
	satuSehatRefHttp "service/internal/infrastructure/transport/http/handlers/satusehat/reference"
	satuSehatUsecaseHttp "service/internal/infrastructure/transport/http/handlers/satusehat/usecase"

	"github.com/gin-gonic/gin"
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
)

// ModuleHandlers menampung seluruh HTTP handler opsional dari berbagai modul.
// Jika sebuah handler bernilai nil, maka rutenya akan otomatis diabaikan (disabled).
type ModuleHandlers struct {
	Auth                    *authHttp.AuthHandler
	RolePages               *roleHttp.RolPagesHandler
	RolePermission          *roleHttp.RolPermissionHandler
	RoleMaster              *roleHttp.RoleMasterHandler
	SatuSehatAuth           *satuSehatRefHttp.AuthHandler
	SatuSehatPatient        *satuSehatRefHttp.PatientHandler
	SatuSehatPractitioner   *satuSehatRefHttp.PractitionerHandler
	SatuSehatOrganization   *satuSehatRefHttp.OrganizationHandler
	SatuSehatLocation       *satuSehatRefHttp.LocationHandler
	SatuSehatKFA            *satuSehatRefHttp.KFAHandler
	SatuSehatKYC            *satuSehatRefHttp.KYCHandler
	SSAllergyIntolerance    *satuSehatUsecaseHttp.AllergyIntoleranceHandler
	SSCarePlan              *satuSehatUsecaseHttp.CarePlanHandler
	SSClinicalImpression    *satuSehatUsecaseHttp.ClinicalImpressionHandler
	SSComposition           *satuSehatUsecaseHttp.CompositionHandler
	SSCondition             *satuSehatUsecaseHttp.ConditionHandler
	SSDiagnosticReport      *satuSehatUsecaseHttp.DiagnosticReportHandler
	SSEncounter             *satuSehatUsecaseHttp.EncounterHandler
	SSEpisodeOfCare         *satuSehatUsecaseHttp.EpisodeOfCareHandler
	SSImagingStudy          *satuSehatUsecaseHttp.ImagingStudyHandler
	SSImmunization          *satuSehatUsecaseHttp.ImmunizationHandler
	SSMedication            *satuSehatUsecaseHttp.MedicationHandler
	SSMedicationDispense    *satuSehatUsecaseHttp.MedicationDispenseHandler
	SSMedicationRequest     *satuSehatUsecaseHttp.MedicationRequestHandler
	SSMedicationStatement   *satuSehatUsecaseHttp.MedicationStatementHandler
	SSObservation           *satuSehatUsecaseHttp.ObservationHandler
	SSProcedure             *satuSehatUsecaseHttp.ProcedureHandler
	SSQuestionnaireResponse *satuSehatUsecaseHttp.QuestionnaireResponseHandler
	SSServiceRequest        *satuSehatUsecaseHttp.ServiceRequestHandler
	SSSpecimen              *satuSehatUsecaseHttp.SpecimenHandler
	SSDicomStudies          *satuSehatUsecaseHttp.DicomStudiesHandler
}

// SetupRoutes mendaftarkan seluruh endpoint API secara dinamis berdasarkan
// modul-modul yang aktif (tidak nil) pada aplikasi.
func SetupRoutes(
	engine *gin.Engine,
	cfg *config.Config,
	cacheManager *cache.Manager,
	healthHandler *healthHttp.HealthHandler,
	h *ModuleHandlers,
) {
	// Handle 405 Method Not Allowed for better client feedback
	engine.HandleMethodNotAllowed = true

	// Health check endpoints - using consistent http.Status constants
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "healthy",
			"timestamp": time.Now().UTC(),
			"service":   "service-general",
			"version":   "1.0.0",
		})
	})

	// Comprehensive health check with dependencies
	engine.GET("/health/complete", healthHandler.HealthCheckComplete)

	// Database health check
	engine.GET("/health/database", healthHandler.HealthCheckDatabase)

	// Redis/cache health check
	engine.GET("/health/cache", healthHandler.HealthCheckCache)

	// External services health check
	engine.GET("/health/external", healthHandler.HealthCheckExternal)

	// Minio service
	engine.GET("/health/minio", healthHandler.HealthCheckMinio)
	engine.POST("/health/minio/upload", healthHandler.TestUploadMinio)

	// All databases health check
	// engine.GET("/health/databases", healthHandler.HealthCheckAllDatabases)

	// Readiness check
	engine.GET("/ready", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "ready",
			"timestamp": time.Now().UTC(),
		})
	})

	// Liveness check
	engine.GET("/live", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":    "alive",
			"timestamp": time.Now().UTC(),
		})
	})

	// Redirect otomatis dari /swagger ke /swagger/index.html
	engine.GET("/swagger", func(c *gin.Context) {
		c.Redirect(http.StatusFound, "/swagger/index.html")
	})

	// Endpoint untuk Swagger UI
	engine.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// API v1 routes
	v1 := engine.Group("/api/v1")
	{
		// Register auth routes (Public)
		if h.Auth != nil {
			h.Auth.RegisterRoutes(v1)
		}

		// Private routes (Membutuhkan Autentikasi)
		protected := v1.Group("")
		protected.Use(middleware.AuthMiddleware(cfg, cacheManager)) // Di-comment sementara untuk proses development/testing
		{
			if h.Auth != nil {
				h.Auth.RegisterProtectedRoutes(protected)
			}

			// --- Routes Modul Master ---
			if h.RolePages != nil {
				h.RolePages.RegisterRoutes(protected)
			}
			if h.RolePermission != nil {
				h.RolePermission.RegisterRoutes(protected)
			}
			if h.RoleMaster != nil {
				h.RoleMaster.RegisterRoutes(protected)
			}

			// --- Routes Modul Satu Sehat ---
			// Buat grup khusus untuk membatasi traffic hit API ke Kemenkes
			satuSehatGroup := protected.Group("")
			satuSehatGroup.Use(middleware.MemoryRateLimitMiddleware(5.0, 10)) // Max 5 RPS, Burst capacity 10
			{
				if h.SatuSehatAuth != nil {
					h.SatuSehatAuth.RegisterRoutes(satuSehatGroup)
				}
				if h.SatuSehatPatient != nil {
					h.SatuSehatPatient.RegisterRoutes(satuSehatGroup)
				}
				if h.SatuSehatPractitioner != nil {
					h.SatuSehatPractitioner.RegisterRoutes(satuSehatGroup)
				}
				if h.SatuSehatOrganization != nil {
					h.SatuSehatOrganization.RegisterRoutes(satuSehatGroup)
				}
				if h.SatuSehatLocation != nil {
					h.SatuSehatLocation.RegisterRoutes(satuSehatGroup)
				}
				if h.SatuSehatKFA != nil {
					h.SatuSehatKFA.RegisterRoutes(satuSehatGroup)
				}
				if h.SatuSehatKYC != nil {
					h.SatuSehatKYC.RegisterRoutes(satuSehatGroup)
				}
				if h.SSAllergyIntolerance != nil {
					h.SSAllergyIntolerance.RegisterRoutes(satuSehatGroup)
				}
				if h.SSCarePlan != nil {
					h.SSCarePlan.RegisterRoutes(satuSehatGroup)
				}
				if h.SSClinicalImpression != nil {
					h.SSClinicalImpression.RegisterRoutes(satuSehatGroup)
				}
				if h.SSComposition != nil {
					h.SSComposition.RegisterRoutes(satuSehatGroup)
				}
				if h.SSCondition != nil {
					h.SSCondition.RegisterRoutes(satuSehatGroup)
				}
				if h.SSDiagnosticReport != nil {
					h.SSDiagnosticReport.RegisterRoutes(satuSehatGroup)
				}
				if h.SSEncounter != nil {
					h.SSEncounter.RegisterRoutes(satuSehatGroup)
				}
				if h.SSEpisodeOfCare != nil {
					h.SSEpisodeOfCare.RegisterRoutes(satuSehatGroup)
				}
				if h.SSImagingStudy != nil {
					h.SSImagingStudy.RegisterRoutes(satuSehatGroup)
				}
				if h.SSImmunization != nil {
					h.SSImmunization.RegisterRoutes(satuSehatGroup)
				}
				if h.SSMedication != nil {
					h.SSMedication.RegisterRoutes(satuSehatGroup)
				}
				if h.SSMedicationDispense != nil {
					h.SSMedicationDispense.RegisterRoutes(satuSehatGroup)
				}
				if h.SSMedicationRequest != nil {
					h.SSMedicationRequest.RegisterRoutes(satuSehatGroup)
				}
				if h.SSMedicationStatement != nil {
					h.SSMedicationStatement.RegisterRoutes(satuSehatGroup)
				}
				if h.SSObservation != nil {
					h.SSObservation.RegisterRoutes(satuSehatGroup)
				}
				if h.SSProcedure != nil {
					h.SSProcedure.RegisterRoutes(satuSehatGroup)
				}
				if h.SSQuestionnaireResponse != nil {
					h.SSQuestionnaireResponse.RegisterRoutes(satuSehatGroup)
				}
				if h.SSServiceRequest != nil {
					h.SSServiceRequest.RegisterRoutes(satuSehatGroup)
				}
				if h.SSSpecimen != nil {
					h.SSSpecimen.RegisterRoutes(satuSehatGroup)
				}
				if h.SSDicomStudies != nil {
					h.SSDicomStudies.RegisterRoutes(satuSehatGroup)
				}
			}
		}
	}
}
