# 📚 Dokumentasi API Satu Sehat (Internal GoPrint)

**Base URL**: `http://localhost:8080/api/v1/satusehat` *(sesuaikan port dengan environment Anda)*
**Otorisasi**: Membutuhkan Header `Authorization: Bearer <token_internal_aplikasi_anda>`.

---

## 📜 Pendahuluan

Dokumen ini menjelaskan cara menggunakan API internal GoPrint yang berfungsi sebagai *proxy* dan *builder* untuk layanan **Satu Sehat Kemenkes**. API ini menyederhanakan *payload* dan mengelola otorisasi secara otomatis.

### Struktur Endpoint
Setiap modul sumber daya (misal: `Patient`, `Encounter`) memiliki 5 endpoint standar:
1.  `POST /<resource>`: Membuat data baru.
2.  `GET /<resource>?param=value`: Mencari data berdasarkan parameter.
3.  `GET /<resource>/:id`: Mengambil detail data berdasarkan ID Kemenkes.
4.  `PUT /<resource>/:id`: Memperbarui data secara utuh (wajib menyertakan *payload* lengkap).
5.  `PATCH /<resource>/:id`: Memperbarui sebagian data menggunakan format *JSON Patch* (RFC 6902).

### Format Respons Error
Jika terjadi kegagalan, API akan mengembalikan respons dengan format berikut:
```json
{
    "code": 400,
    "status": "error",
    "message": "Pesan error utama yang mudah dibaca",
    "data": "Detail teknis error dari Kemenkes atau validasi internal"
}
```

---

## 🔐 Auth (Autentikasi API Kemenkes)
Modul ini digunakan untuk mendapatkan status kredensial dan token aktif dari server Kemenkes. *(Catatan: Request internal API Anda sebenarnya sudah otomatis dibungkus dengan token oleh middleware atau `SatuSehatClient` internal)*.

#### Mengambil Token Aktif (`GET /auth/token`)
Mendapatkan *Access Token* OAuth2 yang di-*generate* menggunakan *Client ID* dan *Client Secret* dari konfigurasi *environment* aplikasi.
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIsIn...",
  "expires_in": 3599
}
```

---

## � Patient (Pasien)
Mendaftarkan dan mengelola data demografi pasien.

#### Membuat Data (`POST /patient`)
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

#### Mencari Data (`GET /patient`)
Parameter pencarian yang didukung:
- `identifier`: Cari berdasarkan NIK atau IHS Number. Contoh: `?identifier=https://fhir.kemkes.go.id/id/nik|3573012345678901`
- `name`: Cari berdasarkan nama pasien.
- `birthdate`: Cari berdasarkan tanggal lahir (YYYY-MM-DD).
- `gender`: Cari berdasarkan jenis kelamin (`male`, `female`, `other`, `unknown`).

#### Operasi Lainnya
- **Mengambil Data (`GET /patient/:id`)**: Mengambil detail pasien berdasarkan ID IHS yang didapat saat pembuatan.
- **Memperbarui Data (`PUT /patient/:id`)**: Mengirimkan kembali *payload* lengkap seperti `POST` untuk memperbarui data.
- **Memperbarui Sebagian (`PATCH /patient/:id`)**: Menggunakan format JSON Patch untuk mengubah data tertentu. Lihat panduan di akhir dokumen.

---

## 🏢 Organization (Organisasi)
Mengelola data referensi organisasi/faskes.

