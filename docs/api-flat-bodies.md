# Flat JSON Bodies — All SatuSehat Use Cases

> Flat (non-nested) POST body for **every** FHIR resource currently
> exposed by `service-satusehat`.
>
> **Status:** ✅ **Implemented**. As of 2026-05-13, all 19 FHIR usecases +
> EpisodeOfCare + QuestionnaireResponse accept the flat shape below. Internal
> mappers compose the nested FHIR R4 payload before submission to SATUSEHAT.
> No `*DTO` types from `internal/satusehat/common` are exposed at the API
> surface anymore.
>
> Conventions used in every body below:
>
> - `*_id` carries the **raw identifier**; the mapper composes
>   `Patient/{id}`, `Practitioner/{id}`, `Encounter/{id}`, etc.
> - `*_name` (or `*_display`) is the human-readable display.
> - `*_code` + `*_display` (+ optional `*_system`) replace nested
>   `CodeableConceptDTO`.
> - `*_value` + `*_unit` (+ optional `*_system`, `*_code`) replace nested
>   `QuantityDTO`.
> - Date-times use ISO-8601 `2026-05-13T08:00:00+07:00` (RFC 3339).
>   Encounter alone accepts the simpler `YYYY-MM-DD HH:MM:SS` form for
>   `period_start` / `period_end` via the project's `CustomTime` (WIB).
> - Status / class / intent values match the validator `oneof=` lists in the
>   DTOs verbatim.
> - `organization_id` is optional — if omitted, the service injects
>   `cfg.SatuSehat.OrgID` at the boundary.

---

## 1. Encounter

`POST /satusehat/encounter`

```json
{
  "encounter_id": "ENC-0012345",
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "patient_name": "Budi Santoso",
  "practitioner_id": "N10000001",
  "practitioner_name": "Dr. Andi",
  "location_id": "a6bab5d0-ba3c-4f73-8450-f44d6ca8e9d4",
  "location_name": "Poli Umum",
  "status": "arrived",
  "class": "AMB",
  "period_start": "2026-05-13 08:00:00",
  "period_end": "2026-05-13 09:30:00",
  "diagnosis_condition_id": "C-0001",
  "diagnosis_use_code": "AD",
  "diagnosis_use_display": "Admission diagnosis",
  "diagnosis_rank": 1
}
```

---

## 2. EpisodeOfCare

`POST /satusehat/episodeofcare`

```json
{
  "episode_of_care_id": "EOC-12345",
  "organization_id": "10000004",
  "managing_organization_id": "10000004",
  "patient_id": "P03647103112",
  "patient_name": "Budi Santoso",
  "care_manager_id": "N10000001",
  "care_manager_name": "Dr. Andi",
  "status": "active",
  "type_code": "HACC",
  "type_display": "Home and Community Care",
  "period_start": "2026-05-13T08:00:00+07:00",
  "period_end": "2026-08-13T17:00:00+07:00",
  "diagnosis_condition_id": "C-0001",
  "diagnosis_role_code": "CC",
  "diagnosis_role_display": "Chief complaint",
  "diagnosis_rank": 1
}
```

---

## 3. Condition

`POST /satusehat/condition`

```json
{
  "condition_id": "COND-9001",
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "patient_name": "Budi Santoso",
  "encounter_id": "ENC-0012345",
  "clinical_status": "active",
  "category_code": "encounter-diagnosis",
  "category_display": "Encounter Diagnosis",
  "code_system": "http://hl7.org/fhir/sid/icd-10",
  "code": "J06.9",
  "code_display": "Acute upper respiratory infection, unspecified",
  "onset_date_time": "2026-05-13T08:15:00+07:00",
  "recorded_date": "2026-05-13T08:20:00+07:00"
}
```

---

## 4. Observation

`POST /satusehat/observation`

```json
{
  "observation_id": "OBS-7001",
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "encounter_id": "ENC-0012345",
  "performer_id": "N10000001",
  "performer_name": "Dr. Andi",
  "status": "final",
  "category_code": "vital-signs",
  "category_display": "Vital Signs",
  "code_system": "http://loinc.org",
  "code": "8867-4",
  "code_display": "Heart rate",
  "effective_datetime": "2026-05-13T08:10:00+07:00",
  "issued": "2026-05-13T08:12:00+07:00",
  "value_quantity_value": 80,
  "value_quantity_unit": "beats/minute",
  "value_quantity_system": "http://unitsofmeasure.org",
  "value_quantity_code": "/min",
  "body_site_code": "40983000",
  "body_site_display": "Arm",
  "interpretation_code": "N",
  "interpretation_display": "Normal",
  "reference_range_low_value": 60,
  "reference_range_high_value": 100,
  "reference_range_unit": "beats/minute",
  "reference_range_text": "Normal adult resting heart rate"
}
```

