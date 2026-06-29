# GEMINI.md — Project Instructions for Gemini (Service SatuSehat)

> **Untuk Gemini CLI** (atau Gemini Code Assist / Cloud Code AI): file ini auto-loaded saat working directory di `medical-service-satusehat`. **Baca seluruhnya sebelum mengeksekusi task apa pun di repo ini.** Jangan re-eksplorasi yang sudah dijelaskan di sini.
>
> Untuk pedoman umum AI (Claude + Gemini + lainnya), lihat juga [docs/PRD.md](file:///home/meninjar/goprint/medical/medical-service-satusehat/docs/PRD.md) dan [SKILL_WORKSPACE.md](file:///home/meninjar/project%20nuxt/medical/SKILL_WORKSPACE.md).

---

## ⚡ TL;DR — Konteks Project (30-detik)

- **Apa**: `service-satusehat` — gateway integrasi SATUSEHAT Kemenkes RI (HL7 FHIR R4, KFA, KYC, Consent, DICOM) berbasis Go 1.25 + Gin (REST) + gRPC + CQRS.
- **Arsitektur**: Clean Architecture + DDD + CQRS. Layer: `transport (handlers) → service → repository → database/external`.
- **Status (2026-06-26)**: Berfungsi sebagai gateway integrasi nasional. REST `:8871` · gRPC `:8872`.
- **Domain bahasa**: User & dokumentasi pakai **Bahasa Indonesia**. Code identifier Bahasa Inggris. Komunikasi dengan user: Indonesia.

---

## 📚 Dokumen yang HARUS Dibaca Dulu (urutan prioritas)

Saat session baru, **baca dokumen ini SEBELUM tool call lain**:

1. **`SKILL_WORKSPACE.md`** — Pedoman orkestrasi lintas-repo.
2. **`docs/PRD.md`** — Spesifikasi fitur & integration scope.
3. **`docs/DEVLOG.md`** — Riwayat keputusan teknis (reverse-chrono). Cek 3 entry teratas untuk konteks recent.
4. **`README.md`** — Petunjuk instalasi & setup.

---

## ⚙️ Konvensi WAJIB (jangan dilanggar tanpa diskusi)

1. **Update DEVLOG**: Setiap task selesai → **WAJIB update DEVLOG** (`docs/DEVLOG.md`). Task belum dianggap selesai sebelum DEVLOG diperbarui.
2. **Layer Dependency**: `transport → service → repository → database`. Jangan dilanggar.
3. **CQRS Pattern**: Write ops memakai `CommandRepository`, Read ops memakai `QueryRepository`. Jangan dicampur.
4. **Struktur Modul**: Harus memiliki file `entity.go`, `dto.go` (field request pointer), `mapper.go`, `service.go`, dan `repository.go`. Handler di folder terpisah.
5. **FHIR R4 Standards**: Integrasi SATUSEHAT wajib mematuhi pemetaan resource HL7 FHIR R4 (Encounter, Patient, Practitioner, Condition, dll).
6. **Error handling**: Wajib menggunakan fluent `pkg/errors` builder. JANGAN `errors.New` polos.
7. **Response Formatter**: Wajib menggunakan `pkg/response`.
8. **Logger**: Wajib menggunakan `logger.WithContext(ctx)` untuk menyebarkan `request_id`.
9. **Validator**: Input wajib lewat `pkg/utils/validator` + `TranslateError()` (Bahasa Indonesia).
10. **Secret**: JANGAN commit `.env`. Perbarui `.env.example` saat ada environment variable baru.
11. **Larangan Run Backend**: **JANGAN jalankan service secara manual** (`go build`/`go run`/`air`/`docker compose`). Gunakan Docker logs untuk verifikasi compile error.
12. **Git Lintas-Repo**: Commit hanya untuk repositori ini. Jangan `git add .` / `git add -A`. Commit file spesifik.

---

## 🚦 Decision Tree untuk Task Umum

### "Tambah integrasi resource FHIR baru"
1. Pahami format JSON schema resource FHIR Kemenkes RI.
2. Tulis DTO dan Entity model yang sesuai di modul terkait.
3. Buat mapper konversi data lokal ke format FHIR R4.
4. Tulis adapter / service client untuk memanggil API SATUSEHAT.
5. Register handler HTTP/gRPC.
6. Catat log mutating audit trail dengan aman.
7. Update `docs/DEVLOG.md`.

---

## 🛑 JANGAN

- ❌ Re-eksplorasi project dari nol — gunakan `docs/*` dan `SKILL_WORKSPACE.md` sebagai konteks otoritatif.
- ❌ Jalankan `git add -A` / `git add .` — selalu add file spesifik.
- ❌ Commit / push tanpa user eksplisit minta.
- ❌ Modifikasi versi Go di `go.mod` (tetap 1.25.0).
- ❌ Jalankan migrasi database (`goose up`) langsung tanpa izin tertulis dari user.

---

## ✅ HARUS

- ✅ Baca `docs/DEVLOG.md` (3 entry teratas) di awal sesi.
- ✅ Ikuti pattern integration existing (FHIR client helper).
- ✅ Update `docs/DEVLOG.md` setelah task selesai.
- ✅ Komunikasi dengan user dalam Bahasa Indonesia.
- ✅ Kalau ragu → tanya user, jangan asumsi.

---

## 🤖 Gemini-Specific Notes

- **Gemini CLI** auto-discovery: `GEMINI.md` di root + di folder yang sedang aktif.
- **Long context advantage**: Gemini punya context window 1M+ token — bisa load `docs/*.md` dan `SKILL_WORKSPACE.md` semua sekaligus tanpa khawatir. Manfaatkan ini, JANGAN buang token untuk re-grep file yang sudah didokumentasikan.
- **Function calling**: gunakan untuk memverifikasi logs atau error compile dari container, jangan rebuild/run manual.
- **Cross-check**: jika ragu dengan detail di repo ini, periksa `SKILL_WORKSPACE.md` yang merupakan petunjuk orkestrasi lintas-repo.
