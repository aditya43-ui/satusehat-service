# Product Requirements Document (PRD)
## Service: `service-satusehat`

| Field | Value |
|---|---|
| Document Version | 1.0 |
| Status | Draft |
| Owner | Backend Engineering — Meninjar |
| Last Updated | 2026-05-13 |
| Repository | `goprint/service-satusehat` |
| Primary Language | Go 1.25 |
| Target Audience | Backend engineers, integration partners, ops, hospital IT (RS) |

---

## 1. Overview

### 1.1 Product Summary

`service-satusehat` is a backend microservice that acts as the **integration gateway** between an Indonesian hospital information system (SIMRS / HIS) and Indonesia's national health platforms — primarily **SATUSEHAT** (HL7 FHIR R4) operated by the Ministry of Health, with secondary integration paths for **BPJS Kesehatan** (VClaim, Antrol, Apotek, Aplicare, IHS) and supporting infrastructure (Keycloak SSO, MinIO object storage, Redis cache, Prometheus observability).

The service exposes a **dual transport surface** (REST via Gin and gRPC) and is built on **Clean Architecture + Domain-Driven Design + CQRS**. It centralizes all outbound calls to SATUSEHAT, normalizes FHIR resource construction, manages OAuth2 tokens, and protects upstream callers from FHIR/JSON-Patch complexity.

### 1.2 Problem Statement

Hospitals in Indonesia are mandated by the Ministry of Health to send clinical encounter data to SATUSEHAT (HL7 FHIR). Direct integration is hard because:

1. **Heterogeneous internal data sources** — most hospitals have data spread across multiple databases (PostgreSQL / MySQL / SQL Server) and legacy systems.
2. **Strict FHIR R4 contract** — resource structure, references, codings, and identifier systems must follow SATUSEHAT IG.
3. **OAuth2 token lifecycle** — SATUSEHAT requires short-lived tokens with automatic refresh and concurrency-safe caching.
4. **Auxiliary services** — KFA (drug master), KYC, Consent, DICOM, BPJS all have different auth, signing, and payload conventions.
5. **Observability & auditability** — every regulated submission must be logged, traceable, and retryable.

A single integration service that solves the above once, and is reusable across hospital products, is significantly cheaper than re-implementing the contract in every product team.

### 1.3 Goals

| # | Goal | Success indicator |
|---|---|---|
| G1 | Provide one canonical Go service that submits every required FHIR resource to SATUSEHAT | All 19 FHIR resources implemented with Create/Update/Patch/Get/Search |
| G2 | Abstract OAuth2 + auxiliary auth (KFA, KYC, Consent, DICOM) from callers | Single client handles token caching and header injection |
| G3 | Be deployable in a hospital data center with minimal ops effort | Single Docker image, distroless, Compose stack with dev/prod profiles |
| G4 | Stay observable in production | Prometheus metrics, structured logs, health endpoints, request tracing fields |
| G5 | Stay vendor-agnostic at the storage layer | Multi-DB support (Postgres, MySQL, SQL Server, MongoDB, SQLite) with read replicas |
| G6 | Be safe to refactor and extend | Clean Architecture boundaries, generated code, CQRS separation |

### 1.4 Non-Goals

- ❌ **Front-end UI** — this is a backend service only.
- ❌ **Long-term clinical data storage** — the hospital's primary database remains the source of truth.
- ❌ **Generic FHIR server** — only SATUSEHAT-required resources and profiles are supported, not the full HL7 FHIR spec.
- ❌ **BPJS claim submission UI/UX** — only the integration layer is in scope.
- ❌ **HL7 v2 messaging** — out of scope.

---


### 1.1 Panduan Eksekusi AI (Agentic Workflow)
Bagi Agen AI yang bertugas mengimplementasikan spesifikasi dalam dokumen ini, **WAJIB** menerapkan siklus berikut (berdasarkan 24 *Agent Skills*):
1. **`spec-driven-development`**: Jangan mulai menulis kode sebelum membuat/memperbarui file spesifikasi teknis (terutama arsitektur dan API contract) di folder `docs/specs/`. Spesifikasi harus eksplisit terkait skema DB, struktur *request/response* (DTO), CQRS, dan edge-cases.
2. **`planning-and-task-breakdown`**: Pecah implementasi fitur menjadi tugas-tugas kecil yang terukur di dalam file `task.md`. Implementasi WAJIB mengikuti urutan Clean Architecture: `Entity/DTO -> Mapper -> Repository -> Service -> Transport (Handler)`.
3. **`doubt-driven-development`**: Jika spesifikasi terasa ambigu atau kontrak payload (terutama integrasi BPJS/SATUSEHAT) tidak jelas, jangan berasumsi. Tanyakan kepada user untuk menghindari kerugian operasional.
4. **`security-and-hardening` & `observability-and-instrumentation`**: Pastikan seluruh input tervalidasi dengan aman, dan setiap proses mutasi data mencatat `request_id` melalui `logger.WithContext(ctx)` guna keperluan pelacakan audit.

