-- ==========================================
-- BAGIAN 1: CREATE TABLE (Tanpa Foreign Key)
-- ==========================================

CREATE TABLE public."AntibioticSrcCategory" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Name" varchar(50) NULL,
	CONSTRAINT "AntibioticSrcCategory_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_AntibioticSrcCategory_Code" UNIQUE ("Code")
);

CREATE TABLE public."AuthPartner" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(50) NULL,
	"Name" varchar(100) NULL,
	"SecretKey" varchar(255) NULL,
	CONSTRAINT "AuthPartner_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_AuthPartner_Code" UNIQUE ("Code"),
	CONSTRAINT "uni_AuthPartner_Name" UNIQUE ("Name")
);

CREATE TABLE public."AntibioticSrc" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Name" varchar(50) NULL,
	"AntibioticSrcCategory_Code" varchar(20) NULL,
	CONSTRAINT "AntibioticSrc_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_AntibioticSrc_Code" UNIQUE ("Code")
);

CREATE TABLE public."Adime" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Employee_Id" int8 NULL,
	"Time" timestamptz NULL,
	"Value" text NULL,
	CONSTRAINT "Adime_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."AdmEmployeeHist" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Employee_Id" int8 NULL,
	"StartedAt" timestamptz NULL,
	"FinishedAt" timestamptz NULL,
	CONSTRAINT "AdmEmployeeHist_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."AmbulanceTransportReq" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Patient_Id" int8 NULL,
	"Diagnoses" varchar(1024) NULL,
	"RequestDate" timestamptz NULL,
	"UsageDate" timestamptz NULL,
	"Address" varchar(100) NULL,
	"RtRw" varchar(10) NULL,
	"Province_Code" varchar(2) NULL,
	"Regency_Code" varchar(4) NULL,
	"District_Code" varchar(6) NULL,
	"Village_Code" varchar(10) NULL,
	"Facility_Code" varchar(10) NULL,
	"Needs_Code" varchar(10) NULL,
	"Contact_Name" varchar(100) NULL,
	"Contact_Relationship_Code" varchar(10) NULL,
	"Contact_PhoneNumber" varchar(20) NULL,
	CONSTRAINT "AmbulanceTransportReq_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Ambulatory" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Class_Code" varchar(10) NULL,
	CONSTRAINT "Ambulatory_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."AntibioticInUse" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"McuOrder_Id" int8 NULL,
	"AntibioticSrc_Id" int8 NULL,
	CONSTRAINT "AntibioticInUse_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."ApMcuOrder" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Substances" text NULL,
	"Fictations" text NULL,
	"Localization" text NULL,
	"ClinicalDiagnoses" text NULL,
	"Stadium" text NULL,
	"ClinicalNotes" text NULL,
	"PastHistory" text NULL,
	"CurrentHistory" text NULL,
	"PrevApMcu" text NULL,
	"PrevApMcuNotes" text NULL,
	"SupportingExams" text NULL,
	"Encounter_Id" int8 NULL,
	"Number" int8 NULL,
	"Doctor_Code" varchar(20) NULL,
	"Status_Code" varchar(10) NOT NULL,
	CONSTRAINT "ApMcuOrder_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_ApMcuOrder_Doctor_Code" UNIQUE ("Doctor_Code")
);

CREATE TABLE public."Appointment" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"PracticeSchedule_Id" int8 NULL,
	"Patient_Id" int8 NULL,
	"Person_ResidentIdentityNumber" varchar(16) NULL,
	"Person_Name" varchar(100) NULL,
	"Person_PhoneNumber" varchar(30) NULL,
	"PaymentMethod_Code" varchar(10) NULL,
	"RefNumber" varchar(20) NULL,
	CONSTRAINT "Appointment_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."ChamberClass" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Name" varchar(50) NOT NULL,
	"Code" varchar(10) NULL,
	CONSTRAINT "ChamberClass_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_ChamberClass_Code" UNIQUE ("Code")
);

CREATE TABLE public."Counter" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(10) NULL,
	"Name" varchar(30) NULL,
	"Number" int2 NULL,
	"Parent_Id" int4 NULL,
	"Type_Code" text NULL,
	"Queue_Code" varchar(5) NULL,
	CONSTRAINT "Counter_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Counter_Code" UNIQUE ("Code")
);

CREATE TABLE public."DevicePackage" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NOT NULL,
	"Name" varchar(50) NULL,
	CONSTRAINT "DevicePackage_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_DevicePackage_Code" UNIQUE ("Code")
);

CREATE TABLE public."DiagnoseSrc" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(10) NULL,
	"Name" varchar(2048) NULL,
	"IndName" varchar(2048) NULL,
	"Duration" int8 NULL,
	"DurationUnit_Code" text NULL,
	CONSTRAINT "DiagnoseSrc_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_DiagnoseSrc_Code" UNIQUE ("Code")
);

CREATE TABLE public."Division" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(10) NULL,
	"Name" varchar(50) NULL,
	"Parent_Code" varchar(10) NULL,
	CONSTRAINT "Division_pkey" PRIMARY KEY ("Id")
);
CREATE UNIQUE INDEX "idx_Division_Code" ON public."Division" USING btree ("Code");

CREATE TABLE public."Bed" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Infra_Id" int8 NOT NULL,
	"Number" int8 NOT NULL,
	CONSTRAINT "Bed_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Billing" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NOT NULL,
	"Item_Code" varchar(50) NOT NULL,
	"Price" numeric NOT NULL,
	"Number" varchar(20) NULL,
	"Code" varchar(20) NULL,
	CONSTRAINT "Billing_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Chamber" (
	"Class_Code" varchar(50) NULL,
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Infra_Id" int8 NOT NULL,
	CONSTRAINT "Chamber_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Chemo" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Status_Code" text NULL,
	"VerifiedAt" timestamptz NULL,
	"VerifiedBy_User_Id" int8 NULL,
	"Bed" varchar(1024) NULL,
	"Needs" varchar(2048) NULL,
	"Specialist_Code" varchar(20) NULL,
	"Doctor_Code" varchar(20) NULL,
	"NextChemoDate" timestamptz NULL,
	"Class_Code" text NULL,
	CONSTRAINT "Chemo_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."ChemoPlan" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Parent_Id" int8 NULL,
	"Protocol_Id" int8 NULL,
	"SeriesNumber" int4 NULL,
	"CycleNumber" int8 NULL,
	"PlanDate" timestamptz NULL,
	"RealizationDate" timestamptz NULL,
	"Notes" text NULL,
	"Status" text NULL,
	"Encounter_Id" int8 NULL,
	"Reasons" text NULL,
	CONSTRAINT "ChemoPlan_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."ChemoProtocol" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Chemo_Id" int8 NULL,
	"Patient_Weight" numeric NULL,
	"Patient_Height" numeric NULL,
	"Diagnoses" text NULL,
	"Duration" int8 NULL,
	"DurationUnit_Code" text NULL,
	"StartDate" timestamptz NULL,
	"EndDate" timestamptz NULL,
	"Interval" int8 NULL,
	"Cycle" int8 NULL,
	"Series" int4 NULL,
	"Status_Code" text NULL,
	"Patient_Id" int8 NULL,
	"VerifiedAt" timestamptz NULL,
	"VerifiedBy_User_Id" int8 NULL,
	CONSTRAINT "ChemoProtocol_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Consultation" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Solution" varchar(10240) NULL,
	"RepliedAt" timestamptz NULL,
	"Date" timestamptz NULL,
	"Problem" varchar(10240) NULL,
	"Doctor_Code" varchar(20) NULL,
	"Unit_Code" varchar(20) NULL,
	"Status_Code" varchar(30) NULL,
	CONSTRAINT "Consultation_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."ControlLetter" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Date" timestamptz NULL,
	"Specialist_Code" varchar(20) NULL,
	"Subspecialist_Code" varchar(20) NULL,
	"Doctor_Code" varchar(20) NULL,
	CONSTRAINT "ControlLetter_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."CpMcuOrder" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Number" int8 NULL,
	"Doctor_Code" varchar(20) NULL,
	"UrgencyLevel_Code" varchar(15) NOT NULL,
	"OtherNotes" text NULL,
	"ExamScheduleDate" timestamptz NULL,
	"Resume" text NULL,
	"Status_Code" varchar(10) NOT NULL,
	"Lab_Type" varchar(10) NULL,
	"Localization" text NULL,
	"Stadium" varchar(50) NULL,
	is_inpatient bool DEFAULT false NULL,
	CONSTRAINT "CpMcuOrder_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."CpMcuOrderItem" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"CpMcuOrder_Id" int8 NULL,
	"McuSrc_Code" varchar(20) NULL,
	"Note" varchar(1024) NULL,
	"Result" text NULL,
	"Status_Code" text NULL,
	CONSTRAINT "CpMcuOrderItem_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."DeathCause" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NOT NULL,
	"Value" text NULL,
	CONSTRAINT "DeathCause_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Device" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(10) NULL,
	"Name" varchar(50) NOT NULL,
	"Uom_Code" varchar(10) NULL,
	"Infra_Id" int4 NULL,
	"Item_Id" int8 NULL,
	"Infra_Code" varchar(10) NULL,
	"Item_Code" varchar(50) NULL,
	CONSTRAINT "Device_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Device_Code" UNIQUE ("Code")
);

CREATE TABLE public."DeviceOrder" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Status_Code" text NULL,
	"Doctor_Code" varchar(20) NULL,
	is_inpatient bool DEFAULT false NULL,
	"Order_Date" text NULL,
	CONSTRAINT "DeviceOrder_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."DeviceOrderItem" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"DeviceOrder_Id" int8 NULL,
	"Quantity" int2 NULL,
	"Device_Code" varchar(10) NULL,
	CONSTRAINT "DeviceOrderItem_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."DevicePackageItem" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"DevicePackage_Code" varchar(20) NOT NULL,
	"Device_Code" varchar(20) NOT NULL,
	CONSTRAINT "DevicePackageItem_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."District" (
	"Id" bigserial NOT NULL,
	"Regency_Code" varchar(4) NULL,
	"Code" varchar(6) NULL,
	"Name" varchar(50) NULL,
	CONSTRAINT "District_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_District_Code" UNIQUE ("Code")
);

CREATE TABLE public."DivisionPosition" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Managerial_Code" varchar(10) NULL,
	"Name" varchar(50) NULL,
	"Employee_Id" int8 NULL,
	"HeadStatus" bool NULL,
	"Division_Code" varchar(10) NULL,
	CONSTRAINT "DivisionPosition_pkey" PRIMARY KEY ("Id")
);
CREATE UNIQUE INDEX division_position_idx ON public."DivisionPosition" USING btree ("Division_Code", "Managerial_Code");

