package condition

import (
	"time"
)

type ConditionRequest struct {
	OrganizationID string `json:"organization_id,omitempty"`
	ConditionID    string `json:"condition_id,omitempty"`

	PatientID   string `json:"patient_id" binding:"required"`
	PatientName string `json:"patient_name,omitempty"`

	EncounterID string `json:"encounter_id" binding:"required"`

	ClinicalStatus string `json:"clinical_status" binding:"required,oneof=active recurrence relapse inactive remission resolved"`

	CategoryCode    string `json:"category_code,omitempty"`
	CategoryDisplay string `json:"category_display,omitempty"`
	CategorySystem  string `json:"category_system,omitempty"`

	CodeSystem  string `json:"code_system,omitempty"`
	Code        string `json:"code" binding:"required"`
	CodeDisplay string `json:"code_display,omitempty"`

	OnsetDateTime *time.Time `json:"onset_date_time,omitempty"`
	RecordedDate  *time.Time `json:"recorded_date,omitempty"`
}

type ConditionPatchRequest []map[string]interface{}