#### Membuat Data (`POST /organization`)
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
    "address": "Jl. Diponegoro No.71, RW.5, Kenari, Kec. Senen, Kota Jakarta Pusat",
    "city": "Jakarta Pusat",
    "postal_code": "10430",
    "country_code": "ID",
    "part_of_id": "1000000"
}
```

#### Mencari Data (`GET /organization`)
Parameter pencarian yang didukung (wajib salah satu):
- `name`: Cari berdasarkan nama organisasi.
- `partof`: Cari berdasarkan ID organisasi induk.
- `identifier`: Cari berdasarkan identifier unik faskes.

#### Operasi Lainnya
- **Mengambil Data (`GET /organization/:id`)**: Mengambil detail organisasi berdasarkan ID Kemenkes.
- **Memperbarui Data (`PUT /organization/:id`)**: Mengirimkan kembali *payload* lengkap seperti `POST`.
- **Memperbarui Sebagian (`PATCH /organization/:id`)**: Menggunakan format JSON Patch.

---

## 👨‍⚕️ Practitioner (Tenaga Kesehatan)
Mengambil dan mengelola data profil tenaga kesehatan (Dokter, Perawat, Bidan, dll).

#### Membuat Data (`POST /practitioner`)
```json
{
  "nik": "3301234567890123",
  "name": "Dr. Siti Aminah",
  "gender": "female",
  "birth_date": "1985-05-15",
  "is_active": true
}
```

#### Mencari Data (`GET /practitioner`)
Parameter pencarian yang didukung:
- `identifier`: Cari berdasarkan NIK. Contoh: `?identifier=https://fhir.kemkes.go.id/id/nik|3301234567890123`
- `name`: Cari berdasarkan nama tenaga kesehatan.

#### Operasi Lainnya
- **Mengambil Data (`GET /practitioner/:id`)**: Mengambil detail Nakes berdasarkan ID IHS Kemenkes.
- **Memperbarui Data (`PUT /practitioner/:id`)**: Mengirimkan kembali *payload* lengkap seperti `POST`.
- **Memperbarui Sebagian (`PATCH /practitioner/:id`)**: Menggunakan format JSON Patch.

---

## 📍 Location (Lokasi / Ruangan)
Mencatat data referensi lokasi fisik pelayanan di dalam Faskes (misal: Poli, Bangsal, Kamar, Bed).

#### Membuat Data (`POST /location`)
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

#### Mencari Data (`GET /location`)
Parameter pencarian yang didukung:
- `identifier`: Cari berdasarkan identifier unik lokasi internal faskes.
- `name`: Cari berdasarkan nama lokasi.
- `organization`: Cari berdasarkan ID Organisasi (faskes) pengelola.

#### Operasi Lainnya
- **Mengambil Data (`GET /location/:id`)**: Mengambil detail lokasi berdasarkan ID Kemenkes.
- **Memperbarui Data (`PUT /location/:id`)**: Mengirimkan kembali *payload* lengkap seperti `POST`.
- **Memperbarui Sebagian (`PATCH /location/:id`)**: Menggunakan format JSON Patch.

---

## 🔍 KFA (Kamus Farmasi dan Alat Kesehatan)
Pencarian data master obat dan alat kesehatan langsung dari database KFA v2 Kemenkes (terpisah dari base FHIR).

#### Mencari Data Master KFA (`GET /kfa/products`)
Parameter pencarian yang didukung:
- `identifier`: Cari berdasarkan kode KFA produk.
- `keyword`: Cari berdasarkan nama obat / alat kesehatan.
- `product_type`: Filter tipe produk (misal: `farmasi`, `alkes`).

---

## 🏥 Encounter (Kunjungan)
Mencatat kunjungan pasien ke faskes.

#### Membuat Data (`POST /encounter`)
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
*Status yang valid: `planned`, `arrived`, `in-progress`, `finished`, `cancelled`.*

#### Mencari Data (`GET /encounter`)
Parameter pencarian yang didukung:
- `patient`: Berdasarkan ID Pasien.
- `status`: Berdasarkan status kunjungan.
- `date`: Berdasarkan tanggal kunjungan (YYYY-MM-DD).

#### Operasi Lainnya
- **Mengambil Data (`GET /encounter/:id`)**: Mengambil detail kunjungan berdasarkan ID Kemenkes.
- **Memperbarui Data (`PUT /encounter/:id`)**: Mengirimkan kembali *payload* lengkap seperti `POST`.
- **Memperbarui Sebagian (`PATCH /encounter/:id`)**: Menggunakan format JSON Patch.

---

## 🤒 Condition (Diagnosis)
Mencatat diagnosis penyakit pasien dalam sebuah kunjungan.

#### Membuat Data (`POST /condition`)
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

