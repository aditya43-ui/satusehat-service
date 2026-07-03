package claimresponse

import "time"

// ClaimResponseRequest — payload FHIR ClaimResponse (respon klaim BPJS-K: Purifikasi / Verifikasi/BAHV).
// Ref: docs Postman Modul Klaim (BPJS) — POST /ClaimResponse. Di-POST oleh sistem untuk
// merepresentasikan hasil purifikasi/verifikasi yang diterima dari BPJS.
type ClaimResponseRequest struct {
	OrganizationID string `json:"organization_id,omitempty"` // Org faskes (requestor)
	InsurerID      string `json:"insurer_id,omitempty"`      // Org BPJS

	ClaimNumber string `json:"claim_number,omitempty"` // identifier claim-number
	BatchNumber string `json:"batch_number,omitempty"` // identifier claim-batch-number

	Status  string `json:"status,omitempty"`   // default: active
	SubType string `json:"sub_type,omitempty"` // purifikasi | verifikasi (terminology.kemkes.go.id)
	Use     string `json:"use,omitempty"`      // default: claim
	Outcome string `json:"outcome,omitempty"`  // default: complete

	PatientID   string     `json:"patient_id" binding:"required"`
	ClaimID     string     `json:"claim_id,omitempty"` // request → Claim/{id}
	Disposition string     `json:"disposition,omitempty"`
	Created     *time.Time `json:"created,omitempty"`

	// Passthrough FHIR (dirakit idrg): identifier tambahan, adjudication (hasil purifikasi/verifikasi).
	Identifier   []map[string]interface{} `json:"identifier,omitempty"`
	Adjudication []map[string]interface{} `json:"adjudication,omitempty"`
	Extension    []map[string]interface{} `json:"extension,omitempty"`
}