CREATE TABLE public."Doctor" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Employee_Id" int8 NULL,
	"IHS_Number" varchar(20) NULL,
	"SIP_Number" varchar(20) NULL,
	"Code" varchar(20) NULL,
	"SIP_ExpiredDate" timestamptz NULL,
	"Unit_Code" varchar(60) NULL,
	CONSTRAINT "Doctor_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Doctor_Code" UNIQUE ("Code"),
	CONSTRAINT "uni_Doctor_IHS_Number" UNIQUE ("IHS_Number"),
	CONSTRAINT "uni_Doctor_SIP_Number" UNIQUE ("SIP_Number")
);

CREATE TABLE public."DoctorFee" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Doctor_Id" int8 NULL,
	"FeeType_Code" varchar(11) NULL,
	"Price" numeric NULL,
	"Item_Id" int8 NULL,
	CONSTRAINT "DoctorFee_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."DoctorUnit" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Doctor_Id" int8 NULL,
	"Unit_Id" int4 NULL,
	CONSTRAINT "DoctorUnit_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Ethnic" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Name" varchar(50) NULL,
	CONSTRAINT "Ethnic_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Ethnic_Code" UNIQUE ("Code")
);

CREATE TABLE public."Goal" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Goal_Name" text NULL,
	CONSTRAINT "Goal_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."HealthCareFacility" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Divre_Code" text NULL,
	"Facility_Code" text NULL,
	"Facility_Name" text NULL,
	"Regency_Id" int8 NULL,
	"Facility_Type" text NULL,
	"Address" text NULL,
	"Phone" text NULL,
	"Status" bool NULL,
	"Reference_Code" text NULL,
	"Governance_code" text NULL,
	"Bpjs_code" text NULL,
	CONSTRAINT "HealthCareFacility_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Icd10" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" text NULL,
	"Name" text NULL,
	"IndName" text NULL,
	"Version" text NULL,
	"Cause" text NULL,
	CONSTRAINT "Icd10_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."InpatientRequest" (
	"Id" bigserial NOT NULL,
	"Encounter_Id" int8 NULL,
	"Infra_Id" int8 NULL,
	"Status_Code" text NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	CONSTRAINT "InpatientRequest_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Installation" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Name" varchar(50) NULL,
	"EncounterClass_Code" varchar(10) NULL,
	CONSTRAINT "Installation_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Installation_Code" UNIQUE ("Code")
);

CREATE TABLE public."EduAssessment" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NOT NULL,
	"GeneralEdus" text NULL,
	"SpecialEdus" text NULL,
	"Assessments" text NULL,
	"Plan" text NULL,
	"FileUrl" varchar(1024) NULL,
	CONSTRAINT "EduAssessment_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."EduAssessmentImpl" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NOT NULL,
	"EduAssessment_Id" int8 NOT NULL,
	"EduNeeds" text NULL,
	"EduMaterials" text NULL,
	"Date" timestamptz NULL,
	"VerifResults" text NULL,
	"Employee_Id" int8 NOT NULL,
	CONSTRAINT "EduAssessmentImpl_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Emergency" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Class_Code" varchar(10) NULL,
	CONSTRAINT "Emergency_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Employee" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"User_Id" int8 NULL,
	"Person_Id" int8 NULL,
	"Number" varchar(20) NULL,
	"Status_Code" varchar(10) NOT NULL,
	"Position_Code" varchar(20) NULL,
	"Contract_ExpiredDate" timestamptz NULL,
	CONSTRAINT "Employee_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Encounter" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Patient_Id" int8 NULL,
	"RegisteredAt" timestamptz NULL,
	"Class_Code" varchar(10) NOT NULL,
	"VisitDate" timestamptz NULL,
	"RefSource_Name" varchar(100) NULL,
	"Appointment_Id" int8 NULL,
	"EarlyEducation" text NULL,
	"MedicalDischargeEducation" text NULL,
	"AdmDischargeEducation" text NULL,
	"DischargeReason" text NULL,
	"Discharge_Method_Code" varchar(16) NULL,
	"Status_Code" varchar(10) NULL,
	"PaymentMethod_Code" varchar(10) NULL,
	"Member_Number" varchar(20) NULL,
	"Ref_Number" varchar(20) NULL,
	"Trx_Number" varchar(20) NULL,
	"Adm_Employee_Id" int8 NULL,
	"Discharge_Date" timestamptz NULL,
	"StartedAt" timestamptz NULL,
	"FinishedAt" timestamptz NULL,
	"RefType_Code" text NULL,
	"NewStatus" bool NULL,
	"Specialist_Code" varchar(20) NULL,
	"Subspecialist_Code" varchar(20) NULL,
	"Appointment_Doctor_Code" varchar(20) NULL,
	"Responsible_Doctor_Code" varchar(20) NULL,
	"InsuranceCompany_Code" varchar(20) NULL,
	"Responsible_Nurse_Code" varchar(20) NULL,
	"Unit_Code" varchar(60) NULL,
	"Subsystem" text NULL,
	"ParticipantGroup" text NULL,
	"VclaimSepType_Code" text NULL,
	"Barcode" text NULL,
	"CareClass" text NULL,
	"ChiefComplaint" text NULL,
	"RefSourceType_Code" text NULL,
	"Diagnosis" text NULL,
	"AccidentStatus_Code" text NULL,
	CONSTRAINT "Encounter_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Encounter_Ref_Number" UNIQUE ("Ref_Number"),
	CONSTRAINT "uni_Encounter_Trx_Number" UNIQUE ("Trx_Number")
);

CREATE TABLE public."EncounterDocument" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Type_Code" text NULL,
	"Name" text NULL,
	"FilePath" text NULL,
	"FileName" text NULL,
	"Upload_Employee_Id" int8 NULL,
	CONSTRAINT "EncounterDocument_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."GeneralConsent" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NOT NULL,
	"Value" text NULL,
	"FileUrl" varchar(1024) NULL,
	CONSTRAINT "GeneralConsent_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Infra" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NOT NULL,
	"Name" varchar(50) NULL,
	"InfraGroup_Code" varchar(20) NULL,
	"Parent_Code" varchar(20) NULL,
	"Item_Code" varchar(50) NULL,
	"Level" int2 NULL,
	CONSTRAINT "Infra_pkey" PRIMARY KEY ("Id")
);
CREATE UNIQUE INDEX "idx_Infra_Code" ON public."Infra" USING btree ("Code");

CREATE TABLE public."Inpatient" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Class_Code" varchar(10) NULL,
	"Infra_Code" varchar(10) NULL,
	"Is_Data_Complete" bool NULL,
	"Discharge_Plan" timestamptz NULL,
	CONSTRAINT "Inpatient_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."InstallationPosition" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(10) NOT NULL,
	"Name" varchar(30) NOT NULL,
	"HeadStatus" bool NULL,
	"Employee_Id" int8 NULL,
	"Installation_Code" varchar(10) NULL,
	CONSTRAINT "InstallationPosition_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_InstallationPosition_Code" UNIQUE ("Code")
);

CREATE TABLE public."InsuranceCompany" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Name" varchar(50) NULL,
	"Regency_Code" varchar(4) NULL,
	"Address" varchar(100) NULL,
	"PhoneNumber" varchar(20) NULL,
	CONSTRAINT "InsuranceCompany_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_InsuranceCompany_Code" UNIQUE ("Code")
);

CREATE TABLE public."Intern" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Person_Id" int8 NULL,
	"Position_Code" varchar(20) NULL,
	"User_Id" int8 NULL,
	CONSTRAINT "Intern_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."InternalReference" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Status_Code" text NULL,
	"Doctor_Code" varchar(20) NULL,
	"SrcDoctor_Code" varchar(20) NULL,
	"SrcNurse_Code" varchar(20) NULL,
	"Nurse_Code" varchar(20) NULL,
	"Unit_Code" varchar(60) NULL,
	CONSTRAINT "InternalReference_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Item" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(50) NULL,
	"Name" varchar(100) NULL,
	"ItemGroup_Code" varchar(15) NULL,
	"Uom_Code" varchar(10) NULL,
	"Stock" int8 NULL,
	"Infra_Code" varchar(10) NULL,
	"BuyingPrice" numeric NULL,
	"SellingPrice" numeric NULL,
	CONSTRAINT "Item_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Item_Code" UNIQUE ("Code")
);

CREATE TABLE public."ItemPrice" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Item_Id" int8 NULL,
	"Price" numeric NULL,
	"InsuranceCompany_Code" varchar(20) NULL,
	CONSTRAINT "ItemPrice_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Language" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(10) NULL,
	"Name" varchar(50) NULL,
	CONSTRAINT "Language_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Language_Code" UNIQUE ("Code")
);

CREATE TABLE public."MaterialPackage" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NOT NULL,
	"Name" varchar(50) NULL,
	CONSTRAINT "MaterialPackage_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_MaterialPackage_Code" UNIQUE ("Code")
);

CREATE TABLE public."McuSrcCategory" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Name" varchar(50) NULL,
	"Scope_Code" varchar(10) NULL,
	CONSTRAINT "McuSrcCategory_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_McuSrcCategory_Code" UNIQUE ("Code")
);
CREATE INDEX "idx_McuSrcCategory_Scope_Code" ON public."McuSrcCategory" USING btree ("Scope_Code");

CREATE TABLE public."MedicineForm" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Name" varchar(50) NULL,
	CONSTRAINT "MedicineForm_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_MedicineForm_Code" UNIQUE ("Code")
);

CREATE TABLE public."MedicineGroup" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(10) NULL,
	"Name" varchar(50) NULL,
	CONSTRAINT "MedicineGroup_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_MedicineGroup_Code" UNIQUE ("Code")
);

CREATE TABLE public."MedicineMethod" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(10) NULL,
	"Name" varchar(50) NULL,
	CONSTRAINT "MedicineMethod_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_MedicineMethod_Code" UNIQUE ("Code")
);

CREATE TABLE public."KFR" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"CreatedBy_Employee_Id" int8 NULL,
	"Type" varchar(15) NULL,
	"Subjective" text NULL,
	"Objective" text NULL,
	"Assessment" text NULL,
	"TreatmentGoals" text NULL,
	"Education" text NULL,
	"Action" text NULL,
	"Frequency" int8 NULL,
	"IntervalUnit_Code" varchar(10) NULL,
	"FollowUpType" varchar(10) NULL,
	"FollowUpNote" text NULL,
	"Status_Code" varchar(10) NOT NULL,
	CONSTRAINT "KFR_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Laborant" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Employee_Id" int8 NULL,
	"IHS_Number" varchar(20) NULL,
	"Code" varchar(20) NULL,
	CONSTRAINT "Laborant_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Laborant_Code" UNIQUE ("Code"),
	CONSTRAINT "uni_Laborant_IHS_Number" UNIQUE ("IHS_Number")
);

CREATE TABLE public."Material" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(10) NULL,
	"Name" varchar(50) NULL,
	"Uom_Code" varchar(10) NULL,
	"Infra_Id" int4 NULL,
	"Stock" int8 NULL,
	"Item_Id" int8 NULL,
	"Infra_Code" varchar(10) NULL,
	"Item_Code" varchar(50) NULL,
	CONSTRAINT "Material_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Material_Code" UNIQUE ("Code")
);

