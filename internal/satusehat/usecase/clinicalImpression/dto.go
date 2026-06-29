package clinicalimpression

import (
	"time"
)

type ClinicalImpressionRequest struct {
	OrganizationID       string `json:"organization_id,omitempty"`
	ClinicalImpressionID string `json:"clinical_impression_id,omitempty"`

	Status string `json:"status" binding:"required,oneof=in-progress completed entered-in-error"`

	CodeSystem  string `json:"code_system,omitempty"`
	Code        string `json:"code,omitempty"`
	CodeDisplay string `json:"code_display,omitempty"`

	Description string `json:"description,omitempty"`

	PatientID      string `json:"patient_id" binding:"required"`
	PatientDisplay string `json:"patient_display,omitempty"`

	EncounterID      string `json:"encounter_id" binding:"required"`
	EncounterDisplay string `json:"encounter_display,omitempty"`

	EffectiveDateTime *time.Time `json:"effective_datetime,omitempty"`
	Date              *time.Time `json:"date,omitempty"`

	AssessorID      string `json:"assessor_id,omitempty"`
	AssessorDisplay string `json:"assessor_display,omitempty"`

	ProblemConditionIDs []string `json:"problem_condition_ids,omitempty"`

	Summary string `json:"summary,omitempty"`

	FindingSystem  string `json:"finding_system,omitempty"`
	FindingCode    string `json:"finding_code,omitempty"`
	FindingDisplay string `json:"finding_display,omitempty"`

	PrognosisSystem  string `json:"prognosis_system,omitempty"`
	PrognosisCode    string `json:"prognosis_code,omitempty"`
	PrognosisDisplay string `json:"prognosis_display,omitempty"`
}

type ClinicalImpressionPatchRequest []map[string]interface{}
