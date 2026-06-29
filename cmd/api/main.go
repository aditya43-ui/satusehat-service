// File: /home/meninjar/goprint/service-general/main.go
package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"service/internal/auth"
	"service/internal/infrastructure/cache"
	"service/internal/infrastructure/config"
	"service/internal/infrastructure/database"
	httpServer "service/internal/infrastructure/transport/http/servers"
	"service/internal/interfaces/minio"
	satuSehatFactory "service/internal/interfaces/satusehat"

	satuSehatAuth "service/internal/satusehat/reference/auth"
	satuSehatKfa "service/internal/satusehat/reference/kfa"
	satuSehatKyc "service/internal/satusehat/reference/kyc"
	satuSehatLocation "service/internal/satusehat/reference/location"
	satuSehatOrganization "service/internal/satusehat/reference/organization"
	satuSehatPatient "service/internal/satusehat/reference/patient"
	satuSehatPractitioner "service/internal/satusehat/reference/practitioner"
	"service/internal/satusehat/usecase/allergyintolerance"
	"service/internal/satusehat/usecase/careplan"
	clinicalimpression "service/internal/satusehat/usecase/clinicalImpression"
	"service/internal/satusehat/usecase/composition"
	satuSehatCondition "service/internal/satusehat/usecase/condition"
	"service/internal/satusehat/usecase/diagnosticreport"
	satuSehatEncounter "service/internal/satusehat/usecase/encounter"
	"service/internal/satusehat/usecase/episodeofcare"
	"service/internal/satusehat/usecase/imagingstudy"
	"service/internal/satusehat/usecase/immunization"
	"service/internal/satusehat/usecase/medication"
	"service/internal/satusehat/usecase/medicationdispense"
	"service/internal/satusehat/usecase/medicationrequest"
	"service/internal/satusehat/usecase/medicationstatement"
	"service/internal/satusehat/usecase/observation"
	satuSehatProcedure "service/internal/satusehat/usecase/procedure"
	"service/internal/satusehat/usecase/questionnaireresponse"
	"service/internal/satusehat/usecase/servicerequest"
	"service/internal/satusehat/usecase/specimen"
	"service/internal/satusehat/usecase/studies"
	"service/pkg/logger"

	// roleComponent "service/internal/master/role/component"
	roleMaster "service/internal/master/role/master"
	rolePages "service/internal/master/role/pages"
	rolePermission "service/internal/master/role/permission"

	_ "service/docs/swagger" // Wajib: import swagger docs yang di-generate oleh swag CLI

	grpcServers "service/internal/infrastructure/transport/grpc/servers"

	"golang.org/x/sync/errgroup"
)

//	@title			GoPrint Service General API
//	@version		1.0.0
//	@description	REST API for Service General including Multi-DB, BPJS, and SatuSehat integrations.
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.email	support@example.com

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

