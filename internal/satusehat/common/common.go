package common

// ReferenceDTO merepresentasikan referensi standar ke resource FHIR lain.
type ReferenceDTO struct {
	Reference string `json:"reference,omitempty"`
	Display   string `json:"display,omitempty"`
}

// CodeableConceptDTO merepresentasikan konsep kode HL7/SNOMED/LOINC.
type CodeableConceptDTO struct {
	System  string `json:"system,omitempty"`
	Code    string `json:"code,omitempty"`
	Display string `json:"display,omitempty"`
	Text    string `json:"text,omitempty"`
}

// QuantityDTO merepresentasikan besaran nilai beserta unit pengukurannya.
type QuantityDTO struct {
	Value  float64 `json:"value,omitempty"`
	Unit   string  `json:"unit,omitempty"`
	System string  `json:"system,omitempty"`
	Code   string  `json:"code,omitempty"`
}
