package condition

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req ConditionRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("Condition")

	if req.OrganizationID != "" && req.ConditionID != "" {
		payload.Set("identifier", []map[string]interface{}{
			{"system": "http://sys-ids.kemkes.go.id/condition/" + req.OrganizationID, "use": "official", "value": req.ConditionID},
		})
	}

	if req.ClinicalStatus != "" {
		payload.Set("clinicalStatus", map[string]interface{}{
			"coding": []map[string]interface{}{
				{"system": "http://terminology.hl7.org/CodeSystem/condition-clinical", "code": req.ClinicalStatus},
			},
		})
	}

	if req.CategoryCode != "" {
		sys := req.CategorySystem
		if sys == "" {
			sys = "http://terminology.hl7.org/CodeSystem/condition-category"
		}
		coding := map[string]interface{}{
			"system": sys,
			"code":   req.CategoryCode,
		}
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
			sys = "http://hl7.org/fhir/sid/icd-10"
		}
		coding := map[string]interface{}{
			"system": sys,
			"code":   req.Code,
		}
		if req.CodeDisplay != "" {
			coding["display"] = req.CodeDisplay
		}
		payload.Set("code", map[string]interface{}{
			"coding": []map[string]interface{}{coding},
		})
	}

	if req.PatientID != "" {
		subject := map[string]interface{}{"reference": "Patient/" + req.PatientID}
		if req.PatientName != "" {
			subject["display"] = req.PatientName
		}
		payload.Set("subject", subject)
	}

	if req.EncounterID != "" {
		payload.Set("encounter", map[string]interface{}{"reference": "Encounter/" + req.EncounterID})
	}

	if req.OnsetDateTime != nil && !req.OnsetDateTime.IsZero() {
		payload.Set("onsetDateTime", req.OnsetDateTime.Format(time.RFC3339))
	}
	if req.RecordedDate != nil && !req.RecordedDate.IsZero() {
		payload.Set("recordedDate", req.RecordedDate.Format(time.RFC3339))
	}

	return payload
}
