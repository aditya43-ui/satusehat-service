package immunization

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req ImmunizationRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("Immunization")

	if req.OrganizationID != "" && req.ImmunizationID != "" {
		payload.Set("identifier", []map[string]interface{}{
			{
				"system": "http://sys-ids.kemkes.go.id/immunization/" + req.OrganizationID,
				"use":    "official",
				"value":  req.ImmunizationID,
			},
		})
	}

	if req.Status != "" {
		payload.Set("status", req.Status)
	}

	if req.VaccineCode != "" {
		sys := req.VaccineCodeSystem
		if sys == "" {
			sys = "http://sys-ids.kemkes.go.id/kfa"
		}
		coding := map[string]interface{}{"system": sys, "code": req.VaccineCode}
		if req.VaccineCodeDisplay != "" {
			coding["display"] = req.VaccineCodeDisplay
		}
		payload.Set("vaccineCode", map[string]interface{}{
			"coding": []map[string]interface{}{coding},
		})
	}

	if req.PatientID != "" {
		subj := map[string]interface{}{"reference": "Patient/" + req.PatientID}
		if req.PatientName != "" {
			subj["display"] = req.PatientName
		}
		payload.Set("patient", subj)
	}

	if req.EncounterID != "" {
		enc := map[string]interface{}{"reference": "Encounter/" + req.EncounterID}
		if req.EncounterDisplay != "" {
			enc["display"] = req.EncounterDisplay
		}
		payload.Set("encounter", enc)
	}

	if req.OccurrenceDateTime != nil && !req.OccurrenceDateTime.IsZero() {
		payload.Set("occurrenceDateTime", req.OccurrenceDateTime.Format(time.RFC3339))
	}

	if req.PrimarySource != nil {
		payload.Set("primarySource", *req.PrimarySource)
	}

	if req.LocationID != "" {
		loc := map[string]interface{}{"reference": "Location/" + req.LocationID}
		if req.LocationName != "" {
			loc["display"] = req.LocationName
		}
		payload.Set("location", loc)
	}

	if req.LotNumber != "" {
		payload.Set("lotNumber", req.LotNumber)
	}

	if req.PerformerID != "" {
		actor := map[string]interface{}{"reference": "Practitioner/" + req.PerformerID}
		if req.PerformerName != "" {
			actor["display"] = req.PerformerName
		}
		payload.Set("performer", []map[string]interface{}{{"actor": actor}})
	}

	if req.DoseQuantityValue != 0 || req.DoseQuantityUnit != "" {
		sys := req.DoseQuantitySystem
		if sys == "" {
			sys = "http://unitsofmeasure.org"
		}
		q := map[string]interface{}{
			"value":  req.DoseQuantityValue,
			"system": sys,
		}
		if req.DoseQuantityUnit != "" {
			q["unit"] = req.DoseQuantityUnit
		}
		if req.DoseQuantityCode != "" {
			q["code"] = req.DoseQuantityCode
		} else if req.DoseQuantityUnit != "" {
			q["code"] = req.DoseQuantityUnit
		}
		payload.Set("doseQuantity", q)
	}

	if req.RouteCode != "" {
		sys := req.RouteSystem
		if sys == "" {
			sys = "http://terminology.hl7.org/CodeSystem/v3-RouteOfAdministration"
		}
		coding := map[string]interface{}{"system": sys, "code": req.RouteCode}
		if req.RouteDisplay != "" {
			coding["display"] = req.RouteDisplay
		}
		payload.Set("route", map[string]interface{}{
			"coding": []map[string]interface{}{coding},
		})
	}

	return payload
}
