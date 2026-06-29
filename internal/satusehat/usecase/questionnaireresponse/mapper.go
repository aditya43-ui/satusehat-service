package questionnaireresponse

import (
	"time"

	"service/internal/interfaces/satusehat"
)

func MapRequestToFHIR(req QuestionnaireResponseRequest) satusehat.FHIRPayload {
	payload := satusehat.NewFHIRPayload("QuestionnaireResponse")

	if req.OrganizationID != "" && req.QuestionnaireResponseID != "" {
		payload.Set("identifier", map[string]interface{}{
			"system": "http://sys-ids.kemkes.go.id/QuestionnaireResponse/" + req.OrganizationID,
			"value":  req.QuestionnaireResponseID,
		})
	}

	if req.QuestionnaireURL != "" {
		payload.Set("questionnaire", req.QuestionnaireURL)
	}
	if req.Status != "" {
		payload.Set("status", req.Status)
	}

	if req.PatientID != "" {
		subject := map[string]interface{}{"reference": "Patient/" + req.PatientID}
		if req.PatientName != "" {
			subject["display"] = req.PatientName
		}
		payload.Set("subject", subject)
	}

	if req.EncounterID != "" {
		encounter := map[string]interface{}{"reference": "Encounter/" + req.EncounterID}
		if req.EncounterDisplay != "" {
			encounter["display"] = req.EncounterDisplay
		}
		payload.Set("encounter", encounter)
	}

	if req.Authored != nil && !req.Authored.IsZero() {
		payload.Set("authored", req.Authored.Format(time.RFC3339))
	}

	if req.AuthorID != "" {
		authorType := req.AuthorType
		if authorType == "" {
			authorType = "Practitioner"
		}
		author := map[string]interface{}{"reference": authorType + "/" + req.AuthorID}
		if req.AuthorName != "" {
			author["display"] = req.AuthorName
		}
		payload.Set("author", author)
	}

	if req.SourceID != "" {
		sourceType := req.SourceType
		if sourceType == "" {
			sourceType = "Patient"
		}
		source := map[string]interface{}{"reference": sourceType + "/" + req.SourceID}
		if req.SourceName != "" {
			source["display"] = req.SourceName
		}
		payload.Set("source", source)
	}

	if len(req.Items) > 0 {
		items := make([]map[string]interface{}, 0, len(req.Items))
		for _, it := range req.Items {
			items = append(items, mapItem(it))
		}
		payload.Set("item", items)
	}

	return payload
}

func mapItem(it ItemDTO) map[string]interface{} {
	out := map[string]interface{}{"linkId": it.LinkID}
	if it.Text != "" {
		out["text"] = it.Text
	}
	if answer := mapAnswer(it); answer != nil {
		out["answer"] = []map[string]interface{}{answer}
	}
	return out
}

func mapAnswer(it ItemDTO) map[string]interface{} {
	switch {
	case it.AnswerBoolean != nil:
		return map[string]interface{}{"valueBoolean": *it.AnswerBoolean}
	case it.AnswerString != "":
		return map[string]interface{}{"valueString": it.AnswerString}
	case it.AnswerInteger != nil:
		return map[string]interface{}{"valueInteger": *it.AnswerInteger}
	case it.AnswerDecimal != nil:
		return map[string]interface{}{"valueDecimal": *it.AnswerDecimal}
	case it.AnswerDate != nil && !it.AnswerDate.IsZero():
		return map[string]interface{}{"valueDate": it.AnswerDate.Format("2006-01-02")}
	case it.AnswerDateTime != nil && !it.AnswerDateTime.IsZero():
		return map[string]interface{}{"valueDateTime": it.AnswerDateTime.Format(time.RFC3339)}
	case it.AnswerQuantityValue != nil:
		q := map[string]interface{}{"value": *it.AnswerQuantityValue}
		if it.AnswerQuantityUnit != "" {
			q["unit"] = it.AnswerQuantityUnit
		}
		if it.AnswerQuantitySystem != "" {
			q["system"] = it.AnswerQuantitySystem
		}
		if it.AnswerQuantityCode != "" {
			q["code"] = it.AnswerQuantityCode
		}
		return map[string]interface{}{"valueQuantity": q}
	case it.AnswerCodingCode != "" || it.AnswerCodingDisplay != "":
		coding := map[string]interface{}{}
		if it.AnswerCodingSystem != "" {
			coding["system"] = it.AnswerCodingSystem
		}
		if it.AnswerCodingCode != "" {
			coding["code"] = it.AnswerCodingCode
		}
		if it.AnswerCodingDisplay != "" {
			coding["display"] = it.AnswerCodingDisplay
		}
		return map[string]interface{}{"valueCoding": coding}
	}
	return nil
}
