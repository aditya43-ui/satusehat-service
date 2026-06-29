package immunization

import (
	"time"
)

type ImmunizationRequest struct {
	OrganizationID string `json:"organization_id,omitempty"`
	ImmunizationID string `json:"immunization_id,omitempty"`

	Status string `json:"status" binding:"required,oneof=completed entered-in-error not-done"`

	VaccineCodeSystem  string `json:"vaccine_code_system,omitempty"`
	VaccineCode        string `json:"vaccine_code" binding:"required"`
	VaccineCodeDisplay string `json:"vaccine_display,omitempty"`

	PatientID   string `json:"patient_id" binding:"required"`
	PatientName string `json:"patient_name,omitempty"`

	EncounterID      string `json:"encounter_id" binding:"required"`
	EncounterDisplay string `json:"encounter_display,omitempty"`

	OccurrenceDateTime *time.Time `json:"occurrence_date_time,omitempty"`
	PrimarySource      *bool      `json:"primary_source,omitempty"`
	LotNumber          string     `json:"lot_number,omitempty"`

	PerformerID   string `json:"performer_id,omitempty"`
	PerformerName string `json:"performer_name,omitempty"`

	LocationID   string `json:"location_id,omitempty"`
	LocationName string `json:"location_name,omitempty"`

	DoseQuantityValue  float64 `json:"dose_quantity_value,omitempty"`
	DoseQuantityUnit   string  `json:"dose_quantity_unit,omitempty"`
	DoseQuantitySystem string  `json:"dose_quantity_system,omitempty"`
	DoseQuantityCode   string  `json:"dose_quantity_code,omitempty"`

	RouteSystem  string `json:"route_system,omitempty"`
	RouteCode    string `json:"route_code,omitempty"`
	RouteDisplay string `json:"route_display,omitempty"`
}

type ImmunizationPatchRequest []map[string]interface{}