CREATE TABLE public."MaterialOrder" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Status_Code" text NULL,
	"Doctor_Code" varchar(20) NULL,
	CONSTRAINT "MaterialOrder_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."MaterialOrderItem" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"MaterialOrder_Id" int8 NULL,
	"Count" int4 NULL,
	"Material_Code" varchar(10) NULL,
	CONSTRAINT "MaterialOrderItem_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."MaterialPackageItem" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"MaterialPackage_Code" varchar(20) NOT NULL,
	"Material_Code" varchar(20) NOT NULL,
	"Count" int4 NULL,
	CONSTRAINT "MaterialPackageItem_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."McuOrder" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Status_Code" varchar(10) NOT NULL,
	"SpecimenPickTime" timestamptz NULL,
	"ExaminationDate" timestamptz NULL,
	"Number" int2 NULL,
	"Temperature" numeric NULL,
	"UrgencyLevel_Code" varchar(15) NOT NULL,
	"Scope_Code" varchar(10) NULL,
	"Doctor_Code" varchar(20) NULL,
	CONSTRAINT "McuOrder_pkey" PRIMARY KEY ("Id")
);
CREATE INDEX "idx_McuOrder_Scope_Code" ON public."McuOrder" USING btree ("Scope_Code");

CREATE TABLE public."McuOrderItem" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"McuOrder_Id" int8 NULL,
	"Result" text NULL,
	"Status_Code" text NULL,
	"ExaminationDate" timestamptz NULL,
	"McuSrc_Code" varchar(20) NULL,
	"Note" varchar(1024) NULL,
	CONSTRAINT "McuOrderItem_pkey" PRIMARY KEY ("Id")
);
CREATE UNIQUE INDEX idx_order_src ON public."McuOrderItem" USING btree ("McuOrder_Id", "McuSrc_Code");

CREATE TABLE public."McuOrderSubItem" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"McuOrderItem_Id" int8 NULL,
	"Result" text NULL,
	"Status_Code" text NULL,
	"McuSubSrc_Code" varchar(20) NULL,
	CONSTRAINT "McuOrderSubItem_pkey" PRIMARY KEY ("Id")
);
CREATE UNIQUE INDEX idx_order_sub_src ON public."McuOrderSubItem" USING btree ("McuSubSrc_Code", "McuOrderItem_Id");

CREATE TABLE public."McuSrc" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NOT NULL,
	"Name" varchar(50) NULL,
	"McuSrcCategory_Code" varchar(20) NULL,
	"Item_Code" varchar(50) NULL,
	CONSTRAINT "McuSrc_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_McuSrc_Code" UNIQUE ("Code")
);

CREATE TABLE public."McuSubSrc" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Name" varchar(50) NULL,
	"McuSrc_Code" varchar(20) NULL,
	"Item_Code" varchar(50) NULL,
	CONSTRAINT "McuSubSrc_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_McuSubSrc_Code" UNIQUE ("Code")
);

CREATE TABLE public."MedicalActionSrc" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Name" varchar(50) NULL,
	"Type_Code" varchar(20) NULL,
	"Item_Code" varchar(50) NULL,
	CONSTRAINT "MedicalActionSrc_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_MedicalActionSrc_Code" UNIQUE ("Code")
);

CREATE TABLE public."MedicalActionSrcItem" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"MedicalActionSrc_Code" varchar(20) NULL,
	"ProcedureSrc_Code" varchar(10) NULL,
	"Item_Code" varchar(50) NULL,
	CONSTRAINT "MedicalActionSrcItem_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Medication" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"IssuedAt" timestamptz NULL,
	"Status_Code" text NULL,
	"Pharmacist_Code" varchar(20) NULL,
	CONSTRAINT "Medication_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."MedicationItem" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Medication_Id" int8 NULL,
	"IsMix" bool NULL,
	"MedicineMix_Id" int8 NULL,
	"Usage" varchar(255) NULL,
	"Interval" int2 NULL,
	"IntervalUnit_Code" text NULL,
	"IsRedeemed" bool NULL,
	"Quantity" numeric NULL,
	"Note" varchar(1024) NULL,
	"Frequency" int4 NULL,
	"Dose" numeric NULL,
	"Medicine_Code" varchar(10) NULL,
	CONSTRAINT "MedicationItem_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."MedicationItemDist" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"MedicationItem_Id" int8 NULL,
	"DateTime" timestamptz NULL,
	"Remain" numeric NULL,
	"Nurse_Code" varchar(20) NULL,
	CONSTRAINT "MedicationItemDist_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Medicine" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(10) NULL,
	"Name" varchar(50) NULL,
	"MedicineGroup_Code" varchar(10) NULL,
	"MedicineMethod_Code" varchar(10) NULL,
	"Uom_Code" varchar(10) NULL,
	"Dose" int2 NULL,
	"Stock" int8 NULL,
	"Infra_Code" varchar(10) NULL,
	"Item_Code" varchar(50) NULL,
	"MedicineForm_Code" varchar(20) NULL,
	CONSTRAINT "Medicine_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Medicine_Code" UNIQUE ("Code")
);

CREATE TABLE public."MedicineMix" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Name" varchar(50) NULL,
	"Uom_Code" varchar(10) NULL,
	CONSTRAINT "MedicineMix_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."MedicineMixItem" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"MedicineMix_Id" int8 NULL,
	"Dose" int2 NULL,
	"Note" text NULL,
	"Medicine_Code" varchar(10) NULL,
	CONSTRAINT "MedicineMixItem_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."MicroMcuOrder" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Number" int8 NULL,
	"Doctor_Code" varchar(20) NULL,
	"Stage_Code" varchar(10) NOT NULL,
	"AxillaryTemp" numeric NULL,
	"OtherNotes" text NULL,
	"Status_Code" varchar(10) NOT NULL,
	CONSTRAINT "MicroMcuOrder_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_MicroMcuOrder_Doctor_Code" UNIQUE ("Doctor_Code")
);

CREATE TABLE public."MicroMcuOrderItem" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"MicroMcuOrder_Id" int8 NULL,
	"McuSrc_Code" varchar(20) NULL,
	"Note" varchar(1024) NULL,
	"Result" text NULL,
	"Status_Code" text NULL,
	CONSTRAINT "MicroMcuOrderItem_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Midwife" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Employee_Id" int8 NULL,
	"IHS_Number" varchar(20) NULL,
	"Code" varchar(20) NULL,
	CONSTRAINT "Midwife_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Midwife_Code" UNIQUE ("Code"),
	CONSTRAINT "uni_Midwife_IHS_Number" UNIQUE ("IHS_Number")
);

CREATE TABLE public."Nurse" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Employee_Id" int8 NULL,
	"IHS_Number" varchar(20) NULL,
	"Code" varchar(20) NULL,
	"Specialist_Code" varchar(10) NULL,
	"Infra_Code" varchar(10) NULL,
	CONSTRAINT "Nurse_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Nurse_Code" UNIQUE ("Code"),
	CONSTRAINT "uni_Nurse_IHS_Number" UNIQUE ("IHS_Number")
);

CREATE TABLE public."Nutritionist" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Employee_Id" int8 NULL,
	"IHS_Number" varchar(20) NULL,
	"Code" varchar(20) NULL,
	CONSTRAINT "Nutritionist_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Nutritionist_Code" UNIQUE ("Code"),
	CONSTRAINT "uni_Nutritionist_IHS_Number" UNIQUE ("IHS_Number")
);

CREATE TABLE public."PharmacyCompany" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Name" varchar(100) NULL,
	"Regency_Code" varchar(4) NULL,
	CONSTRAINT "PharmacyCompany_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_PharmacyCompany_Code" UNIQUE ("Code")
);

CREATE TABLE public."ProcedureSrc" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(10) NULL,
	"Name" varchar(2048) NULL,
	"IndName" varchar(2048) NULL,
	CONSTRAINT "ProcedureSrc_pkey" PRIMARY KEY ("Id")
);
CREATE INDEX "idx_ProcedureSrc_Code" ON public."ProcedureSrc" USING btree ("Code");

CREATE TABLE public."Province" (
	"Id" smallserial NOT NULL,
	"Code" varchar(2) NULL,
	"Name" varchar(50) NULL,
	CONSTRAINT "Province_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Province_Code" UNIQUE ("Code")
);

CREATE TABLE public."Regency" (
	"Id" bigserial NOT NULL,
	"Province_Code" varchar(2) NULL,
	"Code" varchar(4) NULL,
	"Name" varchar(50) NULL,
	CONSTRAINT "Regency_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Regency_Code" UNIQUE ("Code")
);

CREATE TABLE public."Patient" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Person_Id" int8 NULL,
	"RegisteredAt" timestamptz NULL,
	"Status_Code" varchar(10) NOT NULL,
	"Number" varchar(15) NULL,
	"NewBornStatus" bool NULL,
	"RegisteredBy_User_Name" varchar(100) NULL,
	"Parent_Number" varchar(15) NULL,
	CONSTRAINT "Patient_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Patient_Number" UNIQUE ("Number")
);

CREATE TABLE public."Person" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Name" varchar(150) NOT NULL,
	"FrontTitle" varchar(50) NULL,
	"EndTitle" varchar(50) NULL,
	"BirthDate" timestamptz NULL,
	"BirthRegency_Code" varchar(4) NULL,
	"Gender_Code" varchar(10) NULL,
	"ResidentIdentityNumber" varchar(16) NULL,
	"PassportNumber" varchar(20) NULL,
	"DrivingLicenseNumber" varchar(20) NULL,
	"Religion_Code" varchar(10) NULL,
	"Education_Code" varchar(10) NULL,
	"Ocupation_Code" varchar(15) NULL,
	"Ocupation_Name" varchar(50) NULL,
	"Ethnic_Code" varchar(20) NULL,
	"Language_Code" varchar(10) NULL,
	"ResidentIdentityFileUrl" varchar(1024) NULL,
	"PassportFileUrl" varchar(1024) NULL,
	"DrivingLicenseFileUrl" varchar(1024) NULL,
	"FamilyIdentityFileUrl" varchar(1024) NULL,
	"Nationality" varchar(50) NULL,
	"CommunicationIssueStatus" bool NULL,
	"Disability" varchar(100) NULL,
	"Confidence" varchar(512) NULL,
	"MaritalStatus_Code" varchar(10) NULL,
	"BirthPlace" text NULL,
	CONSTRAINT "Person_pkey" PRIMARY KEY ("Id")
);
CREATE UNIQUE INDEX idx_driver_license ON public."Person" USING btree ("DrivingLicenseNumber") WHERE ("DeletedAt" IS NULL);
CREATE UNIQUE INDEX idx_passport ON public."Person" USING btree ("PassportNumber") WHERE ("DeletedAt" IS NULL);
CREATE UNIQUE INDEX idx_resident_identity ON public."Person" USING btree ("ResidentIdentityNumber") WHERE ("DeletedAt" IS NULL);

CREATE TABLE public."PersonAddress" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Person_Id" int8 NULL,
	"Address" varchar(150) NULL,
	"Rt" varchar(2) NULL,
	"Rw" varchar(2) NULL,
	"Village_Code" varchar(10) NULL,
	"PostalRegion_Code" varchar(6) NULL,
	"LocationType_Code" varchar(10) NULL,
	CONSTRAINT "PersonAddress_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."PersonContact" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Person_Id" int8 NULL,
	"Type_Code" varchar(15) NULL,
	"Value" varchar(100) NULL,
	CONSTRAINT "PersonContact_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."PersonInsurance" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Person_Id" int8 NULL,
	"InsuranceCompany_Id" int8 NULL,
	"Ref_Number" varchar(20) NULL,
	"DefaultStatus" bool NULL,
	CONSTRAINT "PersonInsurance_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_PersonInsurance_Ref_Number" UNIQUE ("Ref_Number")
);
CREATE UNIQUE INDEX idx_person_insurance ON public."PersonInsurance" USING btree ("Person_Id", "DefaultStatus");

CREATE TABLE public."PersonRelative" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Person_Id" int8 NULL,
	"Relationship_Code" varchar(100) NOT NULL,
	"Name" varchar(100) NULL,
	"Address" varchar(100) NULL,
	"Village_Code" varchar(10) NULL,
	"Gender_Code" varchar(10) NULL,
	"PhoneNumber" varchar(30) NULL,
	"Education_Code" varchar(10) NULL,
	"Occupation_Code" varchar(10) NULL,
	"Occupation_Name" varchar(50) NULL,
	"Responsible" bool NULL,
	CONSTRAINT "PersonRelative_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Pharmacist" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Employee_Id" int8 NULL,
	"IHS_Number" varchar(20) NULL,
	"Code" varchar(20) NULL,
	CONSTRAINT "Pharmacist_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Pharmacist_Code" UNIQUE ("Code"),
	CONSTRAINT "uni_Pharmacist_IHS_Number" UNIQUE ("IHS_Number")
);

CREATE TABLE public."PostalRegion" (
	"Id" bigserial NOT NULL,
	"Village_Code" varchar(10) NULL,
	"Code" varchar(5) NULL,
	CONSTRAINT "PostalRegion_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_PostalRegion_Code" UNIQUE ("Code")
);

CREATE TABLE public."PracticeSchedule" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Specialist_Code" varchar(20) NULL,
	"Day_Code" int2 NULL,
	"StartTime" varchar(5) NULL,
	"EndTime" varchar(5) NULL,
	"Doctor_Code" varchar(20) NULL,
	CONSTRAINT "PracticeSchedule_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Prescription" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"IssuedAt" timestamptz NULL,
	"Status_Code" text NULL,
	"Doctor_Code" varchar(20) NULL,
	"SpecialistIntern_Id" int8 NULL,
	CONSTRAINT "Prescription_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."PrescriptionItem" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Prescription_Id" int8 NULL,
	"IsMix" bool NULL,
	"MedicineMix_Id" int8 NULL,
	"Usage" varchar(255) NULL,
	"Interval" int2 NULL,
	"IntervalUnit_Code" text NULL,
	"Quantity" numeric NULL,
	"Frequency" int4 NULL,
	"Dose" numeric NULL,
	"Medicine_Code" text NULL,
	"Medicine_Id" int8 NULL,
	"Uom_Id" int4 NULL,
	"MedicineForm_Id" int4 NULL,
	CONSTRAINT "PrescriptionItem_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."ProcedureReport" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Date" timestamptz NOT NULL,
	"Doctor_Code" varchar(10) NULL,
	"Operator_Name" text NULL,
	"Assistant_Name" text NULL,
	"Instrumentor_Name" text NULL,
	"Diagnose" varchar(1024) NULL,
	"Nurse_Name" text NULL,
	"Anesthesia_Doctor_Code" varchar(10) NULL,
	"Anesthesia_Nurse_Name" text NULL,
	"ProcedureValue" text NULL,
	"ExecutionValue" text NULL,
	"Type_Code" text NULL,
	CONSTRAINT "ProcedureReport_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."ProcedureRoom" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Infra_Code" varchar(20) NULL,
	"Type_Code" varchar(10) NULL,
	"Specialist_Code" varchar(20) NULL,
	"Subspecialist_Code" varchar(20) NULL,
	CONSTRAINT "ProcedureRoom_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_ProcedureRoom_Code" UNIQUE ("Code"),
	CONSTRAINT "uni_ProcedureRoom_Infra_Code" UNIQUE ("Infra_Code")
);

CREATE TABLE public."ProcedureRoomOrder" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"MaterialPackage_Code" varchar(20) NULL,
	"Status_Code" varchar(20) NULL,
	CONSTRAINT "ProcedureRoomOrder_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."ProcedureRoomOrderItem" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"ProcedureRoomOrder_Id" int8 NULL,
	"ProcedureRoom_Code" varchar(20) NULL,
	"Note" varchar(255) NULL,
	CONSTRAINT "ProcedureRoomOrderItem_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."RadiologyMcuOrder" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Number" int8 NULL,
	"Doctor_Code" varchar(20) NULL,
	"ClinicalNotes" text NULL,
	"OtherNotes" text NULL,
	"Resume" text NULL,
	"Status_Code" varchar(10) NOT NULL,
	CONSTRAINT "RadiologyMcuOrder_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_RadiologyMcuOrder_Doctor_Code" UNIQUE ("Doctor_Code")
);

CREATE TABLE public."RadiologyMcuOrderItem" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"RadiologyMcuOrder_Id" int8 NULL,
	"McuSrc_Code" varchar(20) NULL,
	"Note" varchar(1024) NULL,
	"Result" text NULL,
	"Status_Code" text NULL,
	CONSTRAINT "RadiologyMcuOrderItem_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Registrator" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Employee_Id" int8 NULL,
	"Installation_Code" varchar(20) NULL,
	CONSTRAINT "Registrator_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Rehab" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"AllocatedVisitCount" int8 NULL,
	"Parent_Encounter_Id" int8 NULL,
	"ExpiredAt" timestamptz NULL,
	"VisitMode_Code" text NULL,
	"Status_Code" text NULL,
	"Frequency" int8 NULL,
	"Interval" text NULL,
	CONSTRAINT "Rehab_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."ResponsibleDoctorHist" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"StartedAt" timestamptz NULL,
	"FinishedAt" timestamptz NULL,
	"Doctor_Code" varchar(20) NULL,
	CONSTRAINT "ResponsibleDoctorHist_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Resume" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NOT NULL,
	"Value" text NULL,
	"FileUrl" varchar(1024) NULL,
	"Doctor_Code" varchar(10) NULL,
	"Status_Code" varchar(10) NOT NULL,
	CONSTRAINT "Resume_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."RoomOrder" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Infra_Code" varchar(10) NULL,
	CONSTRAINT "RoomOrder_pkey" PRIMARY KEY ("Id")
);
CREATE INDEX "idx_RoomOrder_Encounter_Id" ON public."RoomOrder" USING btree ("Encounter_Id");

CREATE TABLE public."Sbar" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Employee_Id" int8 NULL,
	"Time" timestamptz NULL,
	"Value" text NULL,
	CONSTRAINT "Sbar_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Screener" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Employee_Id" int8 NULL,
	"IHS_Number" varchar(20) NULL,
	CONSTRAINT "Screener_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Screener_Code" UNIQUE ("Code"),
	CONSTRAINT "uni_Screener_IHS_Number" UNIQUE ("IHS_Number")
);

CREATE TABLE public."Screening" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Employee_Id" int8 NULL,
	"Type" text NULL,
	"Value" text NULL,
	"Status" text NULL,
	"FileUrl" varchar(1024) NULL,
	CONSTRAINT "Screening_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."SharedTreatment" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Doctor_Code" varchar(20) NULL,
	"Status_Code" varchar(10) NOT NULL,
	"CreatedBy_User_Id" int8 NULL,
	"UpdatedBy_User_Id" int8 NULL,
	"DeletedBy_User_Id" int8 NULL,
	CONSTRAINT "SharedTreatment_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Soapi" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Employee_Id" int8 NULL,
	"Time" timestamptz NULL,
	"Value" text NULL,
	"TypeCode" varchar(15) NULL,
	"FileUrl" varchar(1024) NULL,
	CONSTRAINT "Soapi_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Specialist" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Name" varchar(50) NULL,
	"Installation_Code" varchar(20) NULL,
	CONSTRAINT "Specialist_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Specialist_Code" UNIQUE ("Code")
);

CREATE TABLE public."SpecialistIntern" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Person_Id" int8 NULL,
	"User_Id" int8 NULL,
	"Specialist_Code" varchar(10) NULL,
	"Subspecialist_Code" varchar(10) NULL,
	CONSTRAINT "SpecialistIntern_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."SpecialistPosition" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(10) NOT NULL,
	"Name" varchar(30) NOT NULL,
	"HeadStatus" bool NULL,
	"Employee_Id" int8 NULL,
	"Specialist_Code" varchar(10) NULL,
	CONSTRAINT "SpecialistPosition_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_SpecialistPosition_Code" UNIQUE ("Code")
);

CREATE TABLE public."Subspecialist" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Name" varchar(100) NULL,
	"Specialist_Code" varchar(20) NULL,
	CONSTRAINT "Subspecialist_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Subspecialist_Code" UNIQUE ("Code")
);

CREATE TABLE public."SubspecialistPosition" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(10) NOT NULL,
	"Name" varchar(30) NOT NULL,
	"HeadStatus" bool NULL,
	"Employee_Id" int8 NULL,
	"Subspecialist_Code" varchar(10) NULL,
	CONSTRAINT "SubspecialistPosition_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_SubspecialistPosition_Code" UNIQUE ("Code")
);

CREATE TABLE public."Therapist" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(20) NULL,
	"Employee_Id" int8 NULL,
	"IHS_Number" varchar(20) NULL,
	CONSTRAINT "Therapist_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Therapist_Code" UNIQUE ("Code"),
	CONSTRAINT "uni_Therapist_IHS_Number" UNIQUE ("IHS_Number")
);

CREATE TABLE public."TherapyProgramForm" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"CreatedBy_Employee_Id" int8 NULL,
	"Subjective" text NULL,
	"Objective" text NULL,
	"Assessment" text NULL,
	"Procedure" text NULL,
	"Status_Code" varchar(10) NOT NULL,
	CONSTRAINT "TherapyProgramForm_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."TherapyProtocol" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NOT NULL,
	"Anamnesis" varchar(2048) NULL,
	"MedicalDiagnoses" text NULL,
	"FunctionDiagnoses" text NULL,
	"Procedures" text NULL,
	"SupportingExams" varchar(2048) NULL,
	"Instruction" varchar(2048) NULL,
	"Evaluation" varchar(2048) NULL,
	"WorkCauseStatus" varchar(2048) NULL,
	"Frequency" int8 NULL,
	"IntervalUnit_Code" varchar(10) NULL,
	"Duration" int8 NULL,
	"DurationUnit_Code" varchar(10) NULL,
	"Doctor_Code" varchar(20) NULL,
	"Status_Code" varchar(10) NULL,
	CONSTRAINT "TherapyProtocol_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Uom" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(10) NULL,
	"Name" varchar(50) NULL,
	CONSTRAINT "Uom_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Uom_Code" UNIQUE ("Code")
);

CREATE TABLE public."User" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Name" varchar(50) NOT NULL,
	"Password" varchar(255) NOT NULL,
	"Status_Code" varchar(10) NOT NULL,
	"FailedLoginCount" int2 NULL,
	"ContractPosition_Code" varchar(20) NOT NULL,
	"LoginAttemptCount" int8 NULL,
	"LastSuccessLogin" timestamptz NULL,
	"LastAllowdLogin" timestamptz NULL,
	CONSTRAINT "User_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_User_Name" UNIQUE ("Name")
);

CREATE TABLE public."VclaimSepHist" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"RequestPayload" text NULL,
	"ResponseBody" text NULL,
	"Message" text NULL,
	CONSTRAINT "VclaimSepHist_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Vehicle" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Type_Code" text NULL,
	"PoliceNumber" text NULL,
	"FrameNumber" text NULL,
	"RegNumber" text NULL,
	"AvailableStatus" bool NULL,
	CONSTRAINT "Vehicle_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Unit" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Code" varchar(60) NULL,
	"Name" varchar(60) NULL,
	"Level" int2 NULL,
	"Parent_Id" int4 NULL,
	CONSTRAINT "Unit_pkey" PRIMARY KEY ("Id")
);
CREATE UNIQUE INDEX "idx_Unit_Code" ON public."Unit" USING btree ("Code");

CREATE TABLE public."VehicleHist" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Vehicle_Id" int8 NULL,
	"Date" timestamptz NULL,
	"Data" text NULL,
	"Crud_Code" text NULL,
	CONSTRAINT "VehicleHist_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."UnitPosition" (
	"Id" serial4 NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Unit_Code" varchar(10) NULL,
	"Functional_Code" varchar(10) NULL,
	"Name" varchar(50) NULL,
	"HeadStatus" bool NULL,
	"Employee_Id" int8 NULL,
	CONSTRAINT "UnitPosition_pkey" PRIMARY KEY ("Id")
);
CREATE UNIQUE INDEX unit_position_idx ON public."UnitPosition" USING btree ("Unit_Code", "Functional_Code");

CREATE TABLE public."UserFes" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Name" varchar(100) NULL,
	"AuthPartner_Code" varchar(50) NULL,
	"User_Name" varchar(50) NULL,
	CONSTRAINT "UserFes_pkey" PRIMARY KEY ("Id")
);
CREATE UNIQUE INDEX "idx-userFes-name-authPartner_code" ON public."UserFes" USING btree ("Name", "AuthPartner_Code");

CREATE TABLE public."VaccineData" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Type" varchar(25) NULL,
	"Encounter_Id" int8 NULL,
	"BatchNumber" text NULL,
	"Dose" numeric NULL,
	"DoseOrder" int8 NULL,
	"InjectionLocation" text NULL,
	"GivenDate" timestamptz NULL,
	"ExpirationDate" timestamptz NULL,
	CONSTRAINT "VaccineData_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."VclaimMember" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"CardNumber" varchar(20) NULL,
	"Person_Id" int8 NULL,
	CONSTRAINT "VclaimMember_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_VclaimMember_CardNumber" UNIQUE ("CardNumber")
);

CREATE TABLE public."VclaimReference" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Date" timestamptz NULL,
	"SrcCode" text NULL,
	"SrcName" text NULL,
	"Number" text NULL,
	CONSTRAINT "VclaimReference_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."VclaimSep" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"Encounter_Id" int8 NULL,
	"Number" varchar(19) NULL,
	CONSTRAINT "VclaimSep_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_VclaimSep_Number" UNIQUE ("Number")
);

CREATE TABLE public."VclaimSepControlLetter" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"VclaimSep_Number" varchar(19) NULL,
	"Number" varchar(20) NULL,
	"Value" text NULL,
	"FileUrl" varchar(1024) NULL,
	CONSTRAINT "VclaimSepControlLetter_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_VclaimSepControlLetter_FileUrl" UNIQUE ("FileUrl"),
	CONSTRAINT "uni_VclaimSepControlLetter_Number" UNIQUE ("Number")
);

CREATE TABLE public."VclaimSepPrint" (
	"Id" bigserial NOT NULL,
	"CreatedAt" timestamptz NULL,
	"UpdatedAt" timestamptz NULL,
	"DeletedAt" timestamptz NULL,
	"VclaimSep_Number" varchar(19) NULL,
	"Counter" int8 NULL,
	CONSTRAINT "VclaimSepPrint_pkey" PRIMARY KEY ("Id")
);

CREATE TABLE public."Village" (
	"Id" bigserial NOT NULL,
	"District_Code" varchar(6) NULL,
	"Code" varchar(10) NULL,
	"Name" varchar(50) NULL,
	CONSTRAINT "Village_pkey" PRIMARY KEY ("Id"),
	CONSTRAINT "uni_Village_Code" UNIQUE ("Code")
);

-- =============================================
-- BAGIAN 2: ALTER TABLE (Menambahkan Relasi)
-- =============================================

ALTER TABLE public."Nutritionist" ADD CONSTRAINT "fk_Nutritionist_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");

ALTER TABLE public."Adime" ADD CONSTRAINT "fk_Adime_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");
ALTER TABLE public."Adime" ADD CONSTRAINT "fk_Adime_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."AdmEmployeeHist" ADD CONSTRAINT "fk_AdmEmployeeHist_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");

ALTER TABLE public."AmbulanceTransportReq" ADD CONSTRAINT "fk_AmbulanceTransportReq_District" FOREIGN KEY ("District_Code") REFERENCES public."District"("Code");
ALTER TABLE public."AmbulanceTransportReq" ADD CONSTRAINT "fk_AmbulanceTransportReq_Patient" FOREIGN KEY ("Patient_Id") REFERENCES public."Patient"("Id");
ALTER TABLE public."AmbulanceTransportReq" ADD CONSTRAINT "fk_AmbulanceTransportReq_Province" FOREIGN KEY ("Province_Code") REFERENCES public."Province"("Code");
ALTER TABLE public."AmbulanceTransportReq" ADD CONSTRAINT "fk_AmbulanceTransportReq_Regency" FOREIGN KEY ("Regency_Code") REFERENCES public."Regency"("Code");
ALTER TABLE public."AmbulanceTransportReq" ADD CONSTRAINT "fk_AmbulanceTransportReq_Village" FOREIGN KEY ("Village_Code") REFERENCES public."Village"("Code");

ALTER TABLE public."Ambulatory" ADD CONSTRAINT "fk_Encounter_Ambulatory" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."AntibioticInUse" ADD CONSTRAINT "fk_AntibioticInUse_AntibioticSrc" FOREIGN KEY ("AntibioticSrc_Id") REFERENCES public."AntibioticSrc"("Id");
ALTER TABLE public."AntibioticInUse" ADD CONSTRAINT "fk_AntibioticInUse_McuOrder" FOREIGN KEY ("McuOrder_Id") REFERENCES public."McuOrder"("Id");

ALTER TABLE public."AntibioticSrc" ADD CONSTRAINT "fk_AntibioticSrc_AntibioticSrcCategory" FOREIGN KEY ("AntibioticSrcCategory_Code") REFERENCES public."AntibioticSrcCategory"("Code");

ALTER TABLE public."ApMcuOrder" ADD CONSTRAINT "fk_ApMcuOrder_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."ApMcuOrder" ADD CONSTRAINT "fk_ApMcuOrder_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."Appointment" ADD CONSTRAINT "fk_Appointment_Patient" FOREIGN KEY ("Patient_Id") REFERENCES public."Patient"("Id");
ALTER TABLE public."Appointment" ADD CONSTRAINT "fk_Appointment_PracticeSchedule" FOREIGN KEY ("PracticeSchedule_Id") REFERENCES public."PracticeSchedule"("Id");

ALTER TABLE public."Bed" ADD CONSTRAINT "fk_Infra_Beds" FOREIGN KEY ("Infra_Id") REFERENCES public."Infra"("Id");

ALTER TABLE public."Billing" ADD CONSTRAINT "fk_Billing_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."Chamber" ADD CONSTRAINT "fk_Chamber_Class" FOREIGN KEY ("Class_Code") REFERENCES public."ChamberClass"("Code");
ALTER TABLE public."Chamber" ADD CONSTRAINT "fk_Infra_Chambers" FOREIGN KEY ("Infra_Id") REFERENCES public."Infra"("Id");

ALTER TABLE public."Chemo" ADD CONSTRAINT "fk_Chemo_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."Chemo" ADD CONSTRAINT "fk_Chemo_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");
ALTER TABLE public."Chemo" ADD CONSTRAINT "fk_Chemo_Specialist" FOREIGN KEY ("Specialist_Code") REFERENCES public."Specialist"("Code");
ALTER TABLE public."Chemo" ADD CONSTRAINT "fk_Chemo_VerifiedBy" FOREIGN KEY ("VerifiedBy_User_Id") REFERENCES public."User"("Id");

ALTER TABLE public."ChemoPlan" ADD CONSTRAINT "fk_ChemoPlan_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");
ALTER TABLE public."ChemoPlan" ADD CONSTRAINT "fk_ChemoProtocol_ChemoPlans" FOREIGN KEY ("Protocol_Id") REFERENCES public."ChemoProtocol"("Id");

ALTER TABLE public."ChemoProtocol" ADD CONSTRAINT "fk_ChemoProtocol_Chemo" FOREIGN KEY ("Chemo_Id") REFERENCES public."Chemo"("Id");
ALTER TABLE public."ChemoProtocol" ADD CONSTRAINT "fk_ChemoProtocol_VerifiedBy" FOREIGN KEY ("VerifiedBy_User_Id") REFERENCES public."User"("Id");

ALTER TABLE public."Consultation" ADD CONSTRAINT "fk_Consultation_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."Consultation" ADD CONSTRAINT "fk_Consultation_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");
ALTER TABLE public."Consultation" ADD CONSTRAINT "fk_Consultation_Unit" FOREIGN KEY ("Unit_Code") REFERENCES public."Unit"("Code");

ALTER TABLE public."ControlLetter" ADD CONSTRAINT "fk_ControlLetter_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."ControlLetter" ADD CONSTRAINT "fk_ControlLetter_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");
ALTER TABLE public."ControlLetter" ADD CONSTRAINT "fk_ControlLetter_Specialist" FOREIGN KEY ("Specialist_Code") REFERENCES public."Specialist"("Code");
ALTER TABLE public."ControlLetter" ADD CONSTRAINT "fk_ControlLetter_Subspecialist" FOREIGN KEY ("Subspecialist_Code") REFERENCES public."Subspecialist"("Code");