// @BasePath					/api/v1
// @securityDefinitions.apikey	BearerAuth
// @in							header
// @name						Authorization
// @description				Type "Bearer" followed by a space and JWT token.
func main() {
	// 1. Load Config
	cfg := config.LoadConfig()

	// 2. Init Logger dengan konfigurasi dari config
	loggerConfig := logger.Config{
		Level:        cfg.Logger.Level,
		Format:       cfg.Logger.Format,
		Output:       "both", // Output ke console dan ke dalam folder daily logs
		ServiceName:  "service-general",
		EnableCaller: cfg.Server.Mode != "production", // Aktifkan caller info di non-production
		Environment:  cfg.Server.Mode,                 // Gunakan mode server untuk environment
	}
	logger.Init(loggerConfig)

	// Sekarang bisa menggunakan logger custom
	logger.Default().Info("Application starting",
		logger.String("service", "service-general"),
		logger.String("environment", cfg.Server.Mode),
		logger.String("log_level", cfg.Logger.Level),
	)
	// Di dalam main.go

	//  Asumsikan di config.yaml Anda ada database bernama "postgres" dan "mysql_legacy" contoh pemanggilan dan prosesing 2 tabel berbeda dalam satu repository

	// reportRepo := report.NewRepository(dbService, "postgres", "mysql_legacy")
	// reportService := report.NewService(reportRepo)

	// reportService ini tinggal di-pass ke dalam Handler untuk dijadikan endpoint REST API.

	// Add debugging to see what config was loaded
	logger.Default().Debug("Config loaded",
		logger.Any("database_count", len(cfg.Databases)),
		logger.Any("cache_enabled", cfg.Cache.Enabled),
		logger.Any("server_mode", cfg.Server.Mode),
	)

	if err := cfg.Validate(); err != nil {
		logger.Default().Fatal("Config validation failed", logger.ErrorField(err))
	}

	// --- Inisialisasi Minio Object Storage ---
	minio.Connect()
	logger.Default().Info("Minio Object Storage client initialized")

	// 3. Init Database Manager (Mendukung CQRS & Multi-DB)
	dbService := database.New(cfg)
	defer dbService.Close()

	// Menjalankan Automigrasi
	// if err := dbService.Migrate(); err != nil {
	// 	logger.Default().Warn("Database migration completed with issues", logger.ErrorField(err))
	// }

	// Get Master/Primary GORM DB Connection untuk modul lama
	// Gunakan nama koneksi "postgres" atau "default" sesuai config.yaml Anda
	gormDB, err := dbService.GetGormDB("default")
	if err != nil {
		logger.Default().Fatal("Failed to get primary gorm database", logger.ErrorField(err))
	}

	// --- [NEW] Inisialisasi Message Broker (Kafka / RabbitMQ) untuk Event-Driven Microservice ---
	// Jika bertindak sebagai gateway atau microservice independen, inisialisasi Kafka di sini
	// kafkaProducer := broker.NewKafkaProducer(cfg.Kafka.Brokers)
	// defer kafkaProducer.Close()
	// logger.Default().Info("✅ Kafka Producer initialized")

	// 4. Init Cache using factory pattern
	cacheFactory := cache.NewFactory(cfg.Cache)
	cacheManager, err := cacheFactory.CreateManager()
	if err != nil {
		logger.Default().Fatal("Failed to initialize cache manager", logger.ErrorField(err))
	}
	defer cacheManager.Close()

	if cfg.Cache.Enabled {
		logger.Default().Info("Cache initialized", logger.String("host", cfg.Cache.Redis.Host), logger.Int("port", cfg.Cache.Redis.Port))
	} else {
		logger.Default().Warn("Cache disabled, using NoOp cache")
	}

	// 5. Init Layers (CQRS Repositories & Services)
	// Modul Auth
	authCmdRepo := auth.NewCommandRepository(dbService, "default")
	authQueryRepo := auth.NewQueryRepository(dbService, "default")
	authSvc := auth.NewService(authCmdRepo, authQueryRepo, cacheManager, cfg)

	// Inisialisasi Role Pages, Permission, dan Component (Menggunakan pola CQRS baru)
	rolePagesCmdRepo := rolePages.NewCommandRepository(dbService, "default")
	rolePagesQueryRepo := rolePages.NewQueryRepository(dbService, "default")
	rolePagesService := rolePages.NewService(rolePagesCmdRepo, rolePagesQueryRepo, cacheManager)

	rolePermissionCmdRepo := rolePermission.NewCommandRepository(dbService, "default")
	rolePermissionQueryRepo := rolePermission.NewQueryRepository(dbService, "default")
	rolePermissionService := rolePermission.NewService(rolePermissionCmdRepo, rolePermissionQueryRepo, rolePagesQueryRepo, cacheManager)

	// Modul Role Master (Hasil Generator)
	roleMasterCmdRepo := roleMaster.NewCommandRepository(dbService, "default")
	roleMasterQueryRepo := roleMaster.NewQueryRepository(dbService, "default")
	roleMasterService := roleMaster.NewService(roleMasterCmdRepo, roleMasterQueryRepo, cacheManager)

	// roleComponentCmdRepo := roleComponent.NewCommandRepository(dbService, "postgres")
	// roleComponentQueryRepo := roleComponent.NewQueryRepository(dbService, "postgres")
	// roleComponentService := roleComponent.NewService(roleComponentCmdRepo, roleComponentQueryRepo)

	// --- [NEW] Inisialisasi Modul Satu Sehat ---
	satusehatFact := satuSehatFactory.NewSatuSehatFactory(cfg.SatuSehat)
	satusehatClient := satusehatFact.Client()
	satuSehatAuthSvc := satuSehatAuth.NewService(satusehatClient, cacheManager)
	satuSehatPatientRepo := satuSehatPatient.NewRepository(satusehatClient)
	satuSehatPatientSvc := satuSehatPatient.NewService(satuSehatPatientRepo)

	satuSehatPractitionerRepo := satuSehatPractitioner.NewRepository(satusehatClient)
	satuSehatPractitionerSvc := satuSehatPractitioner.NewService(satuSehatPractitionerRepo)

	satuSehatOrganizationRepo := satuSehatOrganization.NewRepository(satusehatClient)
	satuSehatOrganizationSvc := satuSehatOrganization.NewService(satuSehatOrganizationRepo)

	satuSehatLocationRepo := satuSehatLocation.NewRepository(satusehatClient)
	satuSehatLocationSvc := satuSehatLocation.NewService(satuSehatLocationRepo)

	satuSehatKfaRepo := satuSehatKfa.NewRepository(satusehatClient)
	satuSehatKfaSvc := satuSehatKfa.NewService(satuSehatKfaRepo)

	satuSehatEncounterRepo := satuSehatEncounter.NewRepository(satusehatClient, dbService)
	satuSehatEncounterSvc := satuSehatEncounter.NewService(satuSehatEncounterRepo, cfg.SatuSehat.OrgID)

	satuSehatProcedureRepo := satuSehatProcedure.NewRepository(satusehatClient)
	satuSehatProcedureSvc := satuSehatProcedure.NewService(satuSehatProcedureRepo, cfg.SatuSehat.OrgID)

	satuSehatKycRepo := satuSehatKyc.NewRepository(satusehatClient)
	satuSehatKycSvc := satuSehatKyc.NewService(satuSehatKycRepo, cfg.SatuSehat)

	// AllergyIntolerance
	satuSehatAllergyIntoleranceRepo := allergyintolerance.NewRepository(satusehatClient)
	satuSehatAllergyIntoleranceSvc := allergyintolerance.NewService(satuSehatAllergyIntoleranceRepo, cfg.SatuSehat.OrgID)

	// CarePlan
	satuSehatCarePlanRepo := careplan.NewRepository(satusehatClient)
	satuSehatCarePlanSvc := careplan.NewService(satuSehatCarePlanRepo, cfg.SatuSehat.OrgID)

	// ClinicalImpression
	satuSehatClinicalImpressionRepo := clinicalimpression.NewRepository(satusehatClient)
	satuSehatClinicalImpressionSvc := clinicalimpression.NewService(satuSehatClinicalImpressionRepo, cfg.SatuSehat.OrgID)

	// Composition
	satuSehatCompositionRepo := composition.NewRepository(satusehatClient)
	satuSehatCompositionSvc := composition.NewService(satuSehatCompositionRepo, cfg.SatuSehat.OrgID)

	satuSehatConditionRepo := satuSehatCondition.NewRepository(satusehatClient)
	satuSehatConditionSvc := satuSehatCondition.NewService(satuSehatConditionRepo, cfg.SatuSehat.OrgID)

	// DiagnosticReport
	satuSehatDiagnosticReportRepo := diagnosticreport.NewRepository(satusehatClient)
	satuSehatDiagnosticReportSvc := diagnosticreport.NewService(satuSehatDiagnosticReportRepo, cfg.SatuSehat.OrgID)

	// EpisodeOfCare
	satuSehatEpisodeOfCareRepo := episodeofcare.NewRepository(satusehatClient)
	satuSehatEpisodeOfCareSvc := episodeofcare.NewService(satuSehatEpisodeOfCareRepo, cfg.SatuSehat.OrgID)

	// ImagingStudy
	satuSehatImagingStudyRepo := imagingstudy.NewRepository(satusehatClient)
	satuSehatImagingStudySvc := imagingstudy.NewService(satuSehatImagingStudyRepo, cfg.SatuSehat.OrgID)

	// Immunization
	satuSehatImmunizationRepo := immunization.NewRepository(satusehatClient)
	satuSehatImmunizationSvc := immunization.NewService(satuSehatImmunizationRepo, cfg.SatuSehat.OrgID)

	// Medication
	satuSehatMedicationRepo := medication.NewRepository(satusehatClient)
	satuSehatMedicationSvc := medication.NewService(satuSehatMedicationRepo, cfg.SatuSehat.OrgID)

	// MedicationDispense
	satuSehatMedicationDispenseRepo := medicationdispense.NewRepository(satusehatClient)
	satuSehatMedicationDispenseSvc := medicationdispense.NewService(satuSehatMedicationDispenseRepo, cfg.SatuSehat.OrgID)

	// MedicationRequest
	satuSehatMedicationRequestRepo := medicationrequest.NewRepository(satusehatClient)
	satuSehatMedicationRequestSvc := medicationrequest.NewService(satuSehatMedicationRequestRepo, cfg.SatuSehat.OrgID)

	// MedicationStatement
	satuSehatMedicationStatementRepo := medicationstatement.NewRepository(satusehatClient)
	satuSehatMedicationStatementSvc := medicationstatement.NewService(satuSehatMedicationStatementRepo, cfg.SatuSehat.OrgID)

	// Observation
	satuSehatObservationRepo := observation.NewRepository(satusehatClient)
	satuSehatObservationSvc := observation.NewService(satuSehatObservationRepo, cfg.SatuSehat.OrgID)

	// QuestionnaireResponse
	satuSehatQuestionnaireResponseRepo := questionnaireresponse.NewRepository(satusehatClient)
	satuSehatQuestionnaireResponseSvc := questionnaireresponse.NewService(satuSehatQuestionnaireResponseRepo, cfg.SatuSehat.OrgID)

	// ServiceRequest
	satuSehatServiceRequestRepo := servicerequest.NewRepository(satusehatClient)
	satuSehatServiceRequestSvc := servicerequest.NewService(satuSehatServiceRequestRepo, cfg.SatuSehat.OrgID)

	// Specimen
	satuSehatSpecimenRepo := specimen.NewRepository(satusehatClient)
	satuSehatSpecimenSvc := specimen.NewService(satuSehatSpecimenRepo, cfg.SatuSehat.OrgID)

	// DICOM Studies
	satuSehatStudiesRepo := studies.NewRepository(satusehatClient)
	satuSehatStudiesSvc := studies.NewService(satuSehatStudiesRepo)
	// 6. Server Orchestration (Dual Protocol & Background Workers)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// Gunakan errgroup untuk mengelola lifecycle server
	g, ctx := errgroup.WithContext(ctx)

	// // --- [NEW] Start Background Workers ---
	// if cfg.SatuSehat.Enabled {
	// 	imagingStudyWorkerCfg := imagingStudyWorker.Config{
	// 		DBManager:       dbService,
	// 		InternalBaseURL: fmt.Sprintf("http://localhost:%d%s", cfg.Server.REST.Port, cfg.Swagger.BasePath), // otomatis: http://localhost:8080/api/v1
	// 		InternalToken:   "",                                                                               // Isi token di sini jika endpoint internal Usecase Anda membutuhkan Bearer Token
	// 		OrganizationID:  cfg.SatuSehat.OrgID,
	// 	}
	// 	isWorker := imagingStudyWorker.NewWorker(imagingStudyWorkerCfg)

	// 	g.Go(func() error {
	// 		isWorker.Run(ctx)
	// 		return nil
	// 	})
	// }

	serverCount := 0

	// --- [A] Start REST Server ---
	if cfg.Server.REST.Enabled {
		serverCount++

		// Menggunakan ServiceRegistry agar HTTP Server lebih rapi dan scalable
		registry := &httpServer.ServiceRegistry{
			Config:       cfg,
			DBManager:    dbService,
			PrimaryDB:    gormDB,
			CacheManager: cacheManager,
			AuthService:  authSvc,
			Master: &httpServer.MasterServices{
				RolePages:      rolePagesService,
				RolePermission: rolePermissionService,
				RoleMaster:     roleMasterService,
			},
			SatuSehat: &httpServer.SatuSehatServices{
				Auth:                  satuSehatAuthSvc,
				Patient:               satuSehatPatientSvc,
				Practitioner:          satuSehatPractitionerSvc,
				Organization:          satuSehatOrganizationSvc,
				Location:              satuSehatLocationSvc,
				KFA:                   satuSehatKfaSvc,
				KYC:                   satuSehatKycSvc,
				Encounter:             satuSehatEncounterSvc,
				Procedure:             satuSehatProcedureSvc,
				Condition:             satuSehatConditionSvc,
				AllergyIntolerance:    satuSehatAllergyIntoleranceSvc,
				CarePlan:              satuSehatCarePlanSvc,
				ClinicalImpression:    satuSehatClinicalImpressionSvc,
				Composition:           satuSehatCompositionSvc,
				DiagnosticReport:      satuSehatDiagnosticReportSvc,
				EpisodeOfCare:         satuSehatEpisodeOfCareSvc,
				ImagingStudy:          satuSehatImagingStudySvc,
				Immunization:          satuSehatImmunizationSvc,
				Medication:            satuSehatMedicationSvc,
				MedicationDispense:    satuSehatMedicationDispenseSvc,
				MedicationRequest:     satuSehatMedicationRequestSvc,
				MedicationStatement:   satuSehatMedicationStatementSvc,
				Observation:           satuSehatObservationSvc,
				QuestionnaireResponse: satuSehatQuestionnaireResponseSvc,
				ServiceRequest:        satuSehatServiceRequestSvc,
				Specimen:              satuSehatSpecimenSvc,
				DicomStudies:          satuSehatStudiesSvc,
			},
		}

		restSrv := httpServer.NewHTTPServer(&cfg.Server.REST, registry)

		g.Go(func() error {
			logger.Default().Info("REST API running", logger.Int("port", cfg.Server.REST.Port))
			if err := restSrv.Start(&cfg.Server); err != nil && err != http.ErrServerClosed {
				logger.Default().Error("REST Server error", logger.ErrorField(err))
				return err
			}
			return nil
		})
	}

	// --- [B] Start gRPC Server ---
	if cfg.Server.GRPC.Enabled {
		serverCount++

		// 1. Buat gRPC handlers
		// permissionHandler := grpcHandlers.NewPermissionHandler(rolePermissionService)
		// 2. Buat dan isi gRPC service registry
		grpcRegistry := &grpcServers.ServiceRegistry{
			// PermissionHandler: permissionHandler,
		}

		// 3. Buat gRPC server dengan registry yang sudah diisi
		grpcSrv := grpcServers.NewGRPCServer(
			&cfg.Server.GRPC,
			grpcRegistry,
		)

		g.Go(func() error {
			logger.Default().Info("gRPC Server running", logger.Int("port", cfg.Server.GRPC.Port))
			if err := grpcSrv.Start(); err != nil {
				logger.Default().Error("gRPC Server error", logger.ErrorField(err))
				return err
			}
			return nil
		})

		g.Go(func() error {
			<-ctx.Done()
			logger.Default().Info("Gracefully stopping gRPC Server...")
			grpcSrv.Stop()
			return nil
		})
	}

	if serverCount == 0 {
		logger.Default().Fatal("No server enabled (REST and gRPC are both disabled in config)")
	}

	// --- [C] Graceful Shutdown Listener ---
	g.Go(func() error {
		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		select {
		case <-quit:
			logger.Default().Warn("Signal received, shutting down servers...")
			cancel()
		case <-ctx.Done():
		}
		return nil
	})

	logger.Default().Info("Application started successfully")

	if err := g.Wait(); err != nil {
		logger.Default().Error("Server shutdown with error", logger.ErrorField(err))
	} else {
		logger.Default().Info("Server shutdown successful")
	}
}
