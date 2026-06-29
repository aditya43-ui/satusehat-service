package questionnaireresponse

import (
	"time"
)

type QuestionnaireResponseRequest struct {
	OrganizationID          string `json:"organization_id,omitempty"`
	QuestionnaireResponseID string `json:"questionnaire_response_id,omitempty"`

	QuestionnaireURL string `json:"questionnaire_url" binding:"required"` // e.g. "https://fhir.kemkes.go.id/Questionnaire/Q0007"
	Status           string `json:"status" binding:"required,oneof=in-progress completed amended entered-in-error stopped"`

	PatientID   string `json:"patient_id" binding:"required"`
	PatientName string `json:"patient_name,omitempty"`

	EncounterID      string `json:"encounter_id,omitempty"`
	EncounterDisplay string `json:"encounter_display,omitempty"`

	AuthorID   string `json:"author_id,omitempty"`
	AuthorName string `json:"author_name,omitempty"`
	AuthorType string `json:"author_type,omitempty"` // default: "Practitioner"

	SourceID   string `json:"source_id,omitempty"`
	SourceName string `json:"source_name,omitempty"`
	SourceType string `json:"source_type,omitempty"` // default: "Patient"

	Authored *time.Time `json:"authored" binding:"required"`

	Items []ItemDTO `json:"items" binding:"required,min=1"`
}

// ItemDTO represents one questionnaire answer in flat form.
// Provide exactly one of the answer_* fields per item.
type ItemDTO struct {
	LinkID string `json:"link_id" binding:"required"`
	Text   string `json:"text,omitempty"`

	AnswerBoolean       *bool      `json:"answer_boolean,omitempty"`
	AnswerString        string     `json:"answer_string,omitempty"`
	AnswerInteger       *int       `json:"answer_integer,omitempty"`
	AnswerDecimal       *float64   `json:"answer_decimal,omitempty"`
	AnswerDate          *time.Time `json:"answer_date,omitempty"`
	AnswerDateTime      *time.Time `json:"answer_datetime,omitempty"`
	AnswerQuantityValue *float64   `json:"answer_quantity_value,omitempty"`
	AnswerQuantityUnit  string     `json:"answer_quantity_unit,omitempty"`
	AnswerQuantitySystem string    `json:"answer_quantity_system,omitempty"`
	AnswerQuantityCode  string     `json:"answer_quantity_code,omitempty"`
	AnswerCodingSystem  string     `json:"answer_coding_system,omitempty"`
	AnswerCodingCode    string     `json:"answer_coding_code,omitempty"`
	AnswerCodingDisplay string     `json:"answer_coding_display,omitempty"`
}

type QuestionnaireResponsePatchRequest []map[string]interface{}
