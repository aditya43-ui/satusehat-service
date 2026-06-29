package servicerequest

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req ServiceRequestRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("ServiceRequest")

	if req.OrganizationID != "" && req.ServiceRequestID != "" {
		payload.Set("identifier", []map[string]interface{}{
			{
				"system": "http://sys-ids.kemkes.go.id/servicerequest/" + req.OrganizationID,
				"use":    "official",
				"value":  req.ServiceRequestID,
			},
		})
	}

	if req.Status != "" {
		payload.Set("status", req.Status)
	}

	if req.Intent != "" {
		payload.Set("intent", req.Intent)
	}

	if req.Code != "" {
		coding := map[string]interface{}{
			"system": "http://snomed.info/sct",
			"code":   req.Code,
		}
		if req.Display != "" {
			coding["display"] = req.Display
		}
		payload.Set("code", map[string]interface{}{
			"coding": []map[string]interface{}{coding},
		})
	}

	if req.PatientID != "" {
		subject := map[string]interface{}{
			"reference": "Patient/" + req.PatientID,
		}
		if req.PatientName != "" {
			subject["display"] = req.PatientName
		}
		payload.Set("subject", subject)
	}

	if req.EncounterID != "" {
		payload.Set("encounter", map[string]interface{}{
			"reference": "Encounter/" + req.EncounterID,
		})
	}

	if !req.AuthoredOn.IsZero() {
		payload.Set("authoredOn", req.AuthoredOn.Format(time.RFC3339))
	}
	if req.PerformerID != "" {
		performer := map[string]interface{}{
			"reference": "Practitioner/" + req.PerformerID,
		}
		if req.PerformerName != "" {
			performer["display"] = req.PerformerName
		}
		payload.Set("performer", []map[string]interface{}{performer})
	}

	if req.RequesterID != "" {
		requester := map[string]interface{}{
			"reference": "Practitioner/" + req.RequesterID,
		}
		if req.RequesterName != "" {
			requester["display"] = req.RequesterName
		}
		payload.Set("requester", requester)
	}

	return payload
}
