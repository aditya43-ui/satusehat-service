package claimresponse

import "service/internal/interfaces/satusehat"

const termSys = "http://terminology.kemkes.go.id"

// MapRequestToFHIR — ClaimResponseRequest → FHIR ClaimResponse (R4).
func MapRequestToFHIR(req ClaimResponseRequest) satusehat.FHIRPayload {
	p := satusehat.NewFHIRPayload("ClaimResponse")

	p.Set("status", def(req.Status, "active"))

	// identifier: gabung passthrough + claim-number/batch bila ada.
	ids := append([]map[string]interface{}{}, req.Identifier...)
	if req.BatchNumber != "" && req.InsurerID != "" {
		ids = append(ids, kv("http://sys-ids.kemkes.go.id/claim-batch-number/"+req.InsurerID, req.BatchNumber))
	}
	if req.ClaimNumber != "" && req.InsurerID != "" {
		ids = append(ids, kv("http://sys-ids.kemkes.go.id/claim-number/"+req.InsurerID, req.ClaimNumber))
	}
	if len(ids) > 0 {
		p.Set("identifier", ids)
	}

	p.Set("type", coding("http://terminology.hl7.org/CodeSystem/claim-type", "institutional", "Institutional"))
	if req.SubType != "" {
		p.Set("subType", coding(termSys, req.SubType, title(req.SubType)))
	}
	p.Set("use", def(req.Use, "claim"))
	p.Set("patient", ref("Patient", req.PatientID))
	if req.Created != nil {
		p.Set("created", req.Created.Format("2006-01-02T15:04:05-07:00"))
	}
	if req.InsurerID != "" {
		p.Set("insurer", ref("Organization", req.InsurerID))
	}
	if req.OrganizationID != "" {
		p.Set("requestor", ref("Organization", req.OrganizationID))
	}
	if req.ClaimID != "" {
		p.Set("request", ref("Claim", req.ClaimID))
	}
	p.Set("outcome", def(req.Outcome, "complete"))
	if req.Disposition != "" {
		p.Set("disposition", req.Disposition)
	}
	if len(req.Adjudication) > 0 {
		p.Set("adjudication", req.Adjudication)
	}
	if len(req.Extension) > 0 {
		p.Set("extension", req.Extension)
	}
	return p
}

func def(v, d string) string {
	if v == "" {
		return d
	}
	return v
}

func title(s string) string {
	if s == "" {
		return s
	}
	return string(s[0]-32) + s[1:]
}

func ref(resource, id string) map[string]interface{} {
	return map[string]interface{}{"reference": resource + "/" + id}
}

func kv(system, value string) map[string]interface{} {
	return map[string]interface{}{"system": system, "value": value}
}

func coding(system, code, display string) map[string]interface{} {
	c := map[string]interface{}{"system": system, "code": code}
	if display != "" {
		c["display"] = display
	}
	return map[string]interface{}{"coding": []map[string]interface{}{c}}
}