## 2. Target Users & Personas

### 2.1 Primary Users (callers)

| Persona | Description | What they need |
|---|---|---|
| **SIMRS Backend** | Internal hospital ERP/HIS server that emits clinical events | A stable REST/gRPC contract that hides SATUSEHAT details |
| **Mobile / Web App teams** | Patient portal, clinician app | A simple "submit encounter" API; no FHIR construction |
| **Data engineering** | Builds dashboards / reports on submitted resources | Read-back, search, and audit logs |

### 2.2 Secondary Users

| Persona | Description |
|---|---|
| **Ops / SRE** | Runs the container in the hospital data center. Cares about logs, metrics, health, restart safety |
| **Integration developers** | Add new SATUSEHAT resources, debug failures |
| **Compliance / regulators** | Need audit trail of submissions for inspection |

---

## 3. Functional Requirements

### 3.1 SATUSEHAT FHIR Submission (P0 — core value)

The service MUST expose endpoints for the following 19 FHIR resources, each with **Create / Update (full) / Patch (JSON Patch RFC 6902) / Get-by-ID / Search**:

| # | Resource | Use case |
|---|---|---|
| 1 | AllergyIntolerance | Patient allergy registry |
| 2 | CarePlan | Treatment plan |
| 3 | ClinicalImpression | Clinician assessment |
| 4 | Composition | Document grouping |
| 5 | Condition | Diagnosis (ICD-10) |
| 6 | DiagnosticReport | Lab / radiology report |
| 7 | Encounter | Patient visit (rawat jalan/inap/IGD) |
| 8 | EpisodeOfCare | Continuous treatment episode |
| 9 | ImagingStudy | Radiology study link |
| 10 | Immunization | Vaccination record |
| 11 | Medication | Drug definition |
| 12 | MedicationDispense | Pharmacy dispense event |
| 13 | MedicationRequest | Prescription |
| 14 | MedicationStatement | Patient-reported medication |
| 15 | Observation | Vital signs, lab results |
| 16 | Procedure | Performed procedure (ICD-9-CM) |
| 17 | QuestionnaireResponse | Form / triage answers |
| 18 | ServiceRequest | Lab / radiology order |
| 19 | Specimen | Lab specimen |

Plus minimal **Studies (DICOM)** for imaging linkage.

### 3.2 SATUSEHAT Reference Lookups (P0)

The service MUST proxy reads to the following SATUSEHAT registries:

- **Patient** — search by NIK, name, DOB
- **Practitioner** — by NIK or NPP
- **Organization** — facility registry
- **Location** — room/bed
- **KFA** — Katalog Farmasi Alkes (drug & device master)
- **KYC** — verifikasi identitas
- **Auth** — token refresh / validate

### 3.3 Authentication & Authorization (P0)

- **Inbound**: JWT-based auth (with Keycloak, static-token, and hybrid modes). RBAC via Role / Page / Permission / Access master tables.
- **Outbound**: OAuth2 client-credentials flow to SATUSEHAT with concurrency-safe in-memory token caching (≥1 minute buffer before expiry).
- **BPJS**: HMAC-SHA256 signature header per BPJS spec (cons-id + secret).

### 3.4 Storage & Master Data (P1)

- Multi-database support (Postgres preferred) for:
  - Auth (users, sessions, tokens)
  - Role / Page / Permission / Access (RBAC master)
  - Audit log of SATUSEHAT submissions (TBD — see §7 Gaps)
- Read replicas with round-robin load balancing.
- GORM as primary ORM; raw SQL via sqlx / squirrel where ORM is too slow.

### 3.5 Cross-Cutting (P1)

- Structured logging (logrus) to console + daily file `logs/YYYY/MM/YYYY-MM-DD.log`.
- Prometheus metrics: HTTP request counts/latency, DB pool stats, cache hit ratio, error counters by code.
- Health endpoints: overall + per-database + external (SATUSEHAT, KFA, KYC).
- i18n-ready error system (English + Indonesian).
- Soft-delete + audit trail at the master-data layer.

### 3.6 Object Storage (P2)

- MinIO integration for storing supporting artifacts (e.g., DICOM tarballs, KYC scans, consent PDFs). Buckets configurable per environment.

---

## 4. Non-Functional Requirements

