# 📚 Dokumentasi Lengkap API Satu Sehat (GoPrint Builder)

**Base URL**: `http://localhost:8080/api/v1/satusehat` *(sesuaikan dengan host dan port environment Anda)*
**Otorisasi**: Membutuhkan Header `Authorization: Bearer <token_internal_aplikasi_anda>`

Dokumen ini memuat panduan lengkap penggunaan endpoint dan struktur **JSON Body (POST)** untuk modul referensi (Master Data) maupun modul usecase (Transaksional Klinis). Aplikasi GoPrint bertindak sebagai *proxy builder* yang akan mengonversi struktur JSON datar (*flat*) Anda menjadi format HL7 FHIR standar Kemenkes.

---

## 🔐 1. Otorisasi (Auth) Kemenkes

Aplikasi otomatis mengurus *Access Token* ke Kemenkes. Namun jika dibutuhkan secara manual, Anda bisa menggunakan:

- `GET /reference/auth/token` : Mendapatkan token aktif dari *cache* (digunakan oleh API internal).
- `POST /reference/auth/token/refresh` : Memaksa (*bypass cache*) untuk mengambil token baru.

---

## 🏢 2. Modul Reference (Master Data)

Modul ini digunakan untuk meregistrasi entitas dasar. URL menggunakan prefix `/reference/{modul}`.

### A. Organization (Faskes / Departemen)
`POST /reference/organization`
```json
{
  "active": true,
  "type_code": "prov",
  "type_display": "Healthcare Provider",
  "identifier_system": "http://sys-ids.kemkes.go.id/organization/10000004",
  "identifier_value": "10000004",
  "name": "RSUP Dr. Cipto Mangunkusumo",
  "phone": "021-1500135",
  "email": "info@rscm.co.id",
  "url": "https://www.rscm.co.id/",
  "address": "Jl. Diponegoro No.71, Kenari, Senen",
  "city": "Jakarta Pusat",
  "postal_code": "10430",
  "country_code": "ID",
  "part_of_id": "1000000"
}
```

### B. Location (Lokasi / Ruangan)
`POST /reference/location`
```json
{
  "identifier_system": "http://sys-ids.kemkes.go.id/location/10000004",
  "identifier_value": "LOK-001",
  "status": "active",
  "name": "Poli Penyakit Dalam",
  "description": "Poliklinik Penyakit Dalam Gedung A",
  "physical_type_code": "ro",
  "physical_type_display": "Room",
  "managing_organization_id": "10000004",
  "part_of_id": "b017aa54-f1df-4ec2-9d84-8823815d7228"
}
```

### C. Patient (Pasien)
`POST /reference/patient`
```json
{
  "name": "BUDI SANTOSO",
  "nik": "3573012345678901",
  "ihs_number": "P0123456789",
  "gender": "male",
  "birth_date": "1990-12-31",
  "phone": "08123456789",
  "email": "budi.santoso@example.com",
  "is_active": true
}
```

### D. Practitioner (Tenaga Medis)
`POST /reference/practitioner`
```json
{
  "nik": "3301234567890123",
  "name": "Dr. Siti Aminah, Sp.PD",
  "gender": "female",
  "birth_date": "1985-05-15",
  "is_active": true
}
```

---

## 🩺 3. Modul Usecase (Data Klinis)

Semua endpoint ini langsung menggunakan nama resource, contoh: `POST /encounter` atau `POST /condition`.

### 1. Encounter (Kunjungan)
Mencatat registrasi/kedatangan pasien ke faskes.
`POST /encounter`
```json
{
  "encounter_id": "KUNJ-001",
  "organization_id": "10000004",
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "practitioner_id": "N10000001",
  "practitioner_name": "Dr. Siti Aminah",
  "location_id": "b017aa54-f1df-4ec2-9d84-8823815d7228",
  "location_name": "Poli Umum",
  "status": "arrived",
  "class": "AMB",
  "period_start": "2023-10-12T08:00:00+07:00"
}
```
*(Status valid: `planned`, `arrived`, `in-progress`, `finished`, `cancelled`)*