ALTER TABLE public."CpMcuOrder" ADD CONSTRAINT "fk_CpMcuOrder_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."CpMcuOrder" ADD CONSTRAINT "fk_CpMcuOrder_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."CpMcuOrderItem" ADD CONSTRAINT "fk_CpMcuOrderItem_CpMcuOrder" FOREIGN KEY ("CpMcuOrder_Id") REFERENCES public."CpMcuOrder"("Id");
ALTER TABLE public."CpMcuOrderItem" ADD CONSTRAINT "fk_CpMcuOrderItem_McuSrc" FOREIGN KEY ("McuSrc_Code") REFERENCES public."McuSrc"("Code");

ALTER TABLE public."DeathCause" ADD CONSTRAINT "fk_Encounter_DeathCause" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."Device" ADD CONSTRAINT "fk_Device_Infra" FOREIGN KEY ("Infra_Id") REFERENCES public."Infra"("Id");
ALTER TABLE public."Device" ADD CONSTRAINT "fk_Device_Item" FOREIGN KEY ("Item_Id") REFERENCES public."Item"("Id");
ALTER TABLE public."Device" ADD CONSTRAINT "fk_Device_Uom" FOREIGN KEY ("Uom_Code") REFERENCES public."Uom"("Code");

ALTER TABLE public."DeviceOrder" ADD CONSTRAINT "fk_DeviceOrder_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."DeviceOrder" ADD CONSTRAINT "fk_DeviceOrder_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."DeviceOrderItem" ADD CONSTRAINT "fk_DeviceOrderItem_Device" FOREIGN KEY ("Device_Code") REFERENCES public."Device"("Code");
ALTER TABLE public."DeviceOrderItem" ADD CONSTRAINT "fk_DeviceOrderItem_DeviceOrder" FOREIGN KEY ("DeviceOrder_Id") REFERENCES public."DeviceOrder"("Id");

ALTER TABLE public."DevicePackageItem" ADD CONSTRAINT "fk_DevicePackageItem_Device" FOREIGN KEY ("Device_Code") REFERENCES public."Device"("Code");
ALTER TABLE public."DevicePackageItem" ADD CONSTRAINT "fk_DevicePackageItem_DevicePackage" FOREIGN KEY ("DevicePackage_Code") REFERENCES public."DevicePackage"("Code");

ALTER TABLE public."District" ADD CONSTRAINT "fk_District_Regency" FOREIGN KEY ("Regency_Code") REFERENCES public."Regency"("Code");

ALTER TABLE public."Division" ADD CONSTRAINT "fk_Division_Childrens" FOREIGN KEY ("Parent_Code") REFERENCES public."Division"("Code");

ALTER TABLE public."DivisionPosition" ADD CONSTRAINT "fk_DivisionPosition_Division" FOREIGN KEY ("Division_Code") REFERENCES public."Division"("Code");
ALTER TABLE public."DivisionPosition" ADD CONSTRAINT "fk_DivisionPosition_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");

ALTER TABLE public."Doctor" ADD CONSTRAINT "fk_Doctor_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");
ALTER TABLE public."Doctor" ADD CONSTRAINT "fk_Doctor_Unit" FOREIGN KEY ("Unit_Code") REFERENCES public."Unit"("Code");

ALTER TABLE public."DoctorFee" ADD CONSTRAINT "fk_DoctorFee_Doctor" FOREIGN KEY ("Doctor_Id") REFERENCES public."Doctor"("Id");
ALTER TABLE public."DoctorFee" ADD CONSTRAINT "fk_DoctorFee_Item" FOREIGN KEY ("Item_Id") REFERENCES public."Item"("Id");

ALTER TABLE public."DoctorUnit" ADD CONSTRAINT "fk_DoctorUnit_Doctor" FOREIGN KEY ("Doctor_Id") REFERENCES public."Doctor"("Id");
ALTER TABLE public."DoctorUnit" ADD CONSTRAINT "fk_DoctorUnit_Unit" FOREIGN KEY ("Unit_Id") REFERENCES public."Unit"("Id");

ALTER TABLE public."EduAssessment" ADD CONSTRAINT "fk_EduAssessment_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."EduAssessmentImpl" ADD CONSTRAINT "fk_EduAssessmentImpl_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");
ALTER TABLE public."EduAssessmentImpl" ADD CONSTRAINT "fk_EduAssessment_EduAssesmentImpl" FOREIGN KEY ("EduAssessment_Id") REFERENCES public."EduAssessment"("Id");

ALTER TABLE public."Emergency" ADD CONSTRAINT "fk_Encounter_Emergency" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."Employee" ADD CONSTRAINT "fk_Employee_Person" FOREIGN KEY ("Person_Id") REFERENCES public."Person"("Id");
ALTER TABLE public."Employee" ADD CONSTRAINT "fk_Employee_User" FOREIGN KEY ("User_Id") REFERENCES public."User"("Id");

