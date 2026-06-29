package careplan

import (
	"time"
)

type CarePlanRequest struct {
	OrganizationID string `json:"organization_id,omitempty"`
	CarePlanID     string `json:"care_plan_id,omitempty"`

	Status string `json:"status" binding:"required,oneof=draft active on-hold revoked completed entered-in-error unknown"`
	Intent string `json:"intent" binding:"required,oneof=proposal plan order option"`

	CategoryCode    string `json:"category_code,omitempty"`
	CategoryDisplay string `json:"category_display,omitempty"`
	CategorySystem  string `json:"category_system,omitempty"`

	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`

	PatientID      string `json:"patient_id" binding:"required"`
	PatientDisplay string `json:"patient_display,omitempty"`

	EncounterID      string `json:"encounter_id" binding:"required"`
	EncounterDisplay string `json:"encounter_display,omitempty"`

	CreatedDate *time.Time `json:"created_date,omitempty"`

	AuthorID      string `json:"author_id,omitempty"`
	AuthorDisplay string `json:"author_display,omitempty"`

	GoalIDs []string `json:"goal_ids,omitempty"`
}

type CarePlanPatchRequest []map[string]interface{}
