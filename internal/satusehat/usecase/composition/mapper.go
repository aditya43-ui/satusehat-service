package composition

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req CompositionRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("Composition")

	if req.OrganizationID != "" && req.CompositionID != "" {
		payload.Set("identifier", map[string]interface{}{
			"system": "http://sys-ids.kemkes.go.id/composition/" + req.OrganizationID,
			"value":  req.CompositionID,
		})
	}

	status := req.Status
	if status == "" {
		status = "final"
	}
	payload.Set("status", status)

	if req.TypeCode != "" {
		sys := req.TypeSystem
		if sys == "" {
			sys = "http://loinc.org"
		}
		coding := map[string]interface{}{"system": sys, "code": req.TypeCode}
		if req.TypeDisplay != "" {
			coding["display"] = req.TypeDisplay
		}
		payload.Set("type", map[string]interface{}{
			"coding": []map[string]interface{}{coding},
		})
	}

	if req.CategoryCode != "" {
		sys := req.CategorySystem
		if sys == "" {
			sys = "http://loinc.org"
		}
		coding := map[string]interface{}{"system": sys, "code": req.CategoryCode}
		if req.CategoryDisplay != "" {
			coding["display"] = req.CategoryDisplay
		}
		payload.Set("category", []map[string]interface{}{
			{"coding": []map[string]interface{}{coding}},
		})
	}

	if req.PatientID != "" {
		subj := map[string]interface{}{"reference": "Patient/" + req.PatientID}
		if req.PatientDisplay != "" {
			subj["display"] = req.PatientDisplay
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

	if req.Date != nil && !req.Date.IsZero() {
		payload.Set("date", req.Date.Format(time.RFC3339))
	}
	if req.Title != "" {
		payload.Set("title", req.Title)
	}

	if req.AuthorID != "" {
		author := map[string]interface{}{"reference": "Practitioner/" + req.AuthorID}
		if req.AuthorDisplay != "" {
			author["display"] = req.AuthorDisplay
		}
		payload.Set("author", []map[string]interface{}{author})
	}

	if req.SectionTitle != "" || req.SectionCode != "" || req.SectionText != "" {
		sec := map[string]interface{}{}
		if req.SectionTitle != "" {
			sec["title"] = req.SectionTitle
		}
		if req.SectionCode != "" {
			sys := req.SectionSystem
			if sys == "" {
				sys = "http://loinc.org"
			}
			coding := map[string]interface{}{"system": sys, "code": req.SectionCode}
			if req.SectionDisplay != "" {
				coding["display"] = req.SectionDisplay
			}
			sec["code"] = map[string]interface{}{"coding": []map[string]interface{}{coding}}
		}
		if req.SectionText != "" {
			sec["text"] = map[string]interface{}{
				"status": "generated",
				"div":    `<div xmlns="http://www.w3.org/1999/xhtml">` + req.SectionText + `</div>`,
			}
		}
		payload.Set("section", []map[string]interface{}{sec})
	}

	return payload
}
