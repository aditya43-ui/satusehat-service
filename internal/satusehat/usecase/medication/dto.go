package medication

import "time"

type MedicationRequest struct {
	MedicationID   string     `json:"medication_id,omitempty"`
	OrganizationID string     `json:"organization_id,omitempty"`
	StatusCode     string     `json:"status_code" binding:"required,oneof=active inactive entered-in-error"`
	KfaCode        string     `json:"kfa_code" binding:"required"`
	KfaDisplay     string     `json:"kfa_display,omitempty"`
	FormCode       string     `json:"form_code,omitempty"`
	FormDisplay    string     `json:"form_display,omitempty"`
	ManufacturerID string     `json:"manufacturer_id,omitempty"` // Reference ke ID Organization (Pabrik/Distributor)
	BatchNumber    string     `json:"batch_number,omitempty"`
	ExpirationDate *time.Time `json:"expiration_date,omitempty"`
}

type MedicationPatchRequest []map[string]interface{}
