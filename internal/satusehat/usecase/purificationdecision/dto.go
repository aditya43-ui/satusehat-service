package purificationdecision

import "time"

// PurificationDecisionRequest — keputusan Faskes atas hasil purifikasi (Lanjut/Batal).
// Ref: docs Postman Modul Klaim (BPJS) — POST /PurificationDecision.
// status.coding (terminology.kemkes.go.id), mis. TK000049 = "Lanjut".
type PurificationDecisionRequest struct {
	OrganizationID string `json:"organization_id,omitempty"` // Org faskes (provider)
	InsurerID      string `json:"insurer_id,omitempty"`      // Org BPJS
	DecisionNumber string `json:"decision_number,omitempty"` // identifier value

	StatusCode    string `json:"status_code" binding:"required"` // mis. TK000049
	StatusDisplay string `json:"status_display,omitempty"`       // mis. Lanjut / Batal

	ClaimResponseID string     `json:"claim_response_id" binding:"required"`
	Created         *time.Time `json:"created,omitempty"`

	Extension []map[string]interface{} `json:"extension,omitempty"`
}
