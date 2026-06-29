package clinicalimpression

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req ClinicalImpressionRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("ClinicalImpression")

	if req.OrganizationID != "" && req.ClinicalImpressionID != "" {
		payload.Set("identifier", []map[string]interface{}{
			{
				"system": "http://sys-ids.kemkes.go.id/clinicalimpression/" + req.OrganizationID,
				"use":    "official",
				"value":  req.ClinicalImpressionID,
			},
		})
	}

	if req.Status != "" {
		payload.Set("status", req.Status)
	}

	if req.Code != "" {
		sys := req.CodeSystem
		if sys == "" {
			sys = "http://snomed.info/sct"
		}
		coding := map[string]interface{}{"system": sys, "code": req.Code}
		if req.CodeDisplay != "" {
			coding["display"] = req.CodeDisplay
		}
		payload.Set("code", map[string]interface{}{
			"coding": []map[string]interface{}{coding},
		})
	}

	if req.Description != "" {
		payload.Set("description", req.Description)
	}

	if req.PatientID != "" {
		subject := map[string]interface{}{"reference": "Patient/" + req.PatientID}
		if req.PatientDisplay != "" {
			subject["display"] = req.PatientDisplay
		}
		payload.Set("subject", subject)
	}

	if req.EncounterID != "" {
		encounter := map[string]interface{}{"reference": "Encounter/" + req.EncounterID}
		if req.EncounterDisplay != "" {
			encounter["display"] = req.EncounterDisplay
		}
		payload.Set("encounter", encounter)
	}

	if req.EffectiveDateTime != nil && !req.EffectiveDateTime.IsZero() {
		payload.Set("effectiveDateTime", req.EffectiveDateTime.Format(time.RFC3339))
	}

	if req.Date != nil && !req.Date.IsZero() {
		payload.Set("date", req.Date.Format(time.RFC3339))
	}

	if req.AssessorID != "" {
		assessor := map[string]interface{}{"reference": "Practitioner/" + req.AssessorID}
		if req.AssessorDisplay != "" {
			assessor["display"] = req.AssessorDisplay
		}
		payload.Set("assessor", assessor)
	}

	if len(req.ProblemConditionIDs) > 0 {
		problems := make([]map[string]interface{}, 0, len(req.ProblemConditionIDs))
		for _, id := range req.ProblemConditionIDs {
			problems = append(problems, map[string]interface{}{"reference": "Condition/" + id})
		}
		payload.Set("problem", problems)
	}

	if req.Summary != "" {
		payload.Set("summary", req.Summary)
	}

	if req.FindingCode != "" {
		sys := req.FindingSystem
		if sys == "" {
			sys = "http://hl7.org/fhir/sid/icd-10"
		}
		coding := map[string]interface{}{"system": sys, "code": req.FindingCode}
		if req.FindingDisplay != "" {
			coding["display"] = req.FindingDisplay
		}
		payload.Set("finding", []map[string]interface{}{
			{"itemCodeableConcept": map[string]interface{}{"coding": []map[string]interface{}{coding}}},
		})
	}

	if req.PrognosisCode != "" {
		sys := req.PrognosisSystem
		if sys == "" {
			sys = "http://snomed.info/sct"
		}
		coding := map[string]interface{}{"system": sys, "code": req.PrognosisCode}
		if req.PrognosisDisplay != "" {
			coding["display"] = req.PrognosisDisplay
		}
		payload.Set("prognosisCodeableConcept", []map[string]interface{}{
			{"coding": []map[string]interface{}{coding}},
		})
	}

	return payload
}
