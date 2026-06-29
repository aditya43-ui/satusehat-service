package medicationrequest

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req MedicationRequestRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("MedicationRequest")

	orgID := req.OrganizationID

	// Set Meta Profile
	payload.Set("meta", map[string]interface{}{
		"profile": []string{"https://fhir.kemkes.go.id/r4/StructureDefinition/MedicationRequest"},
	})

	// Set Identifier (ID Resep)
	if orgID != "" && req.MedicationRequestID != "" {
		identifiers := []map[string]interface{}{
			{
				"system": "http://sys-ids.kemkes.go.id/prescription/" + orgID,
				"use":    "official",
				"value":  req.MedicationRequestID,
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
	if req.Intent != "" {
		payload.Set("intent", req.Intent)
	}
	if req.Priority != "" {
		payload.Set("priority", req.Priority)
	}

	// Penanganan Category dinamis dengan default 'outpatient'
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
	payload.Set("category", []map[string]interface{}{
		{
			"coding": []map[string]interface{}{
				{
					"system":  "http://terminology.hl7.org/CodeSystem/medicationrequest-category",
					"code":    categoryCode,
					"display": categoryDisplay,
				},
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

	if req.ReasonCode != "" {
		reasonCoding := map[string]interface{}{
			"system": "http://hl7.org/fhir/sid/icd-10",
			"code":   req.ReasonCode,
		}
		if req.ReasonDisplay != "" {
			reasonCoding["display"] = req.ReasonDisplay
		}
		payload.Set("reasonCode", []map[string]interface{}{
			{"coding": []map[string]interface{}{reasonCoding}},
		})
	}

	if req.CourseOfTherapyCode != "" {
		courseCoding := map[string]interface{}{
			"system": "http://terminology.hl7.org/CodeSystem/medicationrequest-course-of-therapy",
			"code":   req.CourseOfTherapyCode,
		}
		if req.CourseOfTherapyDisplay != "" {
			courseCoding["display"] = req.CourseOfTherapyDisplay
		}
		payload.Set("courseOfTherapyType", map[string]interface{}{
			"coding": []map[string]interface{}{courseCoding},
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
		payload.Set("encounter", map[string]interface{}{"reference": "Encounter/" + req.EncounterID})
	}

	if !req.AuthoredOn.IsZero() {
		payload.Set("authoredOn", req.AuthoredOn.Format(time.RFC3339))
	}

	if req.PractitionerID != "" {
		reqRef := map[string]interface{}{"reference": "Practitioner/" + req.PractitionerID}
		if req.PractitionerName != "" {
			reqRef["display"] = req.PractitionerName
		}
		payload.Set("requester", reqRef)
	}

	if req.DosageText != "" || req.PatientInstr != "" || req.DoseQuantityValue != 0 || req.TimingFrequency != 0 {
		dose := map[string]interface{}{"sequence": 1}
		if req.DosageText != "" {
			dose["text"] = req.DosageText
		}
		if req.AdditionalInstruction != "" {
			dose["additionalInstruction"] = []map[string]interface{}{
				{"text": req.AdditionalInstruction},
			}
		}
		if req.PatientInstr != "" {
			dose["patientInstruction"] = req.PatientInstr
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
		if req.RouteCode != "" {
			routeCoding := map[string]interface{}{
				"system": "http://www.whocc.no/atc",
				"code":   req.RouteCode,
			}
			if req.RouteDisplay != "" {
				routeCoding["display"] = req.RouteDisplay
			}
			dose["route"] = map[string]interface{}{
				"coding": []map[string]interface{}{routeCoding},
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

	if req.DispenseValue != 0 || req.SupplyDuration != 0 || req.DispenseInterval != 0 {
		dispenseReq := map[string]interface{}{}
		if req.DispenseInterval != 0 {
			dispenseReq["dispenseInterval"] = map[string]interface{}{
				"value":  req.DispenseInterval,
				"unit":   "days",
				"system": "http://unitsofmeasure.org",
				"code":   "d",
			}
		}
		if req.ValidityPeriodStart != nil || req.ValidityPeriodEnd != nil {
			validity := map[string]interface{}{}
			if req.ValidityPeriodStart != nil {
				validity["start"] = req.ValidityPeriodStart.Format("2006-01-02")
			}
			if req.ValidityPeriodEnd != nil {
				validity["end"] = req.ValidityPeriodEnd.Format("2006-01-02")
			}
			dispenseReq["validityPeriod"] = validity
		}

		dispenseReq["numberOfRepeatsAllowed"] = 0

		if req.DispenseValue != 0 {
			qty := map[string]interface{}{
				"value":  req.DispenseValue,
				"system": "http://terminology.hl7.org/CodeSystem/v3-orderableDrugForm",
			}
			if req.DispenseUnit != "" {
				qty["unit"] = req.DispenseUnit
				qty["code"] = req.DispenseUnit
			}
			dispenseReq["quantity"] = qty
		}
		if req.SupplyDuration != 0 {
			dispenseReq["expectedSupplyDuration"] = map[string]interface{}{
				"value":  req.SupplyDuration,
				"unit":   "days",
				"system": "http://unitsofmeasure.org",
				"code":   "d",
			}
		}
		if orgID != "" {
			dispenseReq["performer"] = map[string]interface{}{
				"reference": "Organization/" + orgID,
			}
		}
		payload.Set("dispenseRequest", dispenseReq)
	}

	return payload
}
