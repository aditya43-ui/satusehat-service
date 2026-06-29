package procedure

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req ProcedureRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("Procedure")

	if req.OrganizationID != "" && req.ProcedureID != "" {
		payload.Set("identifier", []map[string]interface{}{
			{
				"system": "http://sys-ids.kemkes.go.id/procedure/" + req.OrganizationID,
				"value":  req.ProcedureID,
			},
		})
	}

	if req.Status != "" {
		payload.Set("status", req.Status)
	}

	if req.CategoryCode != "" {
		sys := req.CategorySystem
		if sys == "" {
			sys = "http://snomed.info/sct"
		}
		coding := map[string]interface{}{"system": sys, "code": req.CategoryCode}
		if req.CategoryDisplay != "" {
			coding["display"] = req.CategoryDisplay
		}
		payload.Set("category", map[string]interface{}{
			"coding": []map[string]interface{}{coding},
		})
	}

	if req.Code != "" {
		sys := req.CodeSystem
		if sys == "" {
			sys = "http://hl7.org/fhir/sid/icd-9-cm"
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
		subj := map[string]interface{}{"reference": "Patient/" + req.PatientID}
		if req.PatientName != "" {
			subj["display"] = req.PatientName
		}
		payload.Set("subject", subj)
	}

	if req.EncounterID != "" {
		enc := map[string]interface{}{"reference": "Encounter/" + req.EncounterID}
		if req.EncounterDisplay != "" {
			enc["display"] = req.EncounterDisplay
		}
		payload.Set("encounter", enc)
	}

	if req.PerformedStart != nil && !req.PerformedStart.IsZero() {
		period := map[string]interface{}{"start": req.PerformedStart.Format(time.RFC3339)}
		if req.PerformedEnd != nil && !req.PerformedEnd.IsZero() {
			period["end"] = req.PerformedEnd.Format(time.RFC3339)
		}
		payload.Set("performedPeriod", period)
	} else if req.PerformedDateTime != nil && !req.PerformedDateTime.IsZero() {
		payload.Set("performedDateTime", req.PerformedDateTime.Format(time.RFC3339))
	}

	if req.PerformerID != "" {
		actor := map[string]interface{}{"reference": "Practitioner/" + req.PerformerID}
		if req.PerformerName != "" {
			actor["display"] = req.PerformerName
		}
		payload.Set("performer", []map[string]interface{}{{"actor": actor}})
	}

	if req.ReasonCode != "" {
		sys := req.ReasonSystem
		if sys == "" {
			sys = "http://hl7.org/fhir/sid/icd-10"
		}
		coding := map[string]interface{}{"system": sys, "code": req.ReasonCode}
		if req.ReasonDisplay != "" {
			coding["display"] = req.ReasonDisplay
		}
		payload.Set("reasonCode", []map[string]interface{}{
			{"coding": []map[string]interface{}{coding}},
		})
	}

	if req.BodySiteCode != "" {
		sys := req.BodySiteSystem
		if sys == "" {
			sys = "http://snomed.info/sct"
		}
		coding := map[string]interface{}{"system": sys, "code": req.BodySiteCode}
		if req.BodySiteDisplay != "" {
			coding["display"] = req.BodySiteDisplay
		}
		payload.Set("bodySite", []map[string]interface{}{
			{"coding": []map[string]interface{}{coding}},
		})
	}

	if req.Note != "" {
		payload.Set("note", []map[string]interface{}{{"text": req.Note}})
	}

	return payload
}