ALTER TABLE public."Encounter" ADD CONSTRAINT "fk_Encounter_Adm_Employee" FOREIGN KEY ("Adm_Employee_Id") REFERENCES public."Employee"("Id");
ALTER TABLE public."Encounter" ADD CONSTRAINT "fk_Encounter_Appointment" FOREIGN KEY ("Appointment_Id") REFERENCES public."Appointment"("Id");
ALTER TABLE public."Encounter" ADD CONSTRAINT "fk_Encounter_Appointment_Doctor" FOREIGN KEY ("Appointment_Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."Encounter" ADD CONSTRAINT "fk_Encounter_InsuranceCompany" FOREIGN KEY ("InsuranceCompany_Code") REFERENCES public."InsuranceCompany"("Code");
ALTER TABLE public."Encounter" ADD CONSTRAINT "fk_Encounter_Patient" FOREIGN KEY ("Patient_Id") REFERENCES public."Patient"("Id");
ALTER TABLE public."Encounter" ADD CONSTRAINT "fk_Encounter_Responsible_Doctor" FOREIGN KEY ("Responsible_Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."Encounter" ADD CONSTRAINT "fk_Encounter_Responsible_Nurse" FOREIGN KEY ("Responsible_Nurse_Code") REFERENCES public."Nurse"("Code");
ALTER TABLE public."Encounter" ADD CONSTRAINT "fk_Encounter_Specialist" FOREIGN KEY ("Specialist_Code") REFERENCES public."Specialist"("Code");
ALTER TABLE public."Encounter" ADD CONSTRAINT "fk_Encounter_Subspecialist" FOREIGN KEY ("Subspecialist_Code") REFERENCES public."Subspecialist"("Code");
ALTER TABLE public."Encounter" ADD CONSTRAINT "fk_Encounter_Unit" FOREIGN KEY ("Unit_Code") REFERENCES public."Unit"("Code");

ALTER TABLE public."EncounterDocument" ADD CONSTRAINT "fk_EncounterDocument_Upload_Employee" FOREIGN KEY ("Upload_Employee_Id") REFERENCES public."Employee"("Id");
ALTER TABLE public."EncounterDocument" ADD CONSTRAINT "fk_Encounter_EncounterDocuments" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."GeneralConsent" ADD CONSTRAINT "fk_Encounter_GeneralConsents" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."Infra" ADD CONSTRAINT "fk_Infra_Childrens" FOREIGN KEY ("Parent_Code") REFERENCES public."Infra"("Code");
ALTER TABLE public."Infra" ADD CONSTRAINT "fk_Infra_Item" FOREIGN KEY ("Item_Code") REFERENCES public."Item"("Code");

ALTER TABLE public."Inpatient" ADD CONSTRAINT "fk_Encounter_Inpatient" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");
ALTER TABLE public."Inpatient" ADD CONSTRAINT "fk_Inpatient_Infra" FOREIGN KEY ("Infra_Code") REFERENCES public."Infra"("Code");

ALTER TABLE public."InstallationPosition" ADD CONSTRAINT "fk_InstallationPosition_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");
ALTER TABLE public."InstallationPosition" ADD CONSTRAINT "fk_InstallationPosition_Installation" FOREIGN KEY ("Installation_Code") REFERENCES public."Installation"("Code");

ALTER TABLE public."InsuranceCompany" ADD CONSTRAINT "fk_InsuranceCompany_Regency" FOREIGN KEY ("Regency_Code") REFERENCES public."Regency"("Code");

ALTER TABLE public."Intern" ADD CONSTRAINT "fk_Intern_Person" FOREIGN KEY ("Person_Id") REFERENCES public."Person"("Id");
ALTER TABLE public."Intern" ADD CONSTRAINT "fk_Intern_User" FOREIGN KEY ("User_Id") REFERENCES public."User"("Id");

ALTER TABLE public."InternalReference" ADD CONSTRAINT "fk_Encounter_InternalReferences" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");
ALTER TABLE public."InternalReference" ADD CONSTRAINT "fk_InternalReference_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."InternalReference" ADD CONSTRAINT "fk_InternalReference_Nurse" FOREIGN KEY ("Nurse_Code") REFERENCES public."Nurse"("Code");
ALTER TABLE public."InternalReference" ADD CONSTRAINT "fk_InternalReference_SrcDoctor" FOREIGN KEY ("SrcDoctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."InternalReference" ADD CONSTRAINT "fk_InternalReference_SrcNurse" FOREIGN KEY ("SrcNurse_Code") REFERENCES public."Nurse"("Code");
ALTER TABLE public."InternalReference" ADD CONSTRAINT "fk_InternalReference_Unit" FOREIGN KEY ("Unit_Code") REFERENCES public."Unit"("Code");

ALTER TABLE public."Item" ADD CONSTRAINT "fk_Item_Uom" FOREIGN KEY ("Uom_Code") REFERENCES public."Uom"("Code");

ALTER TABLE public."ItemPrice" ADD CONSTRAINT "fk_ItemPrice_InsuranceCompany" FOREIGN KEY ("InsuranceCompany_Code") REFERENCES public."InsuranceCompany"("Code");
ALTER TABLE public."ItemPrice" ADD CONSTRAINT "fk_ItemPrice_Item" FOREIGN KEY ("Item_Id") REFERENCES public."Item"("Id");

ALTER TABLE public."KFR" ADD CONSTRAINT "fk_KFR_CreatedBy_Employee" FOREIGN KEY ("CreatedBy_Employee_Id") REFERENCES public."Employee"("Id");
ALTER TABLE public."KFR" ADD CONSTRAINT "fk_KFR_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."Laborant" ADD CONSTRAINT "fk_Laborant_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");

ALTER TABLE public."Material" ADD CONSTRAINT "fk_Material_Infra" FOREIGN KEY ("Infra_Id") REFERENCES public."Infra"("Id");
ALTER TABLE public."Material" ADD CONSTRAINT "fk_Material_Item" FOREIGN KEY ("Item_Id") REFERENCES public."Item"("Id");
ALTER TABLE public."Material" ADD CONSTRAINT "fk_Material_Uom" FOREIGN KEY ("Uom_Code") REFERENCES public."Uom"("Code");

ALTER TABLE public."MaterialOrder" ADD CONSTRAINT "fk_MaterialOrder_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."MaterialOrder" ADD CONSTRAINT "fk_MaterialOrder_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."MaterialOrderItem" ADD CONSTRAINT "fk_MaterialOrderItem_Material" FOREIGN KEY ("Material_Code") REFERENCES public."Material"("Code");
ALTER TABLE public."MaterialOrderItem" ADD CONSTRAINT "fk_MaterialOrderItem_MaterialOrder" FOREIGN KEY ("MaterialOrder_Id") REFERENCES public."MaterialOrder"("Id");

ALTER TABLE public."MaterialPackageItem" ADD CONSTRAINT "fk_MaterialPackageItem_Material" FOREIGN KEY ("Material_Code") REFERENCES public."Material"("Code");
ALTER TABLE public."MaterialPackageItem" ADD CONSTRAINT "fk_MaterialPackageItem_MaterialPackage" FOREIGN KEY ("MaterialPackage_Code") REFERENCES public."MaterialPackage"("Code");

ALTER TABLE public."McuOrder" ADD CONSTRAINT "fk_McuOrder_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."McuOrder" ADD CONSTRAINT "fk_McuOrder_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."McuOrderItem" ADD CONSTRAINT "fk_McuOrderItem_McuOrder" FOREIGN KEY ("McuOrder_Id") REFERENCES public."McuOrder"("Id");
ALTER TABLE public."McuOrderItem" ADD CONSTRAINT "fk_McuOrderItem_McuSrc" FOREIGN KEY ("McuSrc_Code") REFERENCES public."McuSrc"("Code");

ALTER TABLE public."McuOrderSubItem" ADD CONSTRAINT "fk_McuOrderSubItem_McuOrderItem" FOREIGN KEY ("McuOrderItem_Id") REFERENCES public."McuOrderItem"("Id");
ALTER TABLE public."McuOrderSubItem" ADD CONSTRAINT "fk_McuOrderSubItem_McuSubSrc" FOREIGN KEY ("McuSubSrc_Code") REFERENCES public."McuSubSrc"("Code");

ALTER TABLE public."McuSrc" ADD CONSTRAINT "fk_McuSrc_Item" FOREIGN KEY ("Item_Code") REFERENCES public."Item"("Code");
ALTER TABLE public."McuSrc" ADD CONSTRAINT "fk_McuSrc_McuSrcCategory" FOREIGN KEY ("McuSrcCategory_Code") REFERENCES public."McuSrcCategory"("Code");

ALTER TABLE public."McuSubSrc" ADD CONSTRAINT "fk_McuSubSrc_Item" FOREIGN KEY ("Item_Code") REFERENCES public."Item"("Code");
ALTER TABLE public."McuSubSrc" ADD CONSTRAINT "fk_McuSubSrc_McuSrc" FOREIGN KEY ("McuSrc_Code") REFERENCES public."McuSrc"("Code");

ALTER TABLE public."MedicalActionSrc" ADD CONSTRAINT "fk_MedicalActionSrc_Item" FOREIGN KEY ("Item_Code") REFERENCES public."Item"("Code");

ALTER TABLE public."MedicalActionSrcItem" ADD CONSTRAINT "fk_MedicalActionSrcItem_Item" FOREIGN KEY ("Item_Code") REFERENCES public."Item"("Code");
ALTER TABLE public."MedicalActionSrcItem" ADD CONSTRAINT "fk_MedicalActionSrcItem_MedicalActionSrc" FOREIGN KEY ("MedicalActionSrc_Code") REFERENCES public."MedicalActionSrc"("Code");

ALTER TABLE public."Medication" ADD CONSTRAINT "fk_Medication_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");
ALTER TABLE public."Medication" ADD CONSTRAINT "fk_Medication_Pharmacist" FOREIGN KEY ("Pharmacist_Code") REFERENCES public."Pharmacist"("Code");

ALTER TABLE public."MedicationItem" ADD CONSTRAINT "fk_MedicationItem_Medication" FOREIGN KEY ("Medication_Id") REFERENCES public."Medication"("Id");
ALTER TABLE public."MedicationItem" ADD CONSTRAINT "fk_MedicationItem_Medicine" FOREIGN KEY ("Medicine_Code") REFERENCES public."Medicine"("Code");
ALTER TABLE public."MedicationItem" ADD CONSTRAINT "fk_MedicationItem_MedicineMix" FOREIGN KEY ("MedicineMix_Id") REFERENCES public."MedicineMix"("Id");

ALTER TABLE public."MedicationItemDist" ADD CONSTRAINT "fk_MedicationItemDist_MedicationItem" FOREIGN KEY ("MedicationItem_Id") REFERENCES public."MedicationItem"("Id");
ALTER TABLE public."MedicationItemDist" ADD CONSTRAINT "fk_MedicationItemDist_Nurse" FOREIGN KEY ("Nurse_Code") REFERENCES public."Nurse"("Code");

ALTER TABLE public."Medicine" ADD CONSTRAINT "fk_Medicine_Infra" FOREIGN KEY ("Infra_Code") REFERENCES public."Infra"("Code");
ALTER TABLE public."Medicine" ADD CONSTRAINT "fk_Medicine_Item" FOREIGN KEY ("Item_Code") REFERENCES public."Item"("Code");
ALTER TABLE public."Medicine" ADD CONSTRAINT "fk_Medicine_MedicineForm" FOREIGN KEY ("MedicineForm_Code") REFERENCES public."MedicineForm"("Code");
ALTER TABLE public."Medicine" ADD CONSTRAINT "fk_Medicine_MedicineGroup" FOREIGN KEY ("MedicineGroup_Code") REFERENCES public."MedicineGroup"("Code");
ALTER TABLE public."Medicine" ADD CONSTRAINT "fk_Medicine_MedicineMethod" FOREIGN KEY ("MedicineMethod_Code") REFERENCES public."MedicineMethod"("Code");
ALTER TABLE public."Medicine" ADD CONSTRAINT "fk_Medicine_Uom" FOREIGN KEY ("Uom_Code") REFERENCES public."Uom"("Code");

ALTER TABLE public."MedicineMix" ADD CONSTRAINT "fk_MedicineMix_Uom" FOREIGN KEY ("Uom_Code") REFERENCES public."Uom"("Code");

ALTER TABLE public."MedicineMixItem" ADD CONSTRAINT "fk_MedicineMixItem_Medicine" FOREIGN KEY ("Medicine_Code") REFERENCES public."Medicine"("Code");
ALTER TABLE public."MedicineMixItem" ADD CONSTRAINT "fk_MedicineMix_MixItems" FOREIGN KEY ("MedicineMix_Id") REFERENCES public."MedicineMix"("Id");

ALTER TABLE public."MicroMcuOrder" ADD CONSTRAINT "fk_MicroMcuOrder_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."MicroMcuOrder" ADD CONSTRAINT "fk_MicroMcuOrder_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."MicroMcuOrderItem" ADD CONSTRAINT "fk_MicroMcuOrderItem_McuSrc" FOREIGN KEY ("McuSrc_Code") REFERENCES public."McuSrc"("Code");
ALTER TABLE public."MicroMcuOrderItem" ADD CONSTRAINT "fk_MicroMcuOrderItem_MicroMcuOrder" FOREIGN KEY ("MicroMcuOrder_Id") REFERENCES public."MicroMcuOrder"("Id");

ALTER TABLE public."Midwife" ADD CONSTRAINT "fk_Midwife_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");

ALTER TABLE public."Nurse" ADD CONSTRAINT "fk_Nurse_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");
ALTER TABLE public."Nurse" ADD CONSTRAINT "fk_Nurse_Infra" FOREIGN KEY ("Infra_Code") REFERENCES public."Infra"("Code");
ALTER TABLE public."Nurse" ADD CONSTRAINT "fk_Nurse_Specialist" FOREIGN KEY ("Specialist_Code") REFERENCES public."Specialist"("Code");

ALTER TABLE public."Patient" ADD CONSTRAINT "fk_Patient_Parent" FOREIGN KEY ("Parent_Number") REFERENCES public."Patient"("Number");
ALTER TABLE public."Patient" ADD CONSTRAINT "fk_Patient_Person" FOREIGN KEY ("Person_Id") REFERENCES public."Person"("Id");

ALTER TABLE public."Person" ADD CONSTRAINT "fk_Person_BirthRegency" FOREIGN KEY ("BirthRegency_Code") REFERENCES public."Regency"("Code");
ALTER TABLE public."Person" ADD CONSTRAINT "fk_Person_Ethnic" FOREIGN KEY ("Ethnic_Code") REFERENCES public."Ethnic"("Code");
ALTER TABLE public."Person" ADD CONSTRAINT "fk_Person_Language" FOREIGN KEY ("Language_Code") REFERENCES public."Language"("Code");

ALTER TABLE public."PersonAddress" ADD CONSTRAINT "fk_PersonAddress_PostalRegion" FOREIGN KEY ("PostalRegion_Code") REFERENCES public."PostalRegion"("Code");
ALTER TABLE public."PersonAddress" ADD CONSTRAINT "fk_PersonAddress_Village" FOREIGN KEY ("Village_Code") REFERENCES public."Village"("Code");
ALTER TABLE public."PersonAddress" ADD CONSTRAINT "fk_Person_Addresses" FOREIGN KEY ("Person_Id") REFERENCES public."Person"("Id");

ALTER TABLE public."PersonContact" ADD CONSTRAINT "fk_Person_Contacts" FOREIGN KEY ("Person_Id") REFERENCES public."Person"("Id");

ALTER TABLE public."PersonInsurance" ADD CONSTRAINT "fk_PersonInsurance_InsuranceCompany" FOREIGN KEY ("InsuranceCompany_Id") REFERENCES public."InsuranceCompany"("Id");
ALTER TABLE public."PersonInsurance" ADD CONSTRAINT "fk_Person_Insurances" FOREIGN KEY ("Person_Id") REFERENCES public."Person"("Id");

ALTER TABLE public."PersonRelative" ADD CONSTRAINT "fk_PersonRelative_Village" FOREIGN KEY ("Village_Code") REFERENCES public."Village"("Code");
ALTER TABLE public."PersonRelative" ADD CONSTRAINT "fk_Person_Relatives" FOREIGN KEY ("Person_Id") REFERENCES public."Person"("Id");

ALTER TABLE public."Pharmacist" ADD CONSTRAINT "fk_Pharmacist_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");

ALTER TABLE public."PostalRegion" ADD CONSTRAINT "fk_PostalRegion_Village" FOREIGN KEY ("Village_Code") REFERENCES public."Village"("Code");

ALTER TABLE public."PracticeSchedule" ADD CONSTRAINT "fk_PracticeSchedule_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."PracticeSchedule" ADD CONSTRAINT "fk_PracticeSchedule_Specialist" FOREIGN KEY ("Specialist_Code") REFERENCES public."Specialist"("Code");

