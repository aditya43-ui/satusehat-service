package claim

import "service/internal/interfaces/satusehat"

// MapRequestToFHIR — ClaimRequest → FHIR Claim (R4). Default sesuai use case BPJS-K
// (type=institutional, use=claim, currency=IDR). Struktur kompleks diambil dari passthrough.
func MapRequestToFHIR(req ClaimRequest) satusehat.FHIRPayload {
	p := satusehat.NewFHIRPayload("Claim")

	if req.OrganizationID != "" && req.ClaimNumber != "" {
		p.Set("identifier", []map[string]interface{}{{
			"system": "http://sys-ids.kemkes.go.id/claim-number/" + req.OrganizationID,
			"value":  req.ClaimNumber,
		}})
	}

	p.Set("status", def(req.Status, "active"))
	p.Set("type", map[string]interface{}{"coding": []map[string]interface{}{{
		"system":  "http://terminology.hl7.org/CodeSystem/claim-type",
		"code":    def(req.TypeCode, "institutional"),
		"display": "Institutional",
	}}})
	p.Set("use", def(req.Use, "claim"))
	p.Set("patient", ref("Patient", req.PatientID))

	if req.PeriodStart != nil || req.PeriodEnd != nil {
		bp := map[string]interface{}{}
		if req.PeriodStart != nil {
			bp["start"] = req.PeriodStart.Format("2006-01-02T15:04:05-07:00")
		}
		if req.PeriodEnd != nil {
			bp["end"] = req.PeriodEnd.Format("2006-01-02T15:04:05-07:00")
		}
		p.Set("billablePeriod", bp)
	}
	if req.Created != nil {
		p.Set("created", req.Created.Format("2006-01-02T15:04:05-07:00"))
	}

	if req.ProviderID != "" {
		p.Set("provider", ref("Organization", req.ProviderID))
	}
	if req.InsurerID != "" {
		p.Set("insurer", ref("Organization", req.InsurerID))
	}

	p.Set("priority", map[string]interface{}{"coding": []map[string]interface{}{{
		"system": "http://terminology.hl7.org/CodeSystem/processpriority", "code": "normal", "display": "Normal",
	}}})

	// insurance (Coverage + No. SEP). Passthrough tidak disediakan → rakit dari field.
	if req.CoverageID != "" {
		ins := map[string]interface{}{
			"sequence": 1,
			"focal":    true,
			"coverage": ref("Coverage", req.CoverageID),
		}
		if req.SepNumber != "" && req.InsurerID != "" {
			ins["identifier"] = map[string]interface{}{
				"system": "http://sys-ids.kemkes.go.id/claim-number/" + req.InsurerID,
				"value":  req.SepNumber,
			}
		}
		p.Set("insurance", []map[string]interface{}{ins})
	}

	setIfAny(p, "diagnosis", req.Diagnosis)
	setIfAny(p, "procedure", req.Procedure)
	setIfAny(p, "supportingInfo", req.SupportingInfo)
	setIfAny(p, "item", req.Item)
	setIfAny(p, "extension", req.Extension)

	if req.TotalValue > 0 {
		p.Set("total", map[string]interface{}{"value": req.TotalValue, "currency": def(req.TotalCurrency, "IDR")})
	}

	return p
}

func def(v, d string) string {
	if v == "" {
		return d
	}
	return v
}

func ref(resource, id string) map[string]interface{} {
	return map[string]interface{}{"reference": resource + "/" + id}
}

func setIfAny(p satusehat.FHIRPayload, key string, arr []map[string]interface{}) {
	if len(arr) > 0 {
		p.Set(key, arr)
	}
}
