# Development Log — `service-satusehat`

> Append-only journal of meaningful changes, decisions, and incidents.
> Newest entries on top. Each entry: **date — author — scope — summary**,
> followed by **what / why / how / impact** as needed.
>
> Conventions:
> - `feat` new feature · `fix` bug fix · `refactor` no behaviour change ·
>   `chore` housekeeping · `infra` ops / build · `docs` documentation ·
>   `incident` production issue · `decision` ADR-lite.

---

## 2026-06-20 — meninjar — `chore` · scope: `config` / `env`

**Summary.** Membersihkan seluruh file env + `config.yaml` agar service hanya
membawa konfigurasi yang relevan untuk SatuSehat. Menghapus env koneksi lain
(BPJS dan database sumber SIMRS/SATUDATA/FARMASI/RIS) yang **terbukti tidak
dipakai kode**.

**What changed:**

- `.env`, `.env.prod`, `.env.xtx` — dihapus blok: semua `BPJS_*`,
  `POSTGRES_SIMRS_*`, `POSTGRES_SATUDATA_*`, `SQLSERVER_FARMASI_*`,
  `MYSQL_RIS_*`, blok DB komentar mati, dan baris non-env invalid
  (`untuk vclaim`, `// ...`). Dipertahankan: server, logger, auth/keycloak,
  `POSTGRES_DEFAULT_*`, Redis, `SATUSEHAT_*` + `KYC_*`, MinIO, Swagger, CORS.
- `.env.example` — dirombak jadi template SatuSehat-only (tanpa BPJS),
  port contoh 8094→8871.
- `config.yaml` — dihapus section `bpjs:` dan database `satudata:` / `simrs:`.

**Why.** Service di-fork dari template `service-general` yang masih membawa
konfigurasi BPJS + multi-DB. Verifikasi kode: `cmd/api/main.go` hanya men-wire
SatuSehat + auth/RBAC dan hanya memakai koneksi DB `"default"`; BPJS tidak
pernah di-instansiasi; SIMRS/SATUDATA/FARMASI/RIS tidak direferensikan.
`config.yaml` tidak dibaca (tidak ada viper) sehingga perubahannya kosmetik
tapi menghindari kebingungan.

**Impact.**

- Tidak ada perubahan perilaku runtime: kode tak pernah memakai koneksi yang
  dihapus. `cfg.Validate()` aman — BPJS kini `Enabled=false`, SatuSehat lengkap,
  DB `default` memenuhi syarat "minimal 1 database".
- ⚠️ Perubahan `.env` baru aktif setelah **container di-restart** (env_file
  di-inject saat start; Air hanya hot-reload kode Go).
- Struct `BpjsConfig` + loader-nya di `internal/infrastructure/config/config.go`
  dan folder `internal/interfaces/bpjs/` masih ada (dead config, harmless).
  Penghapusan kode itu opsional, follow-up terpisah.

---

## 2026-05-13 — meninjar — `refactor` · scope: `episodeofcare`

**Summary.** Refactored `EpisodeOfCare` usecase to remove environment-variable
reads from the mapper and to inject the organisation ID via the service
constructor.

**What changed (uncommitted as of this entry):**

- `internal/satusehat/usecase/episodeofcare/mapper.go`
  - Removed `import "os"` and the `os.Getenv("SATUSEHAT_ORG_ID")` lookup.
  - `MapRequestToFHIR` now takes `OrganizationID` purely from `req`.
- `internal/satusehat/usecase/episodeofcare/repository.go`
  - Removed the `hiddenCtx struct{ context.Context }` wrapper.
  - `executeRequest` now passes `ctx` to `client.DoRequest` directly.
- `internal/satusehat/usecase/episodeofcare/service.go`
  - Constructor `NewService(repo, orgID string)` (was `NewService(repo)`).
  - Both `Create` and `Update` inject `s.orgID` into `req.OrganizationID`
    when the caller leaves it blank.
- `cmd/api/main.go`
  - Updated wiring to pass `cfg.SatuSehat.OrgID` into `episodeofcare.NewService`.

**Why.**

1. Pure mappers (no env reads, no I/O) are trivial to unit-test.
2. Removes hidden global coupling — one binary can now serve multiple
   facility identities by constructing multiple services with different
   `orgID`s.
3. `hiddenCtx` was an undocumented wrapper that obscured the context chain
   and broke `context.Value` propagation.

**Impact.**

- No public API change for HTTP callers.
- Other 18 FHIR usecases still read env in their mappers — they are flagged
  in `DEVPLAN.md` task **T-04** for the same treatment.

**Follow-ups.**

- Add unit tests for `MapRequestToFHIR` (golden-file style).
- Replicate the pattern in remaining usecases.

---

## 2026-05-13 — meninjar — `docs` · scope: project

**Summary.** Authored four governance documents under `docs/`:

- `docs/PRD.md` — Product Requirements Document.
- `docs/analysis.qmd` — Quarto technical analysis (architecture, modules, gaps).
- `docs/DEVLOG.md` — this file.
- `docs/DEVPLAN.md` — forward roadmap with milestones M1–M5.

