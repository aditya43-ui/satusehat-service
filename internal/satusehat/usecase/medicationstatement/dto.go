package medicationstatement

import (
	"time"
)

type MedicationStatementRequest struct {
	OrganizationID string `json:"organization_id,omitempty"`
	StatementID    string `json:"statement_id,omitempty"`

	Status string `json:"status" binding:"required,oneof=active completed entered-in-error intended stopped on-hold unknown not-taken"`

	CategorySystem  string `json:"category_system,omitempty"`
	CategoryCode    string `json:"category_code,omitempty"`
	CategoryDisplay string `json:"category_display,omitempty"`

	MedicationCodeSystem string `json:"medication_code_system,omitempty"`
	MedicationCode       string `json:"medication_code" binding:"required"`
	MedicationDisplay    string `json:"medication_display,omitempty"`

	PatientID   string `json:"patient_id" binding:"required"`
	PatientName string `json:"patient_name,omitempty"`

	EncounterID string `json:"encounter_id,omitempty"`

	EffectiveDateTime *time.Time `json:"effective_date_time,omitempty"`
	DateAsserted      *time.Time `json:"date_asserted,omitempty"`

	InformationSourceID   string `json:"information_source_id,omitempty"`
	InformationSourceName string `json:"information_source_name,omitempty"`
	InformationSourceType string `json:"information_source_type,omitempty"` // default: Patient

	DosageText               string  `json:"dosage_text,omitempty"`
	DosagePatientInstruction string  `json:"dosage_patient_instruction,omitempty"`
	DosageRouteSystem        string  `json:"dosage_route_system,omitempty"`
	DosageRouteCode          string  `json:"dosage_route_code,omitempty"`
	DosageRouteDisplay       string  `json:"dosage_route_display,omitempty"`
	DoseQuantityValue        float64 `json:"dose_quantity_value,omitempty"`
	DoseQuantityUnit         string  `json:"dose_quantity_unit,omitempty"`
}

type MedicationStatementPatchRequest []map[string]interface{}
