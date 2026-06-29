package specimen

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req SpecimenRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("Specimen")

	if req.OrganizationID != "" && req.SpecimenID != "" {
		payload.Set("identifier", []map[string]interface{}{
			{
				"system": "http://sys-ids.kemkes.go.id/specimen/" + req.OrganizationID,
				"value":  req.SpecimenID,
				"assigner": map[string]interface{}{
					"reference": "Organization/" + req.OrganizationID,
				},
			},
		})
	}

	if req.Status != "" {
		payload.Set("status", req.Status)
	}

	if req.TypeCode != "" {
		sys := req.TypeSystem
		if sys == "" {
			sys = "http://snomed.info/sct"
		}
		coding := map[string]interface{}{"system": sys, "code": req.TypeCode}
		if req.TypeDisplay != "" {
			coding["display"] = req.TypeDisplay
		}
		payload.Set("type", map[string]interface{}{"coding": []map[string]interface{}{coding}})
	}

	if req.PatientID != "" {
		subject := map[string]interface{}{"reference": "Patient/" + req.PatientID}
		if req.PatientName != "" {
			subject["display"] = req.PatientName
		}
		payload.Set("subject", subject)
	}

	if req.ReceivedDateTime != nil && !req.ReceivedDateTime.IsZero() {
		payload.Set("receivedTime", req.ReceivedDateTime.Format(time.RFC3339))
	}

	coll := map[string]interface{}{}
	if req.CollectorID != "" {
		collector := map[string]interface{}{"reference": "Practitioner/" + req.CollectorID}
		if req.CollectorName != "" {
			collector["display"] = req.CollectorName
		}
		coll["collector"] = collector
	}
	if req.CollectedDateTime != nil && !req.CollectedDateTime.IsZero() {
		coll["collectedDateTime"] = req.CollectedDateTime.Format(time.RFC3339)
	}
	if req.CollectionQuantityValue != nil {
		sys := req.CollectionQuantitySystem
		if sys == "" {
			sys = "http://unitsofmeasure.org"
		}
		q := map[string]interface{}{"value": *req.CollectionQuantityValue, "system": sys}
		if req.CollectionQuantityUnit != "" {
			q["unit"] = req.CollectionQuantityUnit
			q["code"] = req.CollectionQuantityUnit
		}
		coll["quantity"] = q
	}
	if req.CollectionMethodCode != "" {
		sys := req.CollectionMethodSystem
		if sys == "" {
			sys = "http://snomed.info/sct"
		}
		coding := map[string]interface{}{"system": sys, "code": req.CollectionMethodCode}
		if req.CollectionMethodDisplay != "" {
			coding["display"] = req.CollectionMethodDisplay
		}
		coll["method"] = map[string]interface{}{"coding": []map[string]interface{}{coding}}
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
		coll["bodySite"] = map[string]interface{}{"coding": []map[string]interface{}{coding}}
	}
	if req.FastingStatusCode != "" {
		sys := req.FastingStatusSystem
		if sys == "" {
			sys = "http://terminology.hl7.org/CodeSystem/v2-0916"
		}
		coding := map[string]interface{}{"system": sys, "code": req.FastingStatusCode}
		if req.FastingStatusDisplay != "" {
			coding["display"] = req.FastingStatusDisplay
		}
		coll["fastingStatusCodeableConcept"] = map[string]interface{}{"coding": []map[string]interface{}{coding}}
	}
	if len(coll) > 0 {
		payload.Set("collection", coll)
	}

	if req.ProcessingProcedureCode != "" {
		sys := req.ProcessingProcedureSystem
		if sys == "" {
			sys = "http://snomed.info/sct"
		}
		coding := map[string]interface{}{"system": sys, "code": req.ProcessingProcedureCode}
		if req.ProcessingProcedureDisplay != "" {
			coding["display"] = req.ProcessingProcedureDisplay
		}
		p := map[string]interface{}{
			"procedure": map[string]interface{}{"coding": []map[string]interface{}{coding}},
		}
		if req.ProcessingTimeDateTime != nil && !req.ProcessingTimeDateTime.IsZero() {
			p["timeDateTime"] = req.ProcessingTimeDateTime.Format(time.RFC3339)
		}
		payload.Set("processing", []map[string]interface{}{p})
	}

	if len(req.Conditions) > 0 {
		conditions := make([]map[string]interface{}, 0, len(req.Conditions))
		for _, c := range req.Conditions {
			conditions = append(conditions, map[string]interface{}{"text": c})
		}
		payload.Set("condition", conditions)
	}

	if len(req.RequestServiceRequestIDs) > 0 {
		requests := make([]map[string]interface{}, 0, len(req.RequestServiceRequestIDs))
		for _, id := range req.RequestServiceRequestIDs {
			requests = append(requests, map[string]interface{}{"reference": "ServiceRequest/" + id})
		}
		payload.Set("request", requests)
	}

	return payload
}