### 2. Condition (Diagnosis / Keluhan)
Mencatat hasil diagnosis ICD-10 pasien.
`POST /condition`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "encounter_id": "5fa23d13-0943-4318-b295-eb1ecfa7384a",
  "clinical_status": "active",
  "category_code": "encounter-diagnosis",
  "category_display": "Encounter Diagnosis",
  "code": "J06.9",
  "display": "Acute upper respiratory infection, unspecified",
  "onset_date_time": "2023-10-12T08:15:00+07:00",
  "recorded_date": "2023-10-12T08:20:00+07:00"
}
```

### 3. Observation (Tanda Vital / Fisik)
Mencatat TTV seperti Nadi, Tensi, Suhu, dll (Standar LOINC).
`POST /observation`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "encounter_id": "5fa23d13-0943-4318-b295-eb1ecfa7384a",
  "status": "final",
  "category_code": "vital-signs",
  "category_display": "Vital Signs",
  "code": "8867-4",
  "display": "Heart rate",
  "value": 85,
  "unit": "beats/minute",
  "unit_code": "/min",
  "effective_date_time": "2023-10-12T08:10:00+07:00"
}
```

### 4. Procedure (Tindakan Medis / ICD-9 CM)
Mencatat tindakan medis atau operasi yang dilakukan ke pasien.
`POST /procedure`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "encounter_id": "5fa23d13-0943-4318-b295-eb1ecfa7384a",
  "status": "completed",
  "category_code": "387713003",
  "category_display": "Surgical procedure",
  "code": "373632001",
  "display": "Minor surgery",
  "performed_date_time": "2023-10-12T09:30:00+07:00",
  "practitioner_id": "N10000001",
  "practitioner_name": "Dr. Siti Aminah"
}
```

### 5. AllergyIntolerance (Alergi)
Mencatat alergi pasien terhadap obat, makanan, atau lingkungan.
`POST /allergyintolerance`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "encounter_id": "5fa23d13-0943-4318-b295-eb1ecfa7384a",
  "clinical_status": "active",
  "verification_status": "confirmed",
  "category": "medication",
  "code": "373410000",
  "display": "Allergy to Penicillin",
  "criticality": "high",
  "recorded_date": "2023-10-12T08:10:00+07:00"
}
```

### 6. ClinicalImpression (Kesan Klinis / Triase)
Mencatat hasil asesmen klinis awal atau triase.
`POST /clinicalimpression`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "encounter_id": "5fa23d13-0943-4318-b295-eb1ecfa7384a",
  "status": "completed",
  "description": "Pasien tampak sadar penuh, tidak ada tanda kegawatdaruratan",
  "summary": "Observasi ringan",
  "practitioner_id": "N10000001",
  "practitioner_name": "Dr. Siti Aminah",
  "effective_date_time": "2023-10-12T08:15:00+07:00"
}
```

### 7. Medication (Master Obat per Kunjungan)
Mendefinisikan obat berdasarkan KFA (Kamus Farmasi Alat Kesehatan).
`POST /medication`
```json
{
  "status_code": "active",
  "kfa_code": "93000940",
  "kfa_display": "Paracetamol 500 mg Tablet",
  "form_code": "BS019",
  "form_display": "Tablet",
  "manufacturer_id": "10000004",
  "batch_number": "BATCH12345",
  "expiration_date": "2025-12-31T00:00:00Z"
}
```

### 8. MedicationRequest (Resep Obat)
Permintaan/resep obat dari Dokter ke Apotek.
`POST /medicationrequest`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "encounter_id": "5fa23d13-0943-4318-b295-eb1ecfa7384a",
  "status": "active",
  "intent": "order",
  "medication_id": "medication-uuid-from-kfa",
  "medication_display": "Paracetamol 500 mg Tablet",
  "practitioner_id": "N10000001",
  "authored_on": "2023-10-12T08:45:00+07:00",
  "dosage_instruction": "Diminum 3 kali sehari setelah makan",
  "dispense_quantity": 10
}
```

