package observation

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req ObservationRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("Observation")

	if req.ObservationID != "" && req.OrganizationID != "" {
		payload.Set("identifier", []map[string]interface{}{
			{
				"system": "http://sys-ids.kemkes.go.id/observation/" + req.OrganizationID,
				"value":  req.ObservationID,
			},
		})
	}

	status := req.Status
	if status == "" {
		status = "final"
	}
	payload.Set("status", status)

	if req.CategoryCode != "" {
		sys := req.CategorySystem
		if sys == "" {
			sys = "http://terminology.hl7.org/CodeSystem/observation-category"
		}
		coding := map[string]interface{}{"system": sys, "code": req.CategoryCode}
		if req.CategoryDisplay != "" {
			coding["display"] = req.CategoryDisplay
		}
		payload.Set("category", []map[string]interface{}{
			{"coding": []map[string]interface{}{coding}},
		})
	}

	if req.Code != "" {
		sys := req.CodeSystem
		if sys == "" {
			sys = "http://loinc.org"
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
		payload.Set("encounter", map[string]interface{}{"reference": "Encounter/" + req.EncounterID})
	}

	if req.EffectiveDateTime != nil && !req.EffectiveDateTime.IsZero() {
		payload.Set("effectiveDateTime", req.EffectiveDateTime.Format(time.RFC3339))
	}
	if req.Issued != nil && !req.Issued.IsZero() {
		payload.Set("issued", req.Issued.Format(time.RFC3339))
	}

	if req.PerformerID != "" {
		perf := map[string]interface{}{"reference": "Practitioner/" + req.PerformerID}
		if req.PerformerName != "" {
			perf["display"] = req.PerformerName
		}
		payload.Set("performer", []map[string]interface{}{perf})
	}

	if req.SpecimenID != "" {
		payload.Set("specimen", map[string]interface{}{"reference": "Specimen/" + req.SpecimenID})
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
		payload.Set("bodySite", map[string]interface{}{"coding": []map[string]interface{}{coding}})
	}

	switch {
	case req.ValueQuantityValue != nil:
		sys := req.ValueQuantitySystem
		if sys == "" {
			sys = "http://unitsofmeasure.org"
		}
		q := map[string]interface{}{
			"value":  *req.ValueQuantityValue,
			"system": sys,
		}
		if req.ValueQuantityUnit != "" {
			q["unit"] = req.ValueQuantityUnit
		}
		if req.ValueQuantityCode != "" {
			q["code"] = req.ValueQuantityCode
		}
		payload.Set("valueQuantity", q)
	case req.ValueCode != "":
		sys := req.ValueCodeSystem
		if sys == "" {
			sys = "http://snomed.info/sct"
		}
		coding := map[string]interface{}{"system": sys, "code": req.ValueCode}
		if req.ValueCodeDisplay != "" {
			coding["display"] = req.ValueCodeDisplay
		}
		payload.Set("valueCodeableConcept", map[string]interface{}{"coding": []map[string]interface{}{coding}})
	case req.ValueString != "":
		payload.Set("valueString", req.ValueString)
	case req.ValueBoolean != nil:
		payload.Set("valueBoolean", *req.ValueBoolean)
	}

	if req.InterpretationCode != "" {
		sys := req.InterpretationSystem
		if sys == "" {
			sys = "http://terminology.hl7.org/CodeSystem/v3-ObservationInterpretation"
		}
		coding := map[string]interface{}{"system": sys, "code": req.InterpretationCode}
		if req.InterpretationDisplay != "" {
			coding["display"] = req.InterpretationDisplay
		}
		payload.Set("interpretation", []map[string]interface{}{
			{"coding": []map[string]interface{}{coding}},
		})
	}

	if req.ReferenceRangeLowValue != nil || req.ReferenceRangeHighValue != nil || req.ReferenceRangeText != "" {
		rng := map[string]interface{}{}
		if req.ReferenceRangeLowValue != nil {
			low := map[string]interface{}{"value": *req.ReferenceRangeLowValue, "system": "http://unitsofmeasure.org"}
			if req.ReferenceRangeUnit != "" {
				low["unit"] = req.ReferenceRangeUnit
				low["code"] = req.ReferenceRangeUnit
			}
			rng["low"] = low
		}
		if req.ReferenceRangeHighValue != nil {
			high := map[string]interface{}{"value": *req.ReferenceRangeHighValue, "system": "http://unitsofmeasure.org"}
			if req.ReferenceRangeUnit != "" {
				high["unit"] = req.ReferenceRangeUnit
				high["code"] = req.ReferenceRangeUnit
			}
			rng["high"] = high
		}
		if req.ReferenceRangeText != "" {
			rng["text"] = req.ReferenceRangeText
		}
		payload.Set("referenceRange", []map[string]interface{}{rng})
	}

	return payload
}
