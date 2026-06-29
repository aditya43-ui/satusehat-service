package allergyintolerance

import (
	"time"
)

type AllergyIntoleranceRequest struct {
	OrganizationID string `json:"organization_id,omitempty"`
	AllergyID      string `json:"allergy_id,omitempty"`

	ClinicalStatus     string `json:"clinical_status" binding:"required,oneof=active inactive resolved"`
	VerificationStatus string `json:"verification_status" binding:"required,oneof=unconfirmed presumed confirmed refuted entered-in-error"`

	Category string `json:"category,omitempty"`

	CodeSystem  string `json:"code_system,omitempty"`
	Code        string `json:"code" binding:"required"`
	CodeDisplay string `json:"code_display,omitempty"`

	PatientID   string `json:"patient_id" binding:"required"`
	PatientName string `json:"patient_name,omitempty"`

	EncounterID      string `json:"encounter_id" binding:"required"`
	EncounterDisplay string `json:"encounter_display,omitempty"`

	RecordedDate    *time.Time `json:"recorded_date" binding:"required"`
	RecorderID      string     `json:"recorder_id,omitempty"`
	RecorderDisplay string     `json:"recorder_display,omitempty"`
}

type AllergyIntolerancePatchRequest []map[string]interface{}