### 9. MedicationDispense (Penyerahan Obat / Apotek)
Pencatatan penyerahan obat fisik kepada pasien oleh Apoteker.
`POST /medicationdispense`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "encounter_id": "5fa23d13-0943-4318-b295-eb1ecfa7384a",
  "medication_request_id": "request-uuid-here",
  "status": "completed",
  "medication_id": "medication-uuid-from-kfa",
  "medication_display": "Paracetamol 500 mg Tablet",
  "practitioner_id": "N10000002",
  "practitioner_name": "Apt. Budi, S.Farm",
  "handed_over_date": "2023-10-12T09:00:00+07:00",
  "quantity": 10
}
```

### 10. MedicationStatement (Pernyataan Penggunaan Obat)
Riwayat penggunaan obat yang dilaporkan oleh pasien.
`POST /medicationstatement`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "encounter_id": "5fa23d13-0943-4318-b295-eb1ecfa7384a",
  "status": "active",
  "medication_id": "medication-uuid-from-kfa",
  "medication_display": "Paracetamol 500 mg Tablet",
  "effective_date_time": "2023-10-12T08:00:00+07:00",
  "dosage_text": "Sering minum paracetamol bebas saat pusing"
}
```

### 11. ServiceRequest (Permintaan Layanan / Lab / Rujukan)
Permintaan tindakan lebih lanjut, seperti rujuk poli, rawat inap, atau periksa Lab/Radiologi.
`POST /servicerequest`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "encounter_id": "5fa23d13-0943-4318-b295-eb1ecfa7384a",
  "status": "active",
  "intent": "order",
  "category_code": "108252007",
  "category_display": "Laboratory procedure",
  "code": "1044-2",
  "display": "Darah Rutin",
  "requester_id": "N10000001",
  "authored_on": "2023-10-12T08:30:00+07:00"
}
```

### 12. DiagnosticReport (Hasil Lab / Radiologi)
Laporan hasil pengecekan Lab atau Radiologi (penggabungan berbagai Observation).
`POST /diagnosticreport`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "encounter_id": "5fa23d13-0943-4318-b295-eb1ecfa7384a",
  "status": "final",
  "category_code": "LAB",
  "category_display": "Laboratory",
  "code": "1044-2",
  "display": "Hasil Darah Rutin",
  "practitioner_id": "N10000001",
  "issued_date": "2023-10-12T11:00:00+07:00",
  "conclusion": "Semua indikator darah dalam batas normal."
}
```

### 13. Specimen (Sampel Lab)
Data pencatatan spesimen (misal: darah, urin) yang diambil dari tubuh pasien.
`POST /specimen`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "type_code": "119297000",
  "type_display": "Blood specimen",
  "collection_date_time": "2023-10-12T10:00:00+07:00",
  "collector_id": "N10000001",
  "received_date_time": "2023-10-12T10:15:00+07:00"
}
```

### 14. ImagingStudy (Pencitraan Radiologi)
Data meta terkait pemeriksaan Radiologi (X-Ray, MRI, CT-Scan).
`POST /imagingstudy`
```json
{
  "organization_id": "10000004",
  "accession_number": "USG-SEN25-011815",
  "service_request_id": "c8d9f768-2237-4da5-b0ad-9cd5e6087a9a",
  "patient_id": "P20395354720",
  "patient_name": "Budi Santoso",
  "encounter_id": "5fa23d13-0943-4318-b295-eb1ecfa7384a",
  "status": "available",
  "started": "2025-06-10T11:41:46+07:00",
  "number_of_series": 1,
  "number_of_instances": 5,
  "procedure_code": "CT",
  "procedure_display": "CT Scan",
  "description": "Keterangan hasil pemeriksaan"
}
```

### 15. Immunization (Riwayat Vaksin)
Pencatatan pemberian imunisasi/vaksinasi.
`POST /immunization`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "encounter_id": "5fa23d13-0943-4318-b295-eb1ecfa7384a",
  "status": "completed",
  "vaccine_code": "J07BM01",
  "vaccine_display": "COVID-19 Vaccine",
  "occurrence_date_time": "2023-10-12T09:00:00+07:00",
  "practitioner_id": "N10000001",
  "location_id": "b017aa54-f1df-4ec2-9d84-8823815d7228",
  "dose_quantity": 0.5,
  "dose_unit": "ml"
}
```

