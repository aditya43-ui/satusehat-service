package composition

import (
	"time"
)

type CompositionRequest struct {
	OrganizationID string `json:"organization_id,omitempty"`
	CompositionID  string `json:"composition_id,omitempty"`

	Status string `json:"status" binding:"required,oneof=preliminary final amended entered-in-error"`

	TypeSystem  string `json:"type_system,omitempty"`
	TypeCode    string `json:"type_code" binding:"required"`
	TypeDisplay string `json:"type_display,omitempty"`

	CategorySystem  string `json:"category_system,omitempty"`
	CategoryCode    string `json:"category_code,omitempty"`
	CategoryDisplay string `json:"category_display,omitempty"`

	PatientID      string `json:"patient_id" binding:"required"`
	PatientDisplay string `json:"patient_display,omitempty"`

	EncounterID      string `json:"encounter_id" binding:"required"`
	EncounterDisplay string `json:"encounter_display,omitempty"`

	Date *time.Time `json:"date" binding:"required"`

	AuthorID      string `json:"author_id" binding:"required"`
	AuthorDisplay string `json:"author_display,omitempty"`

	Title string `json:"title" binding:"required"`

	SectionTitle   string `json:"section_title,omitempty"`
	SectionSystem  string `json:"section_system,omitempty"`
	SectionCode    string `json:"section_code,omitempty"`
	SectionDisplay string `json:"section_display,omitempty"`
	SectionText    string `json:"section_text,omitempty"`
}

type CompositionPatchRequest []map[string]interface{}