#### Mencari Data (`GET /condition`)
Parameter pencarian yang didukung:
- `patient`: Berdasarkan ID Pasien.
- `encounter`: Berdasarkan ID Kunjungan.
- `code`: Berdasarkan kode diagnosis (ICD-10).

#### Operasi Lainnya
- **Mengambil Data (`GET /condition/:id`)**: Mengambil detail diagnosis berdasarkan ID Kemenkes.
- **Memperbarui Data (`PUT /condition/:id`)**: Mengirimkan kembali *payload* lengkap seperti `POST`.
- **Memperbarui Sebagian (`PATCH /condition/:id`)**: Menggunakan format JSON Patch.

---

## 🩺 Observation (Tanda Vital)
Mencatat hasil pemeriksaan fisik atau lab.

#### Membuat Data (`POST /observation`)
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

#### Mencari Data (`GET /observation`)
Parameter pencarian yang didukung:
- `patient`: Berdasarkan ID Pasien.
- `encounter`: Berdasarkan ID Kunjungan.
- `code`: Berdasarkan kode observasi (LOINC).

#### Operasi Lainnya
- **Mengambil Data (`GET /observation/:id`)**: Mengambil detail observasi berdasarkan ID Kemenkes.
- **Memperbarui Data (`PUT /observation/:id`)**: Mengirimkan kembali *payload* lengkap seperti `POST`.
- **Memperbarui Sebagian (`PATCH /observation/:id`)**: Menggunakan format JSON Patch.

---

## 💊 Medication (Obat - Master KFA)
Membuat *resource* referensi obat berdasarkan Kamus Farmasi dan Alat Kesehatan (KFA).

#### Membuat Data (`POST /medication`)
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

#### Mencari Data (`GET /medication`)
Parameter pencarian yang didukung:
- `code`: Berdasarkan kode KFA.
- `manufacturer`: Berdasarkan ID Organisasi manufaktur.

#### Operasi Lainnya
- **Mengambil Data (`GET /medication/:id`)**: Mengambil detail obat berdasarkan ID Kemenkes.
- **Memperbarui Data (`PUT /medication/:id`)**: Mengirimkan kembali *payload* lengkap seperti `POST`.
- **Memperbarui Sebagian (`PATCH /medication/:id`)**: Menggunakan format JSON Patch.

---

## 📅 EpisodeOfCare (Episode Perawatan)
Mengaitkan serangkaian *Encounter* yang berkaitan dengan masalah medis yang sama.

#### Membuat Data (`POST /episodeofcare`)
```json
{
  "patient_id": "P0123456789",
  "patient_name": "Budi Santoso",
  "organization_id": "10000004",
  "status": "active",
  "period_start": "2023-10-12T08:00:00+07:00"
}
```

#### Mencari Data (`GET /episodeofcare`)
Parameter pencarian yang didukung:
- `patient`: Berdasarkan ID Pasien.
- `status`: Berdasarkan status episode (`planned`, `waitlist`, `active`, `onhold`, `finished`, `cancelled`).

#### Operasi Lainnya
- **Mengambil Data (`GET /episodeofcare/:id`)**: Mengambil detail episode berdasarkan ID Kemenkes.
- **Memperbarui Data (`PUT /episodeofcare/:id`)**: Mengirimkan kembali *payload* lengkap seperti `POST`.
- **Memperbarui Sebagian (`PATCH /episodeofcare/:id`)**: Menggunakan format JSON Patch.

---

## 💡 Panduan Menggunakan Metode `PATCH`

API Kemenkes (Satu Sehat) mengharuskan update parsial menggunakan standar spesifikasi *JSON Patch (RFC 6902)*.
Format *request body* untuk `PATCH` adalah sebuah *array of objects* yang berisi instruksi perubahan.

Contoh: Mengubah status sebuah **Encounter** dari `arrived` menjadi `finished`.

**PATCH** `/encounter/5fa23d13-0943-4318-b295-eb1ecfa7384a`
```json
[
  {
    "op": "replace",
    "path": "/status",
    "value": "finished"
  }
]
```
*(Ini akan mengirim instruksi ke Kemenkes untuk hanya mengganti parameter status menjadi "finished" tanpa menyentuh data lainnya).*