package episodeofcare

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req EpisodeOfCareRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("EpisodeOfCare")

	orgID := req.OrganizationID

	if orgID != "" && req.EpisodeOfCareID != "" {
		payload.Set("identifier", []map[string]interface{}{
			{
				"system": "http://sys-ids.kemkes.go.id/episode-of-care/" + orgID,
				"value":  req.EpisodeOfCareID,
			},
		})
	}

	if req.Status != "" {
		payload.Set("status", req.Status)
	}

	if req.TypeCode != "" {
		sys := req.TypeSystem
		if sys == "" {
			sys = "http://terminology.kemkes.go.id/CodeSystem/episodeofcare-type"
		}
		coding := map[string]interface{}{"system": sys, "code": req.TypeCode}
		if req.TypeDisplay != "" {
			coding["display"] = req.TypeDisplay
		}
		payload.Set("type", []map[string]interface{}{
			{"coding": []map[string]interface{}{coding}},
		})
	}

	if req.DiagnosisConditionID != "" {
		diag := map[string]interface{}{
			"condition": map[string]interface{}{"reference": "Condition/" + req.DiagnosisConditionID},
		}
		if req.DiagnosisRoleCode != "" {
			sys := req.DiagnosisRoleSystem
			if sys == "" {
				sys = "http://terminology.hl7.org/CodeSystem/diagnosis-role"
			}
			coding := map[string]interface{}{"system": sys, "code": req.DiagnosisRoleCode}
			if req.DiagnosisRoleDisplay != "" {
				coding["display"] = req.DiagnosisRoleDisplay
			}
			diag["role"] = map[string]interface{}{"coding": []map[string]interface{}{coding}}
		}
		if req.DiagnosisRank > 0 {
			diag["rank"] = req.DiagnosisRank
		}
		payload.Set("diagnosis", []map[string]interface{}{diag})
	}

	if req.PatientID != "" {
		subject := map[string]interface{}{"reference": "Patient/" + req.PatientID}
		if req.PatientName != "" {
			subject["display"] = req.PatientName
		}
		payload.Set("patient", subject)
	}

	manageOrg := orgID
	if req.ManagingOrganizationID != "" {
		manageOrg = req.ManagingOrganizationID
	}
	if manageOrg != "" {
		payload.Set("managingOrganization", map[string]interface{}{
			"reference": "Organization/" + manageOrg,
		})
	}

	if req.PeriodStart != nil && !req.PeriodStart.IsZero() {
		period := map[string]interface{}{"start": req.PeriodStart.Format(time.RFC3339)}
		if req.PeriodEnd != nil && !req.PeriodEnd.IsZero() {
			period["end"] = req.PeriodEnd.Format(time.RFC3339)
		}
		payload.Set("period", period)
	}

	if req.CareManagerID != "" {
		cm := map[string]interface{}{"reference": "Practitioner/" + req.CareManagerID}
		if req.CareManagerName != "" {
			cm["display"] = req.CareManagerName
		}
		payload.Set("careManager", cm)
	}

	return payload
}