---

## 5. AllergyIntolerance

`POST /satusehat/allergyintolerance`

```json
{
  "allergy_id": "ALG-0001",
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "patient_name": "Budi Santoso",
  "encounter_id": "ENC-0012345",
  "encounter_display": "Poli Umum 13 Mei 2026",
  "recorder_id": "N10000001",
  "recorder_display": "Dr. Andi",
  "clinical_status": "active",
  "verification_status": "confirmed",
  "category": "medication",
  "code_system": "http://snomed.info/sct",
  "code": "294505008",
  "code_display": "Allergy to amoxicillin",
  "recorded_date": "2026-05-13T08:25:00+07:00"
}
```

---

## 6. CarePlan

`POST /satusehat/careplan`

```json
{
  "care_plan_id": "CP-0001",
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "patient_display": "Budi Santoso",
  "encounter_id": "ENC-0012345",
  "encounter_display": "Poli Umum 13 Mei 2026",
  "author_id": "N10000001",
  "author_display": "Dr. Andi",
  "status": "active",
  "intent": "plan",
  "category_code": "assess-plan",
  "category_display": "Assessment and Plan of Treatment",
  "title": "Rencana perawatan ISPA",
  "description": "Antibiotik 5 hari, kontrol H+3",
  "created_date": "2026-05-13T08:30:00+07:00",
  "goal_ids": ["GOAL-001", "GOAL-002"]
}
```

---

## 7. ClinicalImpression

`POST /satusehat/clinicalimpression`

```json
{
  "clinical_impression_id": "CI-0001",
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "patient_display": "Budi Santoso",
  "encounter_id": "ENC-0012345",
  "encounter_display": "Poli Umum 13 Mei 2026",
  "assessor_id": "N10000001",
  "assessor_display": "Dr. Andi",
  "status": "completed",
  "code_system": "http://snomed.info/sct",
  "code": "162673000",
  "code_display": "General examination of patient",
  "description": "Pasien sadar, demam, batuk produktif",
  "effective_datetime": "2026-05-13T08:35:00+07:00",
  "date": "2026-05-13T08:40:00+07:00",
  "summary": "ISPA non-pneumonia",
  "problem_condition_ids": ["COND-9001"],
  "finding_code": "386661006",
  "finding_display": "Fever",
  "prognosis_code": "170968001",
  "prognosis_display": "Prognosis good"
}
```

---

## 8. Composition

`POST /satusehat/composition`

```json
{
  "composition_id": "COMP-0001",
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "patient_display": "Budi Santoso",
  "encounter_id": "ENC-0012345",
  "encounter_display": "Poli Umum 13 Mei 2026",
  "author_id": "N10000001",
  "author_display": "Dr. Andi",
  "status": "final",
  "type_system": "http://loinc.org",
  "type_code": "11488-4",
  "type_display": "Consult note",
  "category_code": "LP173421-1",
  "category_display": "Report",
  "title": "Catatan Konsultasi Poli Umum",
  "date": "2026-05-13T09:00:00+07:00",
  "section_title": "Anamnesis & Pemeriksaan",
  "section_code": "55109-3",
  "section_display": "Reason for visit Narrative",
  "section_text": "Pasien datang dengan keluhan batuk dan demam selama 3 hari."
}
```

---

## 9. DiagnosticReport

`POST /satusehat/diagnosticreport`

```json
{
  "diagnostic_id": "DR-5001",
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "encounter_id": "ENC-0012345",
  "performer_id": "N10000001",
  "performer_name": "Dr. Andi",
  "status": "final",
  "category_code": "LAB",
  "category_display": "Laboratory",
  "code_system": "http://loinc.org",
  "code": "58410-2",
  "code_display": "Complete blood count (hemogram) panel",
  "effective_datetime": "2026-05-13T09:10:00+07:00",
  "issued": "2026-05-13T09:30:00+07:00",
  "result_observation_ids": ["OBS-7001", "OBS-7002"],
  "specimen_ids": ["SPC-3001"],
  "imaging_study_ids": [],
  "conclusion_code": "162673000",
  "conclusion_display": "General examination of patient",
  "conclusion": "Hasil dalam batas normal"
}
```