ALTER TABLE public."Prescription" ADD CONSTRAINT "fk_Prescription_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."Prescription" ADD CONSTRAINT "fk_Prescription_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");
ALTER TABLE public."Prescription" ADD CONSTRAINT "fk_Prescription_SpecialistIntern" FOREIGN KEY ("SpecialistIntern_Id") REFERENCES public."SpecialistIntern"("Id");

ALTER TABLE public."PrescriptionItem" ADD CONSTRAINT "fk_PrescriptionItem_Medicine" FOREIGN KEY ("Medicine_Id") REFERENCES public."Medicine"("Id");
ALTER TABLE public."PrescriptionItem" ADD CONSTRAINT "fk_PrescriptionItem_MedicineForm" FOREIGN KEY ("MedicineForm_Id") REFERENCES public."MedicineForm"("Id");
ALTER TABLE public."PrescriptionItem" ADD CONSTRAINT "fk_PrescriptionItem_MedicineMix" FOREIGN KEY ("MedicineMix_Id") REFERENCES public."MedicineMix"("Id");
ALTER TABLE public."PrescriptionItem" ADD CONSTRAINT "fk_PrescriptionItem_Prescription" FOREIGN KEY ("Prescription_Id") REFERENCES public."Prescription"("Id");
ALTER TABLE public."PrescriptionItem" ADD CONSTRAINT "fk_PrescriptionItem_Uom" FOREIGN KEY ("Uom_Id") REFERENCES public."Uom"("Id");

ALTER TABLE public."ProcedureReport" ADD CONSTRAINT "fk_ProcedureReport_Anesthesia_Doctor" FOREIGN KEY ("Anesthesia_Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."ProcedureReport" ADD CONSTRAINT "fk_ProcedureReport_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."ProcedureReport" ADD CONSTRAINT "fk_ProcedureReport_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."ProcedureRoom" ADD CONSTRAINT "fk_ProcedureRoom_Infra" FOREIGN KEY ("Infra_Code") REFERENCES public."Infra"("Code");
ALTER TABLE public."ProcedureRoom" ADD CONSTRAINT "fk_ProcedureRoom_Specialist" FOREIGN KEY ("Specialist_Code") REFERENCES public."Specialist"("Code");
ALTER TABLE public."ProcedureRoom" ADD CONSTRAINT "fk_ProcedureRoom_Subspecialist" FOREIGN KEY ("Subspecialist_Code") REFERENCES public."Subspecialist"("Code");

ALTER TABLE public."ProcedureRoomOrder" ADD CONSTRAINT "fk_ProcedureRoomOrder_MaterialPackage" FOREIGN KEY ("MaterialPackage_Code") REFERENCES public."MaterialPackage"("Code");

ALTER TABLE public."ProcedureRoomOrderItem" ADD CONSTRAINT "fk_ProcedureRoomOrderItem_ProcedureRoom" FOREIGN KEY ("ProcedureRoom_Code") REFERENCES public."ProcedureRoom"("Code");
ALTER TABLE public."ProcedureRoomOrderItem" ADD CONSTRAINT "fk_ProcedureRoomOrderItem_ProcedureRoomOrder" FOREIGN KEY ("ProcedureRoomOrder_Id") REFERENCES public."ProcedureRoomOrder"("Id");

ALTER TABLE public."RadiologyMcuOrder" ADD CONSTRAINT "fk_RadiologyMcuOrder_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."RadiologyMcuOrder" ADD CONSTRAINT "fk_RadiologyMcuOrder_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."RadiologyMcuOrderItem" ADD CONSTRAINT "fk_RadiologyMcuOrderItem_McuSrc" FOREIGN KEY ("McuSrc_Code") REFERENCES public."McuSrc"("Code");
ALTER TABLE public."RadiologyMcuOrderItem" ADD CONSTRAINT "fk_RadiologyMcuOrderItem_RadiologyMcuOrder" FOREIGN KEY ("RadiologyMcuOrder_Id") REFERENCES public."RadiologyMcuOrder"("Id");

ALTER TABLE public."Regency" ADD CONSTRAINT "fk_Regency_Province" FOREIGN KEY ("Province_Code") REFERENCES public."Province"("Code");

ALTER TABLE public."Registrator" ADD CONSTRAINT "fk_Registrator_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");
ALTER TABLE public."Registrator" ADD CONSTRAINT "fk_Registrator_Installation" FOREIGN KEY ("Installation_Code") REFERENCES public."Installation"("Code");

ALTER TABLE public."Rehab" ADD CONSTRAINT "fk_Rehab_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."ResponsibleDoctorHist" ADD CONSTRAINT "fk_ResponsibleDoctorHist_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");

ALTER TABLE public."Resume" ADD CONSTRAINT "fk_Resume_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");

ALTER TABLE public."RoomOrder" ADD CONSTRAINT "RoomOrder_Encounter_Id_fkey" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");
ALTER TABLE public."RoomOrder" ADD CONSTRAINT "RoomOrder_Infra_Code_fkey" FOREIGN KEY ("Infra_Code") REFERENCES public."Infra"("Code");
ALTER TABLE public."RoomOrder" ADD CONSTRAINT "fk_RoomOrder_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");
ALTER TABLE public."RoomOrder" ADD CONSTRAINT "fk_RoomOrder_Infra" FOREIGN KEY ("Infra_Code") REFERENCES public."Infra"("Code");

ALTER TABLE public."Sbar" ADD CONSTRAINT "fk_Sbar_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");
ALTER TABLE public."Sbar" ADD CONSTRAINT "fk_Sbar_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."Screener" ADD CONSTRAINT "fk_Screener_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");

ALTER TABLE public."Screening" ADD CONSTRAINT "fk_Screening_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");

ALTER TABLE public."SharedTreatment" ADD CONSTRAINT "fk_SharedTreatment_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."SharedTreatment" ADD CONSTRAINT "fk_SharedTreatment_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."Soapi" ADD CONSTRAINT "fk_Soapi_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");
ALTER TABLE public."Soapi" ADD CONSTRAINT "fk_Soapi_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."Specialist" ADD CONSTRAINT "fk_Specialist_Installation" FOREIGN KEY ("Installation_Code") REFERENCES public."Installation"("Code");

ALTER TABLE public."SpecialistIntern" ADD CONSTRAINT "fk_SpecialistIntern_Person" FOREIGN KEY ("Person_Id") REFERENCES public."Person"("Id");
ALTER TABLE public."SpecialistIntern" ADD CONSTRAINT "fk_SpecialistIntern_Specialist" FOREIGN KEY ("Specialist_Code") REFERENCES public."Specialist"("Code");
ALTER TABLE public."SpecialistIntern" ADD CONSTRAINT "fk_SpecialistIntern_Subspecialist" FOREIGN KEY ("Subspecialist_Code") REFERENCES public."Subspecialist"("Code");
ALTER TABLE public."SpecialistIntern" ADD CONSTRAINT "fk_SpecialistIntern_User" FOREIGN KEY ("User_Id") REFERENCES public."User"("Id");

ALTER TABLE public."SpecialistPosition" ADD CONSTRAINT "fk_SpecialistPosition_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");
ALTER TABLE public."SpecialistPosition" ADD CONSTRAINT "fk_SpecialistPosition_Specialist" FOREIGN KEY ("Specialist_Code") REFERENCES public."Specialist"("Code");

ALTER TABLE public."Subspecialist" ADD CONSTRAINT "fk_Subspecialist_Specialist" FOREIGN KEY ("Specialist_Code") REFERENCES public."Specialist"("Code");

ALTER TABLE public."SubspecialistPosition" ADD CONSTRAINT "fk_SubspecialistPosition_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");
ALTER TABLE public."SubspecialistPosition" ADD CONSTRAINT "fk_SubspecialistPosition_Subspecialist" FOREIGN KEY ("Subspecialist_Code") REFERENCES public."Subspecialist"("Code");

ALTER TABLE public."Therapist" ADD CONSTRAINT "fk_Therapist_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");

ALTER TABLE public."TherapyProgramForm" ADD CONSTRAINT "fk_TherapyProgramForm_CreatedBy_Employee" FOREIGN KEY ("CreatedBy_Employee_Id") REFERENCES public."Employee"("Id");
ALTER TABLE public."TherapyProgramForm" ADD CONSTRAINT "fk_TherapyProgramForm_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."TherapyProtocol" ADD CONSTRAINT "fk_TherapyProtocol_Doctor" FOREIGN KEY ("Doctor_Code") REFERENCES public."Doctor"("Code");
ALTER TABLE public."TherapyProtocol" ADD CONSTRAINT "fk_TherapyProtocol_Encounter" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."Unit" ADD CONSTRAINT "fk_Unit_Parent" FOREIGN KEY ("Parent_Id") REFERENCES public."Unit"("Id");

ALTER TABLE public."UnitPosition" ADD CONSTRAINT "fk_UnitPosition_Employee" FOREIGN KEY ("Employee_Id") REFERENCES public."Employee"("Id");
ALTER TABLE public."UnitPosition" ADD CONSTRAINT "fk_UnitPosition_Unit" FOREIGN KEY ("Unit_Code") REFERENCES public."Unit"("Code");

ALTER TABLE public."UserFes" ADD CONSTRAINT "fk_UserFes_AuthPartner" FOREIGN KEY ("AuthPartner_Code") REFERENCES public."AuthPartner"("Code");
ALTER TABLE public."UserFes" ADD CONSTRAINT "fk_UserFes_User" FOREIGN KEY ("User_Name") REFERENCES public."User"("Name");

ALTER TABLE public."VaccineData" ADD CONSTRAINT "fk_Encounter_Vaccines" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."VclaimMember" ADD CONSTRAINT "fk_Person_VclaimMember" FOREIGN KEY ("Person_Id") REFERENCES public."Person"("Id");

ALTER TABLE public."VclaimReference" ADD CONSTRAINT "fk_Encounter_VclaimReference" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."VclaimSep" ADD CONSTRAINT "fk_Encounter_VclaimSep" FOREIGN KEY ("Encounter_Id") REFERENCES public."Encounter"("Id");

ALTER TABLE public."VclaimSepControlLetter" ADD CONSTRAINT "fk_VclaimSep_ControlLetters" FOREIGN KEY ("VclaimSep_Number") REFERENCES public."VclaimSep"("Number");

ALTER TABLE public."VclaimSepPrint" ADD CONSTRAINT "fk_VclaimSep_Prints" FOREIGN KEY ("VclaimSep_Number") REFERENCES public."VclaimSep"("Number");

ALTER TABLE public."VehicleHist" ADD CONSTRAINT "fk_VehicleHist_Vehicle" FOREIGN KEY ("Vehicle_Id") REFERENCES public."Vehicle"("Id");

ALTER TABLE public."Village" ADD CONSTRAINT "fk_Village_District" FOREIGN KEY ("District_Code") REFERENCES public."District"("Code");