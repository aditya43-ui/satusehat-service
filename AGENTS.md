# AGENTS.md — Universal AI Agent Instructions

> **Standar lintas-tool** untuk agent AI: Google Antigravity, Cursor, Windsurf, GitHub Copilot, Codex, dan agent lain yang membaca `AGENTS.md`.
>
> File ini adalah **anchor ringkas**. Detail lengkap ada di [`docs/PRD.md`](file:///home/meninjar/goprint/medical/medical-service-satusehat/docs/PRD.md) — pedoman untuk service gateway ini, dan [`SKILL_WORKSPACE.md`](file:///home/meninjar/project%20nuxt/medical/SKILL_WORKSPACE.md) untuk orkestrasi lintas-repo.

---

## ⚡ Konteks Project (30-detik)

- **Apa**: `service-satusehat` — gateway integrasi SATUSEHAT Kemenkes RI (HL7 FHIR R4, KFA, KYC, Consent, DICOM) berbasis Go 1.25 + Gin (REST) + gRPC + CQRS.
- **Arsitektur**: Clean Architecture + DDD + CQRS. Layer: `transport (handlers) → service → repository → database/external`.
- **Status (2026-06-26)**: Berfungsi sebagai gateway integrasi nasional. REST `:8871` · gRPC `:8872`.
- **Bahasa**: komunikasi user & dokumentasi → **Bahasa Indonesia**. Code identifier → English.

---


## 🧠 Penggunaan Agent Skills (WAJIB)

Sebagai agen AI, Anda dilengkapi dengan 24 *Agent Skills* (termasuk `securecoder` dan `agent-skills`). Anda **WAJIB** menerapkannya:
1. **Inisiasi**: Mulai sesi dengan skill `context-engineering` dan selalu jadikan `SKILL.md` serta `SKILL_WORKSPACE.md` sebagai sumber kebenaran.
2. **Klarifikasi**: Gunakan `interview-me` jika ada requirement yang ambigu (jangan berasumsi).
3. **Eksekusi**: Terapkan `api-and-interface-design` (untuk kontrak REST/gRPC), `test-driven-development`, `observability-and-instrumentation`, dan `security-and-hardening` saat mengeksekusi kode.
4. **Validasi Ketergantungan**: WAJIB menggunakan `scan_dependencies` jika ingin menambah library baru.
5. **Navigasi Meta**: Gunakan meta-skill `using-agent-skills` sebagai rujukan kapan harus memakai skill lainnya.

## 📚 Bacaan Wajib Sebelum Bekerja (urutan)

1. [`SKILL_WORKSPACE.md`](file:///home/meninjar/project%20nuxt/medical/SKILL_WORKSPACE.md) — **pedoman orkestrasi lintas-repo** (peta port backend/frontend, database bersama, dll). **Sangat penting.**
2. [`docs/PRD.md`](file:///home/meninjar/goprint/medical/medical-service-satusehat/docs/PRD.md) — spesifikasi fitur & integration scope.
3. [`docs/DEVLOG.md`](file:///home/meninjar/goprint/medical/medical-service-satusehat/docs/DEVLOG.md) — decision log (cek 3 entry teratas untuk konteks recent).
4. [`README.md`](file:///home/meninjar/goprint/medical/medical-service-satusehat/README.md) — petunjuk instalasi & setup.

---

## ⚙️ 12 Konvensi WAJIB (ringkas)

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

## 🛑 JANGAN

- ❌ Re-eksplorasi project dari nol — gunakan `docs/*` dan `SKILL_WORKSPACE.md` sebagai konteks otoritatif.
- ❌ Jalankan `git add -A` / `git add .` — selalu add file spesifik.
- ❌ Commit / push tanpa user eksplisit minta.
- ❌ Modifikasi versi Go di `go.mod` (tetap 1.25.0).
- ❌ Jalankan migrasi database (`goose up`) langsung tanpa izin tertulis dari user.

## ✅ HARUS

- ✅ Baca `docs/DEVLOG.md` (3 entry teratas) di awal sesi.
- ✅ Ikuti pattern integration existing (FHIR client helper).
- ✅ Update `docs/DEVLOG.md` setelah task selesai.
- ✅ Komunikasi dengan user dalam Bahasa Indonesia.
- ✅ Kalau ragu → tanya user, jangan asumsi.

---

## 🔑 Operasional Cepat

- REST: `:8871` (default), Swagger `/swagger/index.html`. gRPC: `:8872`.
- Health: `/health`, `/health/database`, `/health/external` (SATUSEHAT).