---

## 10. ImagingStudy

`POST /satusehat/imagingstudy`

```json
{
  "accession_number": "ACC-0001",
  "organization_id": "10000004",
  "service_request_id": "SR-0001",
  "patient_id": "P03647103112",
  "patient_name": "Budi Santoso",
  "encounter_id": "ENC-0012345",
  "practitioner_id": "N10000001",
  "practitioner_name": "Dr. Andi",
  "status": "available",
  "started": "2026-05-13T09:00:00+07:00",
  "number_of_series": 1,
  "number_of_instances": 24,
  "procedure_code": "168731009",
  "procedure_display": "Chest X-ray",
  "description": "Thorax PA"
}
```

---

## 11. Immunization

`POST /satusehat/immunization`

```json
{
  "immunization_id": "IMM-0001",
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "patient_name": "Budi Santoso",
  "encounter_id": "ENC-0012345",
  "encounter_display": "Poli Umum 13 Mei 2026",
  "performer_id": "N10000001",
  "performer_name": "Dr. Andi",
  "location_id": "a6bab5d0-ba3c-4f73-8450-f44d6ca8e9d4",
  "location_name": "Poli Umum",
  "status": "completed",
  "vaccine_code_system": "http://sys-ids.kemkes.go.id/vaccine",
  "vaccine_code": "1010101010",
  "vaccine_display": "Sinovac",
  "occurrence_date_time": "2026-05-13T09:05:00+07:00",
  "primary_source": true,
  "lot_number": "BATCH-XYZ-2026-001",
  "dose_quantity_value": 0.5,
  "dose_quantity_unit": "mL",
  "dose_quantity_system": "http://unitsofmeasure.org",
  "dose_quantity_code": "mL",
  "route_code": "IM",
  "route_display": "Intramuscular"
}
```

---

## 12. Medication

`POST /satusehat/medication`

```json
{
  "medication_id": "MED-0001",
  "organization_id": "10000004",
  "manufacturer_id": "100099999",
  "status_code": "active",
  "kfa_code": "92000001",
  "kfa_display": "Paracetamol 500 mg tablet",
  "form_code": "TAB",
  "form_display": "Tablet",
  "batch_number": "BTC-XYZ-2026-09",
  "expiration_date": "2027-12-31T00:00:00+07:00"
}
```

---

## 13. MedicationDispense

`POST /satusehat/medicationdispense`

```json
{
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "patient_name": "Budi Santoso",
  "encounter_id": "ENC-0012345",
  "practitioner_id": "N10000001",
  "practitioner_name": "Apt. Sari",
  "location_id": "a6bab5d0-ba3c-4f73-8450-f44d6ca8e9d4",
  "location_name": "Apotek Rawat Jalan",
  "medication_id": "MED-0001",
  "medication_display": "Paracetamol 500 mg tablet",
  "medication_request_id": "MR-0001",
  "prescription_id": "RX-2026-000123",
  "prescription_item_id": "RX-2026-000123-1",
  "status": "completed",
  "category": "outpatient",
  "prepared_date": "2026-05-13 09:40:00",
  "handed_over_date": "2026-05-13 09:45:00",
  "quantity_value": 10,
  "quantity_unit": "TAB",
  "days_supply_value": 5,
  "dosage_text": "1 tablet, 3x sehari sesudah makan",
  "timing_frequency": 3,
  "timing_period": 1,
  "timing_period_unit": "d",
  "dose_quantity_value": 1,
  "dose_quantity_unit": "TAB"
}
```

---

## 14. MedicationRequest

`POST /satusehat/medicationrequest`

