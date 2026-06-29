package medication

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req MedicationRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("Medication")

	orgID := req.OrganizationID

	// Set Meta Profile secara otomatis
	payload.Set("meta", map[string]interface{}{
		"profile": []string{"https://fhir.kemkes.go.id/r4/StructureDefinition/Medication"},
	})

	// Set Extension MedicationType (Non-compound) secara otomatis
	payload.Set("extension", []map[string]interface{}{
		{
			"url": "https://fhir.kemkes.go.id/r4/StructureDefinition/MedicationType",
			"valueCodeableConcept": map[string]interface{}{
				"coding": []map[string]interface{}{
					{
						"system":  "http://terminology.kemkes.go.id/CodeSystem/medication-type",
						"code":    "NC",
						"display": "Non-compound",
					},
				},
			},
		},
	})

	if orgID != "" && req.MedicationID != "" {
		payload.Set("identifier", []map[string]interface{}{
			{
				"system": "http://sys-ids.kemkes.go.id/medication/" + orgID,
				"use":    "official",
				"value":  req.MedicationID,
			},
		})
	}

	if req.StatusCode != "" {
		payload.Set("status", req.StatusCode)
	}

	if req.KfaCode != "" {
		coding := map[string]interface{}{
			"system": "http://sys-ids.kemkes.go.id/kfa",
			"code":   req.KfaCode,
		}
		if req.KfaDisplay != "" {
			coding["display"] = req.KfaDisplay
		}
		payload.Set("code", map[string]interface{}{
			"coding": []map[string]interface{}{coding},
		})
	}

	if req.FormCode != "" {
		coding := map[string]interface{}{
			"system": "http://terminology.kemkes.go.id/CodeSystem/medication-form",
			"code":   req.FormCode,
		}
		if req.FormDisplay != "" {
			coding["display"] = req.FormDisplay
		}
		payload.Set("form", map[string]interface{}{
			"coding": []map[string]interface{}{coding},
		})
	}

	if req.ManufacturerID != "" {
		payload.Set("manufacturer", map[string]interface{}{
			"reference": "Organization/" + req.ManufacturerID,
		})
	}

	if req.BatchNumber != "" || req.ExpirationDate != nil {
		batch := map[string]interface{}{}
		if req.BatchNumber != "" {
			batch["lotNumber"] = req.BatchNumber
		}
		if req.ExpirationDate != nil {
			batch["expirationDate"] = req.ExpirationDate.Format(time.RFC3339)
		}
		payload.Set("batch", batch)
	}

	return payload
}
