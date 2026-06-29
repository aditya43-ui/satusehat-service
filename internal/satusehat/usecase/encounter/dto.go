package encounter

import (
	"service/pkg/utils/custom"
)

type EncounterRequest struct {
	OrganizationID string `json:"organization_id,omitempty"`
	EncounterID    string `json:"encounter_id,omitempty"`

	Status string `json:"status" binding:"required,oneof=planned arrived triaged in-progress onleave finished cancelled entered-in-error unknown"`
	Class  string `json:"class" binding:"required,oneof=AMB EMER IMP"`

	PatientID   string `json:"patient_id" binding:"required"`
	PatientName string `json:"patient_name,omitempty"`

	PractitionerID   string `json:"practitioner_id,omitempty"`
	PractitionerName string `json:"practitioner_name,omitempty"`

	LocationID   string `json:"location_id,omitempty"`
	LocationName string `json:"location_name,omitempty"`

	PeriodStart *custom.CustomTime `json:"period_start" binding:"required"`
	PeriodEnd   *custom.CustomTime `json:"period_end,omitempty"`

	DiagnosisConditionID string `json:"diagnosis_condition_id,omitempty"`
	DiagnosisUseSystem   string `json:"diagnosis_use_system,omitempty"`
	DiagnosisUseCode     string `json:"diagnosis_use_code,omitempty"`
	DiagnosisUseDisplay  string `json:"diagnosis_use_display,omitempty"`
	DiagnosisRank        int    `json:"diagnosis_rank,omitempty"`
}

type EncounterPatchRequest []map[string]interface{}
