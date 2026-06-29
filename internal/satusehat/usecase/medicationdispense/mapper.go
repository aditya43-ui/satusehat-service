package medicationdispense

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req MedicationDispenseRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("MedicationDispense")

	orgID := req.OrganizationID

	if orgID != "" && req.PrescriptionID != "" {
		identifiers := []map[string]interface{}{
			{
				"system": "http://sys-ids.kemkes.go.id/prescription/" + orgID,
				"use":    "official",
				"value":  req.PrescriptionID,
			},
		}
		if req.PrescriptionItemID != "" {
			identifiers = append(identifiers, map[string]interface{}{
				"system": "http://sys-ids.kemkes.go.id/prescription-item/" + orgID,
				"use":    "official",
				"value":  req.PrescriptionItemID,
			})
		}
		payload.Set("identifier", identifiers)
	}

	if req.Status != "" {
		payload.Set("status", req.Status)
	}

	categoryCode := "outpatient"
	categoryDisplay := "Outpatient"
	if req.Category != "" {
		categoryCode = req.Category
		switch req.Category {
		case "inpatient":
			categoryDisplay = "Inpatient"
		case "community":
			categoryDisplay = "Community"
		case "discharge":
			categoryDisplay = "Discharge"
		}
	}
	payload.Set("category", map[string]interface{}{
		"coding": []map[string]interface{}{
			{
				"system":  "http://terminology.hl7.org/fhir/CodeSystem/medicationdispense-category",
				"code":    categoryCode,
				"display": categoryDisplay,
			},
		},
	})

	if req.MedicationID != "" {
		medRef := map[string]interface{}{"reference": "Medication/" + req.MedicationID}
		if req.MedicationDisplay != "" {
			medRef["display"] = req.MedicationDisplay
		}
		payload.Set("medicationReference", medRef)
	}

	if req.PatientID != "" {
		subj := map[string]interface{}{"reference": "Patient/" + req.PatientID}
		if req.PatientName != "" {
			subj["display"] = req.PatientName
		}
		payload.Set("subject", subj)
	}

	if req.EncounterID != "" {
		payload.Set("context", map[string]interface{}{"reference": "Encounter/" + req.EncounterID})
	}

	if req.PractitionerID != "" {
		actor := map[string]interface{}{"reference": "Practitioner/" + req.PractitionerID}
		if req.PractitionerName != "" {
			actor["display"] = req.PractitionerName
		}
		payload.Set("performer", []map[string]interface{}{{"actor": actor}})
	}

	if req.LocationID != "" {
		loc := map[string]interface{}{"reference": "Location/" + req.LocationID}
		if req.LocationName != "" {
			loc["display"] = req.LocationName
		}
		payload.Set("location", loc)
	}

	if req.MedicationRequestID != "" {
		payload.Set("authorizingPrescription", []map[string]interface{}{{"reference": "MedicationRequest/" + req.MedicationRequestID}})
	}

	if req.QuantityValue != 0 || req.QuantityUnit != "" {
		qty := map[string]interface{}{
			"system": "http://terminology.hl7.org/CodeSystem/v3-orderableDrugForm",
		}
		if req.QuantityValue != 0 {
			qty["value"] = req.QuantityValue
		}
		if req.QuantityUnit != "" {
			qty["code"] = req.QuantityUnit
		}
		payload.Set("quantity", qty)
	}

	if req.DaysSupplyValue != 0 {
		payload.Set("daysSupply", map[string]interface{}{
			"value":  req.DaysSupplyValue,
			"unit":   "Day",
			"system": "http://unitsofmeasure.org",
			"code":   "d",
		})
	}

	if !req.PreparedDate.IsZero() {
		payload.Set("whenPrepared", req.PreparedDate.Format(time.RFC3339))
	}
	if !req.HandedOverDate.IsZero() {
		payload.Set("whenHandedOver", req.HandedOverDate.Format(time.RFC3339))
	}

	if req.DosageText != "" || req.TimingFrequency != 0 || req.DoseQuantityValue != 0 {
		dose := map[string]interface{}{"sequence": 1}
		if req.DosageText != "" {
			dose["text"] = req.DosageText
		}
		if req.TimingFrequency != 0 && req.TimingPeriod != 0 && req.TimingPeriodUnit != "" {
			dose["timing"] = map[string]interface{}{
				"repeat": map[string]interface{}{
					"frequency":  req.TimingFrequency,
					"period":     req.TimingPeriod,
					"periodUnit": req.TimingPeriodUnit,
				},
			}
		}
		if req.DoseQuantityValue != 0 && req.DoseQuantityUnit != "" {
			dose["doseAndRate"] = []map[string]interface{}{
				{
					"type": map[string]interface{}{
						"coding": []map[string]interface{}{
							{
								"system":  "http://terminology.hl7.org/CodeSystem/dose-rate-type",
								"code":    "ordered",
								"display": "Ordered",
							},
						},
					},
					"doseQuantity": map[string]interface{}{
						"value":  req.DoseQuantityValue,
						"unit":   req.DoseQuantityUnit,
						"system": "http://terminology.hl7.org/CodeSystem/v3-orderableDrugForm",
						"code":   req.DoseQuantityUnit,
					},
				},
			}
		}
		payload.Set("dosageInstruction", []map[string]interface{}{dose})
	}

	return payload
}