| Category | Requirement |
|---|---|
| **Performance** | P95 latency for outbound SATUSEHAT create ≤ 2 s when SATUSEHAT responds within SLA; service overhead ≤ 50 ms |
| **Throughput** | ≥ 100 sustained writes/sec per service instance (single-node, 4 vCPU) |
| **Availability** | 99.5 % (hospital data-center, single instance acceptable; horizontal scale-out supported) |
| **Recovery** | Graceful shutdown ≤ 5 s; in-flight requests drained on SIGTERM |
| **Security** | No plaintext secrets in logs; parameterised queries; CORS configurable; rate-limit configurable |
| **Compliance** | Comply with Permenkes 24/2022 (rekam medis elektronik) and SATUSEHAT IG R4 |
| **Observability** | Every outbound SATUSEHAT call MUST be traceable by `request_id` (TBD) |
| **Portability** | Single static binary, distroless image, runs on amd64 Linux |
| **Time zone** | All timestamps in `Asia/Jakarta` (WIB); FHIR payloads use ISO-8601 with offset |

---

## 5. Out-of-Scope Items (deliberately excluded)

- Web frontend / admin UI.
- Long-term clinical archival.
- HL7 v2 messaging.
- Direct BPJS claim file generation (RBAC adjustments only).
- Patient-facing mobile API.

---

## 6. Dependencies

### 6.1 External (third-party)

| Dependency | Purpose | Notes |
|---|---|---|
| SATUSEHAT FHIR R4 | Primary integration target | Staging + production base URLs configured |
| SATUSEHAT KFA | Drug & device master | Separate endpoint and headers |
| SATUSEHAT KYC | Identity verification | Public/private key pair (config) |
| SATUSEHAT Consent | Patient consent | Webhook secret configured |
| SATUSEHAT DICOM | Imaging study upload | Separate endpoint |
| BPJS VClaim / Antrol / Apotek / Aplicare / IHS | Insurance integration | HMAC-SHA256 auth |
| Keycloak | SSO / OAuth2 issuer (optional) | JWKS validation |

### 6.2 Internal

- PostgreSQL 12+ (recommended)
- Redis 6+ (cache, rate limit) — optional but recommended
- MinIO (S3-compatible) — optional
- Prometheus (scrape `/metrics`)

---

## 7. Known Gaps & Risks (carried into DEVPLAN)

| # | Gap | Severity | Owner |
|---|---|---|---|
| R1 | gRPC defined in config but **no `.proto` files committed** and registry empty | High | Backend |
| R2 | Test coverage ≈ **2 files total** — query builder only | High | Backend |
| R3 | Rate-limit config exists, no middleware implementation | Medium | Backend |
| R4 | Auth logout does **not** revoke refresh tokens in DB (TODO at `internal/auth/service.go:233`) | Medium | Backend |
| R5 | KYC service does **not** persist verification to local DB | Medium | Backend |
| R6 | Role / Page / Permission services have TODO cache invalidation | Medium | Backend |
| R7 | 60+ `fmt.Printf` debug calls in `pkg/utils/query` — leaks to stdout in production | Low | Backend |
| R8 | NoOp cache fallback is not thread-safe (map without mutex) | Medium | Backend |
| R9 | Folder name typo `internal/master/role/accses/` (should be `access`) | Low | Backend |
| R10 | Audit-log table for SATUSEHAT submissions not yet defined | High | Backend |
| R11 | No CI pipeline committed | High | Ops |

---

## 8. Success Metrics

| Metric | Target |
|---|---|
| SATUSEHAT submission success rate | ≥ 99 % (per resource, per hour) |
| Coverage on `internal/satusehat/usecase/**` | ≥ 60 % unit tests, ≥ 1 contract test per resource |
| Mean time to add a new FHIR resource | ≤ 1 day (via code-gen template) |
| Production crash-loop incidents | 0 per month |
| Time to detect a SATUSEHAT outage | ≤ 1 minute via health probe |

---

## 9. Release Plan (high-level — detail in DEVPLAN)

| Milestone | Description |
|---|---|
| **M1 — Hardening** | Fix R1–R3, R7, R8; add CI; unit-test FHIR mappers |
| **M2 — Auditability** | Implement audit-log table + middleware emitting one row per outbound SATUSEHAT call (R10) |
| **M3 — gRPC** | Author proto files; generate; register services for at least Encounter, EpisodeOfCare, Condition, Observation |
| **M4 — Resilience** | Retry with exponential backoff for SATUSEHAT 5xx + circuit breaker |
| **M5 — Production rollout** | Single-instance deployment in hospital DC, then dual-instance behind nginx |

---

## 10. Glossary

| Term | Meaning |
|---|---|
| **SATUSEHAT** | Indonesia national health-data platform (Kemenkes) |
| **FHIR** | Fast Healthcare Interoperability Resources (HL7 R4) |
| **BPJS** | Badan Penyelenggara Jaminan Sosial — Indonesian social insurance |
| **VClaim** | BPJS claim API |
| **KFA** | Katalog Farmasi Alkes (drug/device catalogue) |
| **KYC** | Know Your Customer — identity verification |
| **SIMRS** | Sistem Informasi Manajemen Rumah Sakit (hospital management system) |
| **IHS** | Indonesia Health Services (BPJS) |
| **CQRS** | Command Query Responsibility Segregation |
| **DDD** | Domain-Driven Design |
