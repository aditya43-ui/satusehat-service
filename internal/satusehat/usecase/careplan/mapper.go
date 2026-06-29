package careplan

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req CarePlanRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("CarePlan")

	if req.OrganizationID != "" && req.CarePlanID != "" {
		payload.Set("identifier", []map[string]interface{}{
			{"system": "http://sys-ids.kemkes.go.id/careplan/" + req.OrganizationID, "use": "official", "value": req.CarePlanID},
		})
	}

	if req.Status != "" {
		payload.Set("status", req.Status)
	}
	if req.Intent != "" {
		payload.Set("intent", req.Intent)
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
		payload.Set("category", []map[string]interface{}{
			{"coding": []map[string]interface{}{coding}},
		})
	}

	if req.Title != "" {
		payload.Set("title", req.Title)
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
	if req.AuthorID != "" {
		author := map[string]interface{}{"reference": "Practitioner/" + req.AuthorID}
		if req.AuthorDisplay != "" {
			author["display"] = req.AuthorDisplay
		}
		payload.Set("author", author)
	}
	if req.CreatedDate != nil && !req.CreatedDate.IsZero() {
		payload.Set("created", req.CreatedDate.Format(time.RFC3339))
	}
	if len(req.GoalIDs) > 0 {
		goals := make([]map[string]interface{}, 0, len(req.GoalIDs))
		for _, id := range req.GoalIDs {
			goals = append(goals, map[string]interface{}{"reference": "Goal/" + id})
		}
		payload.Set("goal", goals)
	}

	return payload
}
