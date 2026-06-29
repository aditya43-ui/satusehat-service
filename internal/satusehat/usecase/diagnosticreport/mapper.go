package diagnosticreport

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req DiagnosticReportRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("DiagnosticReport")

	if req.OrganizationID != "" && req.DiagnosticID != "" {
		payload.Set("identifier", []map[string]interface{}{
			{
				"system": "http://sys-ids.kemkes.go.id/diagnostic/" + req.OrganizationID + "/lab",
				"use":    "official",
				"value":  req.DiagnosticID,
			},
		})
	}

	if req.Status != "" {
		payload.Set("status", req.Status)
	}

	if req.CategoryCode != "" {
		sys := req.CategorySystem
		if sys == "" {
			sys = "http://terminology.hl7.org/CodeSystem/v2-0074"
		}
		coding := map[string]interface{}{"system": sys, "code": req.CategoryCode}
		if req.CategoryDisplay != "" {
			coding["display"] = req.CategoryDisplay
		}
		payload.Set("category", []map[string]interface{}{
			{"coding": []map[string]interface{}{coding}},
		})
	}

	if req.Code != "" {
		sys := req.CodeSystem
		if sys == "" {
			sys = "http://loinc.org"
		}
		coding := map[string]interface{}{"system": sys, "code": req.Code}
		if req.CodeDisplay != "" {
			coding["display"] = req.CodeDisplay
		}
		payload.Set("code", map[string]interface{}{
			"coding": []map[string]interface{}{coding},
		})
	}

	if req.PatientID != "" {
		payload.Set("subject", map[string]interface{}{"reference": "Patient/" + req.PatientID})
	}
	if req.EncounterID != "" {
		payload.Set("encounter", map[string]interface{}{"reference": "Encounter/" + req.EncounterID})
	}

	if req.EffectiveDateTime != nil && !req.EffectiveDateTime.IsZero() {
		payload.Set("effectiveDateTime", req.EffectiveDateTime.Format(time.RFC3339))
	}
	if req.Issued != nil && !req.Issued.IsZero() {
		payload.Set("issued", req.Issued.Format(time.RFC3339))
	}

	if req.PerformerID != "" {
		perf := map[string]interface{}{"reference": "Practitioner/" + req.PerformerID}
		if req.PerformerName != "" {
			perf["display"] = req.PerformerName
		}
		payload.Set("performer", []map[string]interface{}{perf})
	}

	if len(req.ResultObservationIDs) > 0 {
		results := make([]map[string]interface{}, 0, len(req.ResultObservationIDs))
		for _, id := range req.ResultObservationIDs {
			results = append(results, map[string]interface{}{"reference": "Observation/" + id})
		}
		payload.Set("result", results)
	}
	if len(req.SpecimenIDs) > 0 {
		specs := make([]map[string]interface{}, 0, len(req.SpecimenIDs))
		for _, id := range req.SpecimenIDs {
			specs = append(specs, map[string]interface{}{"reference": "Specimen/" + id})
		}
		payload.Set("specimen", specs)
	}
	if len(req.BasedOnIDs) > 0 {
		basedOns := make([]map[string]interface{}, 0, len(req.BasedOnIDs))
		for _, id := range req.BasedOnIDs {
			basedOns = append(basedOns, map[string]interface{}{"reference": "ServiceRequest/" + id})
		}
		payload.Set("basedOn", basedOns)
	}
	if len(req.ImagingStudyIDs) > 0 {
		studies := make([]map[string]interface{}, 0, len(req.ImagingStudyIDs))
		for _, id := range req.ImagingStudyIDs {
			studies = append(studies, map[string]interface{}{"reference": "ImagingStudy/" + id})
		}
		payload.Set("imagingStudy", studies)
	}

	if req.ConclusionCode != "" {
		sys := req.ConclusionCodeSystem
		if sys == "" {
			sys = "http://snomed.info/sct"
		}
		coding := map[string]interface{}{"system": sys, "code": req.ConclusionCode}
		if req.ConclusionCodeDisplay != "" {
			coding["display"] = req.ConclusionCodeDisplay
		}
		payload.Set("conclusionCode", []map[string]interface{}{
			{"coding": []map[string]interface{}{coding}},
		})
	}

	if req.Conclusion != "" {
		payload.Set("conclusion", req.Conclusion)
	}

	return payload
}
