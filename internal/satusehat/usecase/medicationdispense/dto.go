package medicationdispense

import "time"

type MedicationDispenseRequest struct {
	OrganizationID      string    `json:"organization_id,omitempty"`
	PatientID           string    `json:"patient_id" binding:"required"`
	PatientName         string    `json:"patient_name,omitempty"`
	EncounterID         string    `json:"encounter_id" binding:"required"`
	MedicationID        string    `json:"medication_id" binding:"required"`
	MedicationDisplay   string    `json:"medication_display,omitempty"`
	PractitionerID      string    `json:"practitioner_id" binding:"required"`
	PractitionerName    string    `json:"practitioner_name,omitempty"`
	LocationID          string    `json:"location_id" binding:"required"`
	LocationName        string    `json:"location_name,omitempty"`
	PrescriptionID      string    `json:"prescription_id" binding:"required"` // Digunakan untuk Prescription Identifier
	PrescriptionItemID  string    `json:"prescription_item_id,omitempty"`     // Digunakan untuk Identifier item
	Status              string    `json:"status" binding:"required,oneof=preparation in-progress cancelled on-hold completed entered-in-error stopped declined unknown"`
	Category            string    `json:"category,omitempty" binding:"omitempty,oneof=outpatient inpatient community discharge"`
	PreparedDate        time.Time `json:"prepared_date" binding:"required"`
	HandedOverDate      time.Time `json:"handed_over_date" binding:"required"`
	QuantityValue       float64   `json:"quantity_value" binding:"required"`
	QuantityUnit        string    `json:"quantity_unit" binding:"required"` // e.g. TAB
	DaysSupplyValue     int       `json:"days_supply_value,omitempty"`
	DosageText          string    `json:"dosage_text,omitempty"`
	TimingFrequency     int       `json:"timing_frequency,omitempty"`
	TimingPeriod        int       `json:"timing_period,omitempty"`
	TimingPeriodUnit    string    `json:"timing_period_unit,omitempty"` // e.g. d, h
	DoseQuantityValue   float64   `json:"dose_quantity_value,omitempty"`
	DoseQuantityUnit    string    `json:"dose_quantity_unit,omitempty"`             // e.g. TAB
	MedicationRequestID string    `json:"medication_request_id" binding:"required"` //Digunakan untuk authorizingPrescriptionID
}

type MedicationDispensePatchRequest []map[string]interface{}
