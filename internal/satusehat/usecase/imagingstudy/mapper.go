package imagingstudy

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req ImagingStudyRequest) satusehat.FHIRPayload {
	orgID := req.OrganizationID

	payload := satusehat.NewFHIRPayload("ImagingStudy")

	if req.Status != "" {
		payload.Set("status", req.Status)
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

	if !req.Started.IsZero() {
		payload.Set("started", req.Started.Format(time.RFC3339))
	}

	if orgID != "" && req.AccessionNumber != "" {
		payload.Set("identifier", []map[string]interface{}{
			{
				"use": "official",
				"type": map[string]interface{}{
					"coding": []map[string]interface{}{
						{
							"system": "http://terminology.hl7.org/CodeSystem/v2-0203",
							"code":   "ACSN",
						},
					},
				},
				"system": "http://sys-ids.kemkes.go.id/acsn/" + orgID,
				"value":  req.AccessionNumber,
			},
		})
	}

	if req.ServiceRequestID != "" {
		payload.Set("basedOn", []map[string]interface{}{
			{
				"reference": "ServiceRequest/" + req.ServiceRequestID,
			},
		})
	}

	if req.ProcedureCode != "" {
		modality := map[string]interface{}{
			"system": "http://dicom.nema.org/resources/ontology/DCM",
			"code":   req.ProcedureCode,
		}
		if req.ProcedureDisplay != "" {
			modality["display"] = req.ProcedureDisplay
		}
		payload.Set("modality", []map[string]interface{}{modality})
	}

	// Menambahkan Practitioner ke kolom `interpreter` jika disediakan
	if req.PractitionerID != "" {
		interpreter := map[string]interface{}{
			"reference": "Practitioner/" + req.PractitionerID,
		}
		if req.PractitionerName != "" {
			interpreter["display"] = req.PractitionerName
		}
		payload.Set("interpreter", []map[string]interface{}{interpreter})
	}

	if req.EncounterID != "" {
		payload.Set("encounter", map[string]interface{}{
			"reference": "Encounter/" + req.EncounterID,
		})
	}

	if req.NumberOfSeries != nil {
		payload.Set("numberOfSeries", *req.NumberOfSeries)
	}

	if req.NumberOfInstances != nil {
		payload.Set("numberOfInstances", *req.NumberOfInstances)
	}

	if req.Description != "" {
		payload.Set("description", req.Description)
	}

	return payload
}
