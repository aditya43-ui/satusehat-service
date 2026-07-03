package purificationdecision

import "service/internal/interfaces/satusehat"

const termSys = "http://terminology.kemkes.go.id"

// MapRequestToFHIR — PurificationDecisionRequest → FHIR PurificationDecision.
func MapRequestToFHIR(req PurificationDecisionRequest) satusehat.FHIRPayload {
	p := satusehat.NewFHIRPayload("PurificationDecision")

	if req.DecisionNumber != "" && req.OrganizationID != "" {
		p.Set("identifier", []map[string]interface{}{{
			"system": "http://sys-ids.kemkes.go.id/purificationdecision/" + req.OrganizationID,
			"value":  req.DecisionNumber,
		}})
	}
	if req.Created != nil {
		p.Set("created", req.Created.Format("2006-01-02T15:04:05-07:00"))
	}
	statusCoding := map[string]interface{}{"system": termSys, "code": req.StatusCode}
	if req.StatusDisplay != "" {
		statusCoding["display"] = req.StatusDisplay
	}
	p.Set("status", map[string]interface{}{"coding": []map[string]interface{}{statusCoding}})

	if req.InsurerID != "" {
		p.Set("insurer", map[string]interface{}{"reference": "Organization/" + req.InsurerID})
	}
	if req.OrganizationID != "" {
		p.Set("provider", map[string]interface{}{"reference": "Organization/" + req.OrganizationID})
	}
	p.Set("claimResponse", map[string]interface{}{"reference": "ClaimResponse/" + req.ClaimResponseID})

	if len(req.Extension) > 0 {
		p.Set("extension", req.Extension)
	}
	return p
}
