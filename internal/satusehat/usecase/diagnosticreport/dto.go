package diagnosticreport

import (
	"time"
)

type DiagnosticReportRequest struct {
	OrganizationID string `json:"organization_id,omitempty"`
	DiagnosticID   string `json:"diagnostic_id,omitempty"`

	Status string `json:"status" binding:"required,oneof=registered partial preliminary final amended corrected appended cancelled entered-in-error unknown"`

	CategorySystem  string `json:"category_system,omitempty"`
	CategoryCode    string `json:"category_code,omitempty"`
	CategoryDisplay string `json:"category_display,omitempty"`

	CodeSystem  string `json:"code_system,omitempty"`
	Code        string `json:"code" binding:"required"`
	CodeDisplay string `json:"code_display,omitempty"`

	PatientID   string `json:"patient_id" binding:"required"`
	EncounterID string `json:"encounter_id" binding:"required"`

	EffectiveDateTime *time.Time `json:"effective_datetime,omitempty"`
	Issued            *time.Time `json:"issued,omitempty"`

	PerformerID   string `json:"performer_id,omitempty"`
	PerformerName string `json:"performer_name,omitempty"`

	ResultObservationIDs []string `json:"result_observation_ids,omitempty"`
	SpecimenIDs          []string `json:"specimen_ids,omitempty"`
	BasedOnIDs           []string `json:"based_on_service_request_ids,omitempty"`
	ImagingStudyIDs      []string `json:"imaging_study_ids,omitempty"`

	ConclusionCodeSystem  string `json:"conclusion_code_system,omitempty"`
	ConclusionCode        string `json:"conclusion_code,omitempty"`
	ConclusionCodeDisplay string `json:"conclusion_code_display,omitempty"`
	Conclusion            string `json:"conclusion,omitempty"`
}

type DiagnosticReportPatchRequest []map[string]interface{}
