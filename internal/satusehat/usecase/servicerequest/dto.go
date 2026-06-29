package servicerequest

import "time"

type ServiceRequestRequest struct {
	OrganizationID   string    `json:"organization_id,omitempty"`
	ServiceRequestID string    `json:"service_request_id,omitempty"`
	PatientID        string    `json:"patient_id,omitempty"`
	PatientName   string    `json:"patient_name,omitempty"`
	EncounterID   string    `json:"encounter_id,omitempty"`
	RequesterID   string    `json:"requester_id,omitempty"`
	PerformerID   string    `json:"performer_id,omitempty"`
	RequesterName string    `json:"requester_name,omitempty"`
	PerformerName string    `json:"performer_name,omitempty"`
	Status        string    `json:"status,omitempty" binding:"omitempty,oneof=draft active on-hold revoked completed entered-in-error unknown"`
	Intent        string    `json:"intent,omitempty" binding:"omitempty,oneof=proposal plan directive order original-order reflex-order filler-order instance-order option"`
	Code          string    `json:"code,omitempty"`
	Display       string    `json:"display,omitempty"`
	AuthoredOn    time.Time `json:"authored_on,omitempty"`
}

type ServiceRequestPatchRequest []map[string]interface{}