```json
{
  "medicationrequest_id": "MR-0001",
  "prescription_item_id": "RX-2026-000123-1",
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "patient_name": "Budi Santoso",
  "encounter_id": "ENC-0012345",
  "practitioner_id": "N10000001",
  "practitioner_name": "Dr. Andi",
  "medication_id": "MED-0001",
  "medication_display": "Paracetamol 500 mg tablet",
  "category": "outpatient",
  "priority": "routine",
  "status": "active",
  "intent": "order",
  "authored_on": "2026-05-13T09:35:00+07:00",
  "reason_code": "J06.9",
  "reason_display": "Acute upper respiratory infection",
  "course_of_therapy_code": "acute",
  "course_of_therapy_display": "Short course (acute) therapy",
  "dosage_text": "1 tablet, 3x sehari sesudah makan",
  "additional_instruction": "Habiskan",
  "patient_instruction": "Minum dengan air putih",
  "timing_frequency": 3,
  "timing_period": 1,
  "timing_period_unit": "d",
  "route_code": "O",
  "route_display": "Oral",
  "dose_quantity_value": 1,
  "dose_quantity_unit": "TAB",
  "dispense_interval": 0,
  "dispense_value": 15,
  "dispense_unit": "TAB",
  "supply_duration": 5,
  "validity_period_start": "2026-05-13T09:35:00+07:00",
  "validity_period_end": "2026-05-20T23:59:59+07:00"
}
```

---

## 15. MedicationStatement

`POST /satusehat/medicationstatement`

```json
{
  "statement_id": "MS-0001",
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "patient_name": "Budi Santoso",
  "encounter_id": "ENC-0012345",
  "information_source_id": "P03647103112",
  "information_source_name": "Budi Santoso",
  "status": "active",
  "category_code": "outpatient",
  "category_display": "Outpatient",
  "medication_code_system": "http://sys-ids.kemkes.go.id/kfa",
  "medication_code": "92000001",
  "medication_display": "Paracetamol 500 mg tablet",
  "effective_date_time": "2026-05-13T09:35:00+07:00",
  "date_asserted": "2026-05-13T09:36:00+07:00",
  "dosage_text": "1 tablet, 3x sehari sesudah makan",
  "dosage_patient_instruction": "Minum dengan air putih",
  "dosage_route_code": "O",
  "dosage_route_display": "Oral",
  "dose_quantity_value": 1,
  "dose_quantity_unit": "TAB"
}
```

---

## 16. Procedure

`POST /satusehat/procedure`

```json
{
  "procedure_id": "PROC-0001",
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "encounter_id": "ENC-0012345",
  "performer_id": "N10000001",
  "performer_name": "Dr. Andi",
  "status": "completed",
  "category_code": "103693007",
  "category_display": "Diagnostic procedure",
  "code_system": "http://hl7.org/fhir/sid/icd-9-cm",
  "code": "89.61",
  "code_display": "Continuous blood gas monitoring",
  "performed_date_time": "2026-05-13T09:15:00+07:00",
  "performed_start": "2026-05-13T09:15:00+07:00",
  "performed_end": "2026-05-13T09:25:00+07:00",
  "reason_code": "J06.9",
  "reason_display": "Acute upper respiratory infection",
  "body_site_code": "302551006",
  "body_site_display": "Entire thorax",
  "note": "Pasien kooperatif, prosedur selesai tanpa komplikasi"
}
```

---

## 17. QuestionnaireResponse

`POST /satusehat/questionnaireresponse`

```json
{
  "questionnaire_response_id": "QR-0001",
  "organization_id": "10000004",
  "questionnaire_url": "https://fhir.kemkes.go.id/Questionnaire/Q0007",
  "patient_id": "P03647103112",
  "encounter_id": "ENC-0012345",
  "author_id": "N10000001",
  "author_name": "Dr. Andi",
  "source_id": "P03647103112",
  "source_name": "Budi Santoso",
  "status": "completed",
  "authored": "2026-05-13T09:00:00+07:00",
  "items": [
    {
      "link_id": "1",
      "text": "Apakah Anda merokok?",
      "answer_boolean": false
    },
    {
      "link_id": "2",
      "text": "Berapa suhu tubuh Anda hari ini?",
      "answer_quantity_value": 38.2,
      "answer_quantity_unit": "Cel"
    }
  ]
}
```

> Note: `items` remains an array because QuestionnaireResponse is
> intrinsically list-shaped — but each item is itself flat.

---

## 18. ServiceRequest

`POST /satusehat/servicerequest`

```json
{
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "patient_name": "Budi Santoso",
  "encounter_id": "ENC-0012345",
  "requester_id": "N10000001",
  "requester_name": "Dr. Andi",
  "performer_id": "N10000099",
  "performer_name": "Lab Pathology Unit",
  "status": "active",
  "intent": "order",
  "code": "58410-2",
  "display": "Complete blood count panel",
  "authored_on": "2026-05-13T09:05:00+07:00"
}
```

---

## 19. Specimen

