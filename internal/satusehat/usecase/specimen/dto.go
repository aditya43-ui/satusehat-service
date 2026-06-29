package specimen

import (
	"time"
)

type SpecimenRequest struct {
	OrganizationID string `json:"organization_id,omitempty"`
	SpecimenID     string `json:"specimen_id,omitempty"`

	Status string `json:"status" binding:"required,oneof=available unavailable unsatisfactory entered-in-error"`

	TypeSystem  string `json:"type_system,omitempty"`
	TypeCode    string `json:"type_code" binding:"required"`
	TypeDisplay string `json:"type_display,omitempty"`

	PatientID   string `json:"patient_id" binding:"required"`
	PatientName string `json:"patient_name,omitempty"`

	ReceivedDateTime *time.Time `json:"received_date_time,omitempty"`

	CollectedDateTime *time.Time `json:"collected_date_time,omitempty"`
	CollectorID       string     `json:"collector_id,omitempty"`
	CollectorName     string     `json:"collector_name,omitempty"`

	CollectionQuantityValue  *float64 `json:"collection_quantity_value,omitempty"`
	CollectionQuantityUnit   string   `json:"collection_quantity_unit,omitempty"`
	CollectionQuantitySystem string   `json:"collection_quantity_system,omitempty"`

	CollectionMethodSystem  string `json:"collection_method_system,omitempty"`
	CollectionMethodCode    string `json:"collection_method_code,omitempty"`
	CollectionMethodDisplay string `json:"collection_method_display,omitempty"`

	BodySiteSystem  string `json:"body_site_system,omitempty"`
	BodySiteCode    string `json:"body_site_code,omitempty"`
	BodySiteDisplay string `json:"body_site_display,omitempty"`

	FastingStatusSystem  string `json:"fasting_status_system,omitempty"`
	FastingStatusCode    string `json:"fasting_status_code,omitempty"`
	FastingStatusDisplay string `json:"fasting_status_display,omitempty"`

	ProcessingProcedureSystem  string     `json:"processing_procedure_system,omitempty"`
	ProcessingProcedureCode    string     `json:"processing_procedure_code,omitempty"`
	ProcessingProcedureDisplay string     `json:"processing_procedure_display,omitempty"`
	ProcessingTimeDateTime     *time.Time `json:"processing_time_datetime,omitempty"`

	Conditions []string `json:"conditions,omitempty"`

	RequestServiceRequestIDs []string `json:"request_service_request_ids,omitempty"`
}

type SpecimenPatchRequest []map[string]interface{}