### 16. CarePlan (Rencana Perawatan)
Rencana medis/keperawatan masa depan untuk pasien.
`POST /careplan`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "encounter_id": "5fa23d13-0943-4318-b295-eb1ecfa7384a",
  "status": "active",
  "intent": "plan",
  "title": "Rencana Diet Diabetes",
  "description": "Pasien harus mematuhi diet rendah gula dan karbohidrat.",
  "practitioner_id": "N10000001",
  "created_date": "2023-10-12T10:00:00+07:00"
}
```

### 17. EpisodeOfCare (Episode Perawatan)
Mengelompokkan serangkaian *Encounter* (kunjungan) yang merujuk pada satu kondisi klinis / episode penyakit yang sama.
`POST /episodeofcare`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "organization_id": "10000004",
  "status": "active",
  "period_start": "2023-10-01T08:00:00+07:00",
  "managing_practitioner_id": "N10000001"
}
```

### 18. Composition (Resume Medis / Dokumen Klinis)
Menggabungkan berbagai resource (Diagnosis, Obat, Tindakan) menjadi satu resume / dokumen medis yang utuh secara hukum.
`POST /composition`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "encounter_id": "5fa23d13-0943-4318-b295-eb1ecfa7384a",
  "status": "final",
  "title": "Resume Pulang Pasien Rawat Inap",
  "date": "2023-10-15T12:00:00Z",
  "practitioner_id": "N10000001",
  "practitioner_name": "Dr. Siti Aminah"
}
```

### 19. QuestionnaireResponse (Hasil Kuesioner Medis)
Jawaban atas asesmen spesifik (misal: Skrining Nyeri, Skrining Jatuh, Tumbuh Kembang).
`POST /questionnaireresponse`
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "encounter_id": "5fa23d13-0943-4318-b295-eb1ecfa7384a",
  "status": "completed",
  "questionnaire_url": "http://sys-ids.kemkes.go.id/questionnaire/Nyeri",
  "authored_date": "2023-10-12T08:05:00+07:00",
  "author_practitioner_id": "N10000002",
  "items": [
    {
      "linkId": "1.1",
      "text": "Tingkat Nyeri (0-10)",
      "answer_value_integer": 3
    }
  ]
}
```

---

## 💡 4. Standar Method GET, PUT, & PATCH

Semua modul mendukung operasi standar berikut:

1. **`GET /{modul}`** (Search)
   Melakukan pencarian. Parameter URL (*Query String*) disesuaikan per modul (misal `?patient=P012...` atau `?name=Budi`).
   
2. **`GET /{modul}/{id}`** (Get By ID)
   Mengambil detail data utuh (dalam format FHIR) berdasarkan ID kemenkes.

3. **`PUT /{modul}/{id}`** (Replace)
   Mengirimkan *payload JSON* penuh persis seperti `POST` untuk me-replace (mengganti) dokumen lama.

4. **`PATCH /{modul}/{id}`** (Partial Update)
   Menggunakan standar *JSON Patch (RFC 6902)* untuk mengubah satu/beberapa *field* spesifik tanpa mengubah keseluruhan dokumen.
   **Contoh Request Body (Array of Objects):**
   ```json
   [
     {
       "op": "replace",
       "path": "/status",
       "value": "finished"
     }
   ]
   ```