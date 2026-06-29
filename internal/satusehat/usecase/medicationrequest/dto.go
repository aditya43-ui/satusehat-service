package medicationrequest

import "time"

type MedicationRequestRequest struct {
	MedicationRequestID    string     `json:"medicationrequest_id,omitempty"` // Contoh: No Resep / ID Resep
	PrescriptionItemID     string     `json:"prescription_item_id,omitempty"`
	OrganizationID         string     `json:"organization_id,omitempty"`
	Category               string     `json:"category,omitempty" binding:"omitempty,oneof=outpatient inpatient community discharge"`
	Priority               string     `json:"priority,omitempty" binding:"omitempty,oneof=routine urgent asap stat"`
	PatientID              string     `json:"patient_id" binding:"required"`
	PatientName            string     `json:"patient_name,omitempty"`
	EncounterID            string     `json:"encounter_id" binding:"required"`
	PractitionerID         string     `json:"practitioner_id" binding:"required"`
	PractitionerName       string     `json:"practitioner_name,omitempty"`
	MedicationID           string     `json:"medication_id" binding:"required"` // Reference to Medication resource
	MedicationDisplay      string     `json:"medication_display,omitempty"`
	ReasonCode             string     `json:"reason_code,omitempty"`
	ReasonDisplay          string     `json:"reason_display,omitempty"`
	CourseOfTherapyCode    string     `json:"course_of_therapy_code,omitempty"`
	CourseOfTherapyDisplay string     `json:"course_of_therapy_display,omitempty"`
	Status                 string     `json:"status" binding:"required,oneof=active on-hold cancelled completed entered-in-error stopped draft unknown"`
	Intent                 string     `json:"intent" binding:"required,oneof=proposal plan order original-order reflex-order filler-order instance-order option"`
	AuthoredOn             time.Time  `json:"authored_on" binding:"required"`
	DosageText             string     `json:"dosage_text,omitempty"`
	AdditionalInstruction  string     `json:"additional_instruction,omitempty"`
	PatientInstr           string     `json:"patient_instruction"`
	TimingFrequency        int        `json:"timing_frequency,omitempty"`
	TimingPeriod           int        `json:"timing_period,omitempty"`
	TimingPeriodUnit       string     `json:"timing_period_unit,omitempty"` // d (days), h (hours), etc
	RouteCode              string     `json:"route_code,omitempty"`         // e.g. O (Oral)
	RouteDisplay           string     `json:"route_display,omitempty"`
	DoseQuantityValue      float64    `json:"dose_quantity_value,omitempty"`
	DoseQuantityUnit       string     `json:"dose_quantity_unit,omitempty"` // TAB, CAP, etc
	DispenseInterval       int        `json:"dispense_interval,omitempty"`
	DispenseValue          float64    `json:"dispense_value,omitempty"`
	DispenseUnit           string     `json:"dispense_unit,omitempty"`
	SupplyDuration         int        `json:"supply_duration,omitempty"` // in days
	ValidityPeriodStart    *time.Time `json:"validity_period_start,omitempty"`
	ValidityPeriodEnd      *time.Time `json:"validity_period_end,omitempty"`
}

type MedicationRequestPatchRequest []map[string]interface{}