**Why.** No high-level product or planning docs existed in the repo. The
`README.md` and `DOCUMENT.md` cover how to run the service but not
**what / why / next**.

**Impact.** Onboarding new engineers, communicating with stakeholders, and
prioritising the next two quarters of work all become tractable.

---

## 2026-05-06 — meninjar — `infra` · scope: Docker / Compose

**Summary.** Bumped builder image and reshaped the production Compose file.

**What changed:**

- `Dockerfile`
  - `golang:1.22-alpine` → `golang:1.25-alpine`.
  - Added `WORKDIR /app` in the final stage.
  - Dropped the `HEALTHCHECK` directive (was firing `/app/main -health`
    which is not a real flag — false-failing the container).
  - Exposed port changed `8080 → 8196` to match the production listener.
- `docker-compose.dev.yml`
  - Container name renamed `service-satusehat → service-satusehat-dev`.
  - Removed the orphan `CONFIG_PATH=/app/config.yaml` env (config is loaded
    relative to the binary, not from an absolute path).
- `docker-compose.prod.yml`
  - Replaced the inline `redis:7-alpine` sidecar with an external network
    `service-general_default` that points to the **shared** Redis run by
    `service-general`.
  - Ports remapped to `8196:8196` (REST) and `8197:8197` (gRPC, future).
  - Container renamed `service-general → service-satusehat-prod`.

**Why.**

- Hospital DC standardises on Go 1.25 toolchain.
- Health-check directive was generating false alarms and feeding monitoring
  noise — the in-app `/api/v1/health` endpoint already exists.
- Redis must be shared with `service-general` to avoid duplicate caches and
  inconsistent rate-limit state.

**Impact.**

- Production stack now requires the external network to exist:
  `docker network create -d bridge service-general_default` (or it is
  created by the `service-general` repo's compose).

---

## 2026-05-04 → 2026-05-12 — meninjar — `chore` · scope: logs

**Summary.** Daily log files accumulated under `logs/2026/05/` from dev runs.

These are *not* committed; they exist only on the developer machine. They
should be ignored via `.gitignore` (already excluded by repo policy).

---

## af32d9c — meninjar — `feat` · scope: SATUSEHAT (commit on `main`)

**Title:** *penambahan All case Satu sehat*

**Summary.** Added the remaining FHIR usecases so that the service now
covers all 19 SATUSEHAT resources end-to-end.

**Resources covered in this commit batch:**
AllergyIntolerance, CarePlan, ClinicalImpression, Composition, Condition,
DiagnosticReport, Encounter, EpisodeOfCare, ImagingStudy, Immunization,
Medication, MedicationDispense, MedicationRequest, MedicationStatement,
Observation, Procedure, QuestionnaireResponse, ServiceRequest, Specimen
(plus minimal Studies for DICOM linkage).

**Pattern adopted:** `dto.go` + `mapper.go` + `repository.go` + `service.go`
per resource (see Appendix B of `docs/analysis.qmd`).

---

## 135c631 — meninjar — `fix` · scope: SATUSEHAT (commit on `main`)

**Title:** *Perbaikan Service Satu Sehat*

**Summary.** Bug fixes across the SATUSEHAT client and early usecase
modules. Details inferred from the commit name; per-file detail to be
back-filled when a follow-up retrospective is held.

---

## 0adf9ef — meninjar — `fix` · scope: SATUSEHAT (commit on `main`)

**Title:** *Perbaikan Service Satu sehat*

**Summary.** Further fixes after the initial SATUSEHAT integration
landing. Same back-fill caveat as above.

---

## 4e59b96 — meninjar — `chore` · scope: gRPC (commit on `main`)

**Title:** *perbaikan GRPC Generate*

**Summary.** Adjusted the gRPC code-generation tooling. Net result is that
the generation pipeline *runs*, but no `.proto` files are committed and
no services are registered yet — see DEVPLAN M3.

---

## edfaa88 — meninjar — `feat` · scope: project (initial commit)

**Title:** *first commit*

**Summary.** Initial scaffold of the service: Clean Architecture layout,
multi-DB layer, error system, query builder, auth module, master/role
modules, BPJS client skeleton, SATUSEHAT client skeleton, Dockerfile,
docker-compose files, Makefile, README.md, DOCUMENT.md.

---

# How to add a new entry

Copy this template to the top of the file:

```markdown
## YYYY-MM-DD — author — `feat|fix|refactor|chore|infra|docs|incident|decision` · scope: <module>

**Summary.** One-paragraph headline.

**What changed.** Bullet list of file-level changes (use `file:line` when
useful).

**Why.** Motivation, constraint, or incident that drove the change.

**Impact.** API breakage? Ops action required? New env vars?

**Follow-ups.** Anything intentionally left for later, linked to a
DEVPLAN task ID when applicable.
```

Rules:

- **Append-only.** Never rewrite a historical entry; correct via a new
  entry that references the older one.
- **Decisions** that are architecturally load-bearing get their own entry
  with the `decision` label (think ADR-lite).
- Keep entries terse: enough that a new joiner can reconstruct the
  reasoning without paging in the PR.
