package allergyintolerance

import (
	"strings"
	"time"

	"service/internal/interfaces/satusehat"
)

var verificationStatusDisplayMap = map[string]string{
	"unconfirmed":      "Unconfirmed",
	"presumed":         "Presumed",
	"confirmed":        "Confirmed",
	"refuted":          "Refuted",
	"entered-in-error": "Entered in Error",
}

func MapRequestToFHIR(req AllergyIntoleranceRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("AllergyIntolerance")

	if req.OrganizationID != "" && req.AllergyID != "" {
		payload.Set("identifier", []map[string]interface{}{
			{
				"system": "http://sys-ids.kemkes.go.id/allergy/" + req.OrganizationID,
				"use":    "official",
				"value":  req.AllergyID,
			},
		})
	}

	if req.ClinicalStatus != "" {
		payload.Set("clinicalStatus", map[string]interface{}{
			"coding": []map[string]interface{}{
				{
					"system":  "http://terminology.hl7.org/CodeSystem/allergyintolerance-clinical",
					"code":    req.ClinicalStatus,
					"display": strings.ToUpper(req.ClinicalStatus[:1]) + req.ClinicalStatus[1:],
				},
			},
		})
	}

	if req.VerificationStatus != "" {
		display, ok := verificationStatusDisplayMap[req.VerificationStatus]
		if !ok {
			display = strings.ToUpper(req.VerificationStatus[:1]) + req.VerificationStatus[1:]
		}
		payload.Set("verificationStatus", map[string]interface{}{
			"coding": []map[string]interface{}{
				{"system": "http://terminology.hl7.org/CodeSystem/allergyintolerance-verification", "code": req.VerificationStatus, "display": display},
			},
		})
	}

	if req.Category != "" {
		payload.Set("category", []string{req.Category})
	}

	if req.Code != "" {
		sys := req.CodeSystem
		if sys == "" {
			sys = "http://snomed.info/sct"
		}
		coding := map[string]interface{}{
			"system": sys,
			"code":   req.Code,
		}
		if req.CodeDisplay != "" {
			coding["display"] = req.CodeDisplay
		}
		payload.Set("code", map[string]interface{}{
			"coding": []map[string]interface{}{coding},
		})
	}

	if req.PatientID != "" {
		subject := map[string]interface{}{"reference": "Patient/" + req.PatientID}
		if req.PatientName != "" {
			subject["display"] = req.PatientName
		}
		payload.Set("patient", subject)
	}

	if req.EncounterID != "" {
		encounter := map[string]interface{}{"reference": "Encounter/" + req.EncounterID}
		if req.EncounterDisplay != "" {
			encounter["display"] = req.EncounterDisplay
		}
		payload.Set("encounter", encounter)
	}

	if req.RecordedDate != nil && !req.RecordedDate.IsZero() {
		payload.Set("recordedDate", req.RecordedDate.Format(time.RFC3339))
	}

	if req.RecorderID != "" {
		recorder := map[string]interface{}{"reference": "Practitioner/" + req.RecorderID}
		if req.RecorderDisplay != "" {
			recorder["display"] = req.RecorderDisplay
		}
		payload.Set("recorder", recorder)
	}

	return payload
}