`POST /satusehat/specimen`

```json
{
  "specimen_id": "SPC-3001",
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "status": "available",
  "type_system": "http://terminology.hl7.org/CodeSystem/v2-0487",
  "type_code": "BLD",
  "type_display": "Whole blood",
  "received_date_time": "2026-05-13T09:20:00+07:00",
  "collected_date_time": "2026-05-13T09:10:00+07:00",
  "collector_id": "N10000001",
  "collector_name": "Perawat Sari",
  "collection_quantity_value": 5,
  "collection_quantity_unit": "mL",
  "collection_method_code": "129316008",
  "collection_method_display": "Aspiration - action",
  "body_site_code": "368208006",
  "body_site_display": "Left upper arm",
  "fasting_status_code": "F",
  "fasting_status_display": "Fasting",
  "processing_procedure_code": "9718006",
  "processing_procedure_display": "Centrifugation",
  "processing_time_datetime": "2026-05-13T09:25:00+07:00",
  "conditions": ["refrigerated"],
  "request_service_request_ids": ["SR-0001"]
}
```

---

## 20. Studies (DICOM upload)

The `studies` module today is a thin DICOM tarball uploader and does not
take a structured body. The recommended request is `multipart/form-data`:

```http
POST /satusehat/dicom/studies/upload
Content-Type: multipart/form-data; boundary=----WebKitFormBoundaryXyz

------WebKitFormBoundaryXyz
Content-Disposition: form-data; name="organization_id"

10000004
------WebKitFormBoundaryXyz
Content-Disposition: form-data; name="patient_id"

P03647103112
------WebKitFormBoundaryXyz
Content-Disposition: form-data; name="accession_number"

ACC-0001
------WebKitFormBoundaryXyz
Content-Disposition: form-data; name="file"; filename="study.tar.gz"
Content-Type: application/gzip

<binary DICOM tarball>
------WebKitFormBoundaryXyz--
```

If a JSON metadata-only endpoint is added later, the flat shape would be:

```json
{
  "organization_id": "10000004",
  "patient_id": "P03647103112",
  "accession_number": "ACC-0001",
  "study_instance_uid": "1.2.840.113619.2.5.1762583153.1762583153.1762583153.1",
  "modality": "CR",
  "started": "2026-05-13T09:00:00+07:00",
  "description": "Thorax PA"
}
```

---

## Implementation notes

All 20 usecases now share the same internal pattern:

1. **`dto.go`** — pure flat struct with `binding:"required,..."` tags only on
   scalar fields. No `common.ReferenceDTO` / `common.CodeableConceptDTO`
   imports at the request boundary.
2. **`mapper.go`** — `MapRequestToFHIR(req)` builds the nested FHIR R4
   payload. Mapper is pure: no env reads, no I/O.
3. **`repository.go`** — thin wrapper over `SatuSehatClient.DoRequest`. The
   old `hiddenCtx` wrapper has been removed; `context.Context` flows through
   unchanged for proper deadline / cancellation propagation.
4. **`service.go`** — `NewService(repo, orgID string)` takes the default
   organisation identifier from `cfg.SatuSehat.OrgID` at boot. When a caller
   omits `organization_id` in the request body, the service fills it in.

The full audit trail is in `docs/DEVLOG.md`.

## Field-shape cheat-sheet (nested → flat)

| Nested FHIR field | Flat replacement |
| --- | --- |
| `subject.reference = "Patient/X"` | `patient_id = "X"` (+ optional `patient_name`) |
| `encounter.reference = "Encounter/X"` | `encounter_id = "X"` (+ `encounter_display`) |
| `performer[].reference = "Practitioner/X"` | `performer_id = "X"` (+ `performer_name`) |
| `location[].location.reference = "Location/X"` | `location_id = "X"` (+ `location_name`) |
| `managingOrganization.reference = "Organization/X"` | `managing_organization_id = "X"` |
| `code.system / .code / .display` | `code_system / code / code_display` |
| `category.code / .display` | `category_code / category_display` |
| `valueQuantity.value / .unit / .system / .code` | `value_quantity_value / _unit / _system / _code` |
| `period.start / .end` | `period_start / period_end` |
| `dosage.route.code` | `route_code` / `dosage_route_code` |

When a list of references exists (e.g. `result_observation_ids`,
`specimen_ids`), keep it as a flat array of IDs — the mapper turns them
into `["Observation/{id}", …]`.
