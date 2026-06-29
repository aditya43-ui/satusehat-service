package procedure

import (
	"time"
)

type ProcedureRequest struct {
	OrganizationID string `json:"organization_id,omitempty"`
	ProcedureID    string `json:"procedure_id,omitempty"`

	Status string `json:"status" binding:"required,oneof=preparation in-progress not-done on-hold stopped completed entered-in-error unknown"`

	CategoryCode    string `json:"category_code,omitempty"`
	CategoryDisplay string `json:"category_display,omitempty"`
	CategorySystem  string `json:"category_system,omitempty"`

	CodeSystem  string `json:"code_system,omitempty"`
	Code        string `json:"code" binding:"required"`
	CodeDisplay string `json:"code_display,omitempty"`

	PatientID   string `json:"patient_id" binding:"required"`
	PatientName string `json:"patient_name,omitempty"`

	EncounterID      string `json:"encounter_id,omitempty"`
	EncounterDisplay string `json:"encounter_display,omitempty"`

	PerformedDateTime *time.Time `json:"performed_date_time,omitempty"`
	PerformedStart    *time.Time `json:"performed_start,omitempty"`
	PerformedEnd      *time.Time `json:"performed_end,omitempty"`

	PerformerID   string `json:"performer_id,omitempty"`
	PerformerName string `json:"performer_name,omitempty"`

	ReasonCode    string `json:"reason_code,omitempty"`
	ReasonDisplay string `json:"reason_display,omitempty"`
	ReasonSystem  string `json:"reason_system,omitempty"`

	BodySiteCode    string `json:"body_site_code,omitempty"`
	BodySiteDisplay string `json:"body_site_display,omitempty"`
	BodySiteSystem  string `json:"body_site_system,omitempty"`

	Note string `json:"note,omitempty"`
}

type ProcedurePatchRequest []map[string]interface{}
