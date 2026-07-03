package claim

import "time"

// ClaimRequest — payload buat FHIR Claim (BPJS-K) ke SATUSEHAT.
// Ref kontrak: docs/26. Use Case - Modul Klaim (BPJS).postman_collection.json (POST /Claim).
// Field inti bertipe; struktur kompleks (diagnosis/procedure/insurance/supportingInfo/item)
// diterima sebagai passthrough dari orchestrator (service-idrg) yang sudah merakit FHIR-nya.
type ClaimRequest struct {
	OrganizationID string `json:"organization_id,omitempty"`
	ClaimNumber    string `json:"claim_number,omitempty"` // identifier value (claim-number)

	Status   string `json:"status,omitempty"`    // default: active
	TypeCode string `json:"type_code,omitempty"` // default: institutional
	Use      string `json:"use,omitempty"`       // default: claim

	PatientID   string     `json:"patient_id" binding:"required"`
	PeriodStart *time.Time `json:"period_start,omitempty"`
	PeriodEnd   *time.Time `json:"period_end,omitempty"`
	Created     *time.Time `json:"created,omitempty"`

	ProviderID string `json:"provider_id,omitempty"` // Organization (faskes) — default orgID
	InsurerID  string `json:"insurer_id,omitempty"`  // Organization BPJS
	CoverageID string `json:"coverage_id,omitempty"`
	SepNumber  string `json:"sep_number,omitempty"` // insurance.identifier value (No. SEP)

	TotalValue    float64 `json:"total_value,omitempty"`
	TotalCurrency string  `json:"total_currency,omitempty"` // default: IDR

	// Passthrough FHIR (dirakit oleh idrg). Bila diisi, dipakai apa adanya.
	Diagnosis      []map[string]interface{} `json:"diagnosis,omitempty"`
	Procedure      []map[string]interface{} `json:"procedure,omitempty"`
	SupportingInfo []map[string]interface{} `json:"supporting_info,omitempty"`
	Item           []map[string]interface{} `json:"item,omitempty"`
	Extension      []map[string]interface{} `json:"extension,omitempty"`
}
