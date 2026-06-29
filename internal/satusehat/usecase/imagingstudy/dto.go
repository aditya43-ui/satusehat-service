package imagingstudy

import "time"

type ImagingStudyRequest struct {
	OrganizationID    string    `json:"organization_id,omitempty"`
	AccessionNumber   string    `json:"accession_number,omitempty"`
	ServiceRequestID  string    `json:"service_request_id,omitempty"`
	PatientID         string    `json:"patient_id,omitempty"`
	PatientName       string    `json:"patient_name,omitempty"`
	EncounterID       string    `json:"encounter_id,omitempty"`
	PractitionerID    string    `json:"practitioner_id,omitempty"`
	PractitionerName  string    `json:"practitioner_name,omitempty"`
	Status            string    `json:"status,omitempty"` // registered, available, cancelled, entered-in-error, unknown
	Started           time.Time `json:"started,omitempty"`
	NumberOfSeries    *int      `json:"number_of_series,omitempty"`
	NumberOfInstances *int      `json:"number_of_instances,omitempty"`
	ProcedureCode     string    `json:"procedure_code,omitempty"` // SNOMED CT
	ProcedureDisplay  string    `json:"procedure_display,omitempty"`
	Description       string    `json:"description,omitempty"`
}

type ImagingStudyPatchRequest []map[string]interface{}
