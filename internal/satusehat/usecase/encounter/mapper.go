package encounter

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req EncounterRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("Encounter")

	if req.OrganizationID != "" && req.EncounterID != "" {
		payload.Set("identifier", []map[string]interface{}{
			{
				"system": "http://sys-ids.kemkes.go.id/encounter/" + req.OrganizationID,
				"value":  req.EncounterID,
			},
		})
	}

	if req.Status != "" {
		payload.Set("status", req.Status)
	}

	if req.Class != "" {
		classDisplay := "ambulatory"
		switch req.Class {
		case "EMER":
			classDisplay = "emergency"
		case "IMP":
			classDisplay = "inpatient encounter"
		}
		payload.Set("class", map[string]interface{}{
			"system":  "http://terminology.hl7.org/CodeSystem/v3-ActCode",
			"code":    req.Class,
			"display": classDisplay,
		})
	}

	if req.PatientID != "" {
		subj := map[string]interface{}{"reference": "Patient/" + req.PatientID}
		if req.PatientName != "" {
			subj["display"] = req.PatientName
		}
		payload.Set("subject", subj)
	}

	if req.PractitionerID != "" {
		ind := map[string]interface{}{"reference": "Practitioner/" + req.PractitionerID}
		if req.PractitionerName != "" {
			ind["display"] = req.PractitionerName
		}
		payload.Set("participant", []map[string]interface{}{
			{
				"type": []map[string]interface{}{
					{
						"coding": []map[string]interface{}{
							{"system": "http://terminology.hl7.org/CodeSystem/v3-ParticipationType", "code": "ATND", "display": "attender"},
						},
					},
				},
				"individual": ind,
			},
		})
	}

	if req.LocationID != "" {
		locRef := map[string]interface{}{"reference": "Location/" + req.LocationID}
		if req.LocationName != "" {
			locRef["display"] = req.LocationName
		}
		locEntry := map[string]interface{}{"location": locRef}
		if req.PeriodStart != nil && !req.PeriodStart.IsZero() {
			period := map[string]interface{}{"start": req.PeriodStart.Format(time.RFC3339)}
			if req.PeriodEnd != nil && !req.PeriodEnd.IsZero() {
				period["end"] = req.PeriodEnd.Format(time.RFC3339)
			}
			locEntry["period"] = period
		}
		payload.Set("location", []map[string]interface{}{locEntry})
	}

	period := map[string]interface{}{}
	if req.PeriodStart != nil && !req.PeriodStart.IsZero() {
		period["start"] = req.PeriodStart.Format(time.RFC3339)
	}
	if req.PeriodEnd != nil && !req.PeriodEnd.IsZero() {
		period["end"] = req.PeriodEnd.Format(time.RFC3339)
	}
	if len(period) > 0 {
		payload.Set("period", period)
	}

	if req.DiagnosisConditionID != "" {
		diag := map[string]interface{}{
			"condition": map[string]interface{}{"reference": "Condition/" + req.DiagnosisConditionID},
		}
		if req.DiagnosisUseCode != "" {
			sys := req.DiagnosisUseSystem
			if sys == "" {
				sys = "http://terminology.hl7.org/CodeSystem/diagnosis-role"
			}
			useCoding := map[string]interface{}{"system": sys, "code": req.DiagnosisUseCode}
			if req.DiagnosisUseDisplay != "" {
				useCoding["display"] = req.DiagnosisUseDisplay
			}
			diag["use"] = map[string]interface{}{"coding": []map[string]interface{}{useCoding}}
		}
		if req.DiagnosisRank > 0 {
			diag["rank"] = req.DiagnosisRank
		}
		payload.Set("diagnosis", []map[string]interface{}{diag})
	}

	if req.Status != "" && req.PeriodStart != nil && !req.PeriodStart.IsZero() {
		shp := map[string]interface{}{"start": req.PeriodStart.Format(time.RFC3339)}
		if req.PeriodEnd != nil && !req.PeriodEnd.IsZero() {
			shp["end"] = req.PeriodEnd.Format(time.RFC3339)
		}
		payload.Set("statusHistory", []map[string]interface{}{
			{"status": req.Status, "period": shp},
		})
	}

	if req.OrganizationID != "" {
		payload.Set("serviceProvider", map[string]interface{}{
			"reference": "Organization/" + req.OrganizationID,
		})
	}

	return payload
}
