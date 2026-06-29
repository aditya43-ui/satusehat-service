package medicationstatement

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req MedicationStatementRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("MedicationStatement")

	if req.OrganizationID != "" && req.StatementID != "" {
		payload.Set("identifier", []map[string]interface{}{
			{
				"system": "http://sys-ids.kemkes.go.id/medicationstatement/" + req.OrganizationID,
				"use":    "official",
				"value":  req.StatementID,
			},
		})
	}

	if req.Status != "" {
		payload.Set("status", req.Status)
	}

	if req.CategoryCode != "" {
		sys := req.CategorySystem
		if sys == "" {
			sys = "http://terminology.hl7.org/CodeSystem/medication-statement-category"
		}
		coding := map[string]interface{}{"system": sys, "code": req.CategoryCode}
		if req.CategoryDisplay != "" {
			coding["display"] = req.CategoryDisplay
		}
		payload.Set("category", map[string]interface{}{
			"coding": []map[string]interface{}{coding},
		})
	}

	if req.MedicationCode != "" {
		sys := req.MedicationCodeSystem
		if sys == "" {
			sys = "http://sys-ids.kemkes.go.id/kfa"
		}
		coding := map[string]interface{}{"system": sys, "code": req.MedicationCode}
		if req.MedicationDisplay != "" {
			coding["display"] = req.MedicationDisplay
		}
		payload.Set("medicationCodeableConcept", map[string]interface{}{
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
		payload.Set("context", map[string]interface{}{"reference": "Encounter/" + req.EncounterID})
	}

	if req.EffectiveDateTime != nil && !req.EffectiveDateTime.IsZero() {
		payload.Set("effectiveDateTime", req.EffectiveDateTime.Format(time.RFC3339))
	}
	if req.DateAsserted != nil && !req.DateAsserted.IsZero() {
		payload.Set("dateAsserted", req.DateAsserted.Format(time.RFC3339))
	}

	if req.InformationSourceID != "" {
		srcType := req.InformationSourceType
		if srcType == "" {
			srcType = "Patient"
		}
		infoSource := map[string]interface{}{"reference": srcType + "/" + req.InformationSourceID}
		if req.InformationSourceName != "" {
			infoSource["display"] = req.InformationSourceName
		}
		payload.Set("informationSource", infoSource)
	}

	if req.DosageText != "" || req.DosagePatientInstruction != "" || req.DosageRouteCode != "" || req.DoseQuantityValue != 0 {
		dose := map[string]interface{}{"sequence": 1}
		if req.DosageText != "" {
			dose["text"] = req.DosageText
		}
		if req.DosagePatientInstruction != "" {
			dose["patientInstruction"] = req.DosagePatientInstruction
		}
		if req.DosageRouteCode != "" {
			sys := req.DosageRouteSystem
			if sys == "" {
				sys = "http://terminology.hl7.org/CodeSystem/v3-RouteOfAdministration"
			}
			coding := map[string]interface{}{"system": sys, "code": req.DosageRouteCode}
			if req.DosageRouteDisplay != "" {
				coding["display"] = req.DosageRouteDisplay
			}
			dose["route"] = map[string]interface{}{"coding": []map[string]interface{}{coding}}
		}
		if req.DoseQuantityValue != 0 && req.DoseQuantityUnit != "" {
			dose["doseAndRate"] = []map[string]interface{}{
				{
					"doseQuantity": map[string]interface{}{
						"value":  req.DoseQuantityValue,
						"unit":   req.DoseQuantityUnit,
						"system": "http://terminology.hl7.org/CodeSystem/v3-orderableDrugForm",
						"code":   req.DoseQuantityUnit,
					},
				},
			}
		}
		payload.Set("dosage", []map[string]interface{}{dose})
	}

	return payload
}
