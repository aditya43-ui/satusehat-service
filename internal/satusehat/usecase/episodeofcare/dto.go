package episodeofcare

import (
	"time"
)

type EpisodeOfCareRequest struct {
	OrganizationID  string `json:"organization_id,omitempty"`
	EpisodeOfCareID string `json:"episode_of_care_id,omitempty"`

	Status string `json:"status" binding:"required,oneof=planned waitlist active onhold finished cancelled entered-in-error"`

	TypeSystem  string `json:"type_system,omitempty"`
	TypeCode    string `json:"type_code,omitempty"`
	TypeDisplay string `json:"type_display,omitempty"`

	PatientID   string `json:"patient_id" binding:"required"`
	PatientName string `json:"patient_name,omitempty"`

	ManagingOrganizationID string `json:"managing_organization_id,omitempty"`

	PeriodStart *time.Time `json:"period_start,omitempty"`
	PeriodEnd   *time.Time `json:"period_end,omitempty"`

	CareManagerID   string `json:"care_manager_id,omitempty"`
	CareManagerName string `json:"care_manager_name,omitempty"`

	DiagnosisConditionID string `json:"diagnosis_condition_id,omitempty"`
	DiagnosisRoleSystem  string `json:"diagnosis_role_system,omitempty"`
	DiagnosisRoleCode    string `json:"diagnosis_role_code,omitempty"`
	DiagnosisRoleDisplay string `json:"diagnosis_role_display,omitempty"`
	DiagnosisRank        int    `json:"diagnosis_rank,omitempty"`
}

type EpisodeOfCarePatchRequest []map[string]interface{}
