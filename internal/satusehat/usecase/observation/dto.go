package observation

import (
	"time"
)

type ObservationRequest struct {
	OrganizationID string `json:"organization_id,omitempty"`
	ObservationID  string `json:"observation_id,omitempty"`

	Status string `json:"status" binding:"required,oneof=registered preliminary final amended corrected cancelled entered-in-error unknown"`

	CategorySystem  string `json:"category_system,omitempty"`
	CategoryCode    string `json:"category_code" binding:"required"`
	CategoryDisplay string `json:"category_display,omitempty"`

	CodeSystem  string `json:"code_system,omitempty"`
	Code        string `json:"code" binding:"required"`
	CodeDisplay string `json:"code_display,omitempty"`

	PatientID   string `json:"patient_id" binding:"required"`
	PatientName string `json:"patient_name,omitempty"`

	EncounterID string `json:"encounter_id,omitempty"`

	EffectiveDateTime *time.Time `json:"effective_datetime,omitempty"`
	Issued            *time.Time `json:"issued,omitempty"`

	PerformerID   string `json:"performer_id,omitempty"`
	PerformerName string `json:"performer_name,omitempty"`

	SpecimenID string `json:"specimen_id,omitempty"`

	BodySiteSystem  string `json:"body_site_system,omitempty"`
	BodySiteCode    string `json:"body_site_code,omitempty"`
	BodySiteDisplay string `json:"body_site_display,omitempty"`

	ValueQuantityValue  *float64 `json:"value_quantity_value,omitempty"`
	ValueQuantityUnit   string   `json:"value_quantity_unit,omitempty"`
	ValueQuantitySystem string   `json:"value_quantity_system,omitempty"`
	ValueQuantityCode   string   `json:"value_quantity_code,omitempty"`

	ValueCodeSystem  string `json:"value_code_system,omitempty"`
	ValueCode        string `json:"value_code,omitempty"`
	ValueCodeDisplay string `json:"value_code_display,omitempty"`

	ValueString  string `json:"value_string,omitempty"`
	ValueBoolean *bool  `json:"value_boolean,omitempty"`

	InterpretationSystem  string `json:"interpretation_system,omitempty"`
	InterpretationCode    string `json:"interpretation_code,omitempty"`
	InterpretationDisplay string `json:"interpretation_display,omitempty"`

	ReferenceRangeLowValue  *float64 `json:"reference_range_low_value,omitempty"`
	ReferenceRangeHighValue *float64 `json:"reference_range_high_value,omitempty"`
	ReferenceRangeUnit      string   `json:"reference_range_unit,omitempty"`
	ReferenceRangeText      string   `json:"reference_range_text,omitempty"`
}

type ObservationPatchRequest []map[string]interface{}
