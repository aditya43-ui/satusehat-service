# 🏥 Satu Sehat Interface Library

Package ini menyediakan abstraksi HTTP Client dan Payload Builder yang sangat fleksibel untuk mempermudah integrasi dengan API Satu Sehat (Kemenkes) berbasis standar HL7 FHIR.

## 🌟 Fitur Utama

1. **Auto Auth Management**: Klien ini otomatis me-*request* dan men-*cache* Access Token (OAuth2) di memori. Anda tidak perlu lagi memikirkan urusan *refresh token* yang kedaluwarsa.
2. **Dynamic Endpoint Detection**: Mendukung *routing* dinamis ke `BaseURL` (FHIR), `AuthURL`, `ConsentURL`, maupun `KFAURL`.
3. **FHIRPayload Builder**: Pembuat JSON (*Fluent API / Chaining*) dinamis untuk menyusun *request* spesifikasi FHIR tanpa perlu membuat *struct* raksasa yang kaku.

---

## 📚 Panduan Penggunaan Klien API

Klien API dapat di-*inject* melalui *dependency injection* di `main.go`. Klien ini otomatis menempelkan `Authorization: Bearer <token>` pada setiap *request*.

```go
// Contoh pemanggilan POST ke endpoint /Patient
func (r *patientRepository) CreatePatient(ctx context.Context, payload interface{}) ([]byte, error) {
	// Gunakan DoRequest untuk Base FHIR (BaseURL)
	response, err := r.satuSehatClient.DoRequest(ctx, "POST", "/Patient", payload)
	if err != nil {
		return nil, err
	}
	return response, nil
}
```

---

## 🛠️ Panduan Penggunaan `FHIRPayload` (JSON Builder)

Standar data FHIR banyak menggunakan *array of objects* (`identifier`, `name`, `telecom`, `address`, dll). Menggunakan `struct` Go konvensional akan membuat *codebase* sangat kotor dan tidak fleksibel terhadap perubahan versi FHIR.

Gunakan `FHIRPayload` untuk merakit JSON *on-the-fly*!

### 1. Inisialisasi Resource
Gunakan `NewFHIRPayload` untuk memulai pembuatan resource. Fungsi ini mewajibkan parameter `resourceType`.

```go
payload := satusehat.NewFHIRPayload("Patient")
```

### 2. Menggunakan `.Set()` (Menyimpan Nilai Tunggal)
`.Set()` digunakan untuk atribut *single value* seperti `active`, `gender`, `birthDate`, dll.

```go
payload.
    Set("active", true).
    Set("gender", "male").
    Set("birthDate", "1990-01-01")
```

### 3. Menggunakan `.Append()` (Menyimpan Array Objek)
`.Append()` adalah "senjata utama" untuk FHIR. Atribut seperti NIK, Paspor, No HP, Email, biasanya bertipe *array* di FHIR. Jika *key* (misal: `identifier`) belum ada, `.Append` otomatis membuatnya. Jika sudah ada, ia akan menambahkannya ke urutan berikutnya.

```go
payload.
    // Menambahkan NIK ke dalam array identifier
    Append("identifier", map[string]interface{}{
        "use":    "official",
        "system": "https://fhir.kemkes.go.id/id/nik",
        "value":  "3573012345678901",
    }).
    // Menambahkan NRM (Nomor Rekam Medis) ke array identifier yang sama
    Append("identifier", map[string]interface{}{
        "use":    "usual",
        "system": "https://fhir.kemkes.go.id/id/ihs-number",
        "value":  "P0123456789",
    })
```

---

## 💻 Realisasi Kode (Full Example)

Berikut adalah contoh nyata bagaimana Anda menyusun Payload Pendaftaran Pasien dan langsung mengirimkannya via API Client:

```go
package main

import (
	"context"
	"log"
	"service/internal/interfaces/satusehat"
)

func RegisterPatientExample(ctx context.Context, client satusehat.SatuSehatClient) {
	// 1. Merakit JSON FHIR Payload dengan gaya Chaining (Builder)
	payload := satusehat.NewFHIRPayload("Patient").
		Set("active", true).
		Append("identifier", map[string]interface{}{
			"use":    "official",
			"system": "https://fhir.kemkes.go.id/id/nik",
			"value":  "3573012345678901",
		}).
		Append("name", map[string]interface{}{
			"use":  "official",
			"text": "BUDI SANTOSO",
		}).
		Append("telecom", map[string]interface{}{
			"system": "phone",
			"value":  "08123456789",
			"use":    "mobile",
		}).
		Append("telecom", map[string]interface{}{
			"system": "email",
			"value":  "budi.santoso@example.com",
			"use":    "home",
		}).
		Set("gender", "male").
		Set("birthDate", "1990-12-31")

	// 2. Kirim ke Satu Sehat Kemenkes (Secara otomatis token Bearer disisipkan)
	responseBody, err := client.DoRequest(ctx, "POST", "/Patient", payload)
	if err != nil {
		log.Printf("Gagal mendaftarkan pasien: %v\n", err)
		return
	}

	log.Printf("Sukses mendaftarkan pasien! Response Kemenkes: %s\n", string(responseBody))
}
```

Dengan fitur *chaining* ini, *service* Anda akan bebas dari deklarasi `struct` yang panjangnya bisa ratusan baris, dan penulisan kode akan jauh lebih bersih dan mudah dibaca!