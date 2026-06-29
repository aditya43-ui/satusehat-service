# Development Plan — `service-satusehat`

| Field | Value |
|---|---|
| Document Version | 1.0 |
| Status | Draft for review |
| Owner | Backend Engineering — Meninjar |
| Last Updated | 2026-05-13 |
| Horizon | ~2 quarters (M1 → M5) |

> Read this together with `PRD.md` (the *what / why*), `analysis.qmd` (the
> *as-is*), and `DEVLOG.md` (the *what happened*).

---

## 1. Guiding Principles

1. **Stabilise before extending.** The 19 FHIR resources work end-to-end but
   rest on near-zero tests and an empty gRPC surface. Hardening comes before
   new features.
2. **Pure mappers.** No I/O, no env reads inside `Map*ToFHIR`. Inject from
   the service layer. This is the `EpisodeOfCare` pattern applied broadly.
3. **One change, one PR, one CI run.** No more multi-resource omnibus
   commits like `af32d9c`.
4. **Audit everything that talks to SATUSEHAT.** Outbound calls are
   regulator-visible artefacts.
5. **Production readiness is a checklist, not a feeling.** See §6.

---

## 2. Milestones

### M1 — Hardening (Weeks 1–3)

> Goal: make the codebase *safe to refactor*. After M1, every PR runs
> through CI with linting + tests, the obvious bugs are fixed, and the
> debug-log noise is gone.

| ID    | Task                                                                     | Effort | Owner    | Acceptance |
|-------|--------------------------------------------------------------------------|--------|----------|------------|
| T-01  | Set up CI (GitHub Actions): `go vet`, `golangci-lint`, `go test`, `make audit`, `make security-check`, Docker build | 1 d    | Backend  | PR triggers green pipeline; status checks required for `main` |
| T-02  | Add `golangci-lint.yml` config (errcheck, gocritic, gosec, revive, govet) | 0.5 d  | Backend  | `make lint` returns clean baseline (existing offenders allowlisted, not silenced) |
| T-03  | Unit tests for every `Map*ToFHIR` (19 resources)                         | 4 d    | Backend  | Golden-file tests; `make test` ≥ 60 % coverage on `internal/satusehat/usecase/**` |
| T-04  | Apply the `EpisodeOfCare` pattern to the remaining 18 usecases (drop env reads, accept `orgID` via constructor) | 3 d    | Backend  | No usecase mapper imports `os`; main.go wires `cfg.SatuSehat.OrgID` for all |
| T-05  | Replace `fmt.Printf` debug in `pkg/utils/query/**` with `logger.Debug()` | 0.5 d  | Backend  | `grep -R "fmt.Printf" pkg/utils/query` returns 0 |
| T-06  | Make `NoOp` cache goroutine-safe with `sync.RWMutex`                     | 0.5 d  | Backend  | Race-detector tests pass under load |
| T-07  | Delete commented-out Kafka producer + ImagingStudy worker from `cmd/api/main.go`, or move behind a `KAFKA_ENABLED` feature flag | 0.5 d  | Backend  | `main.go` has no commented blocks > 3 lines |
| T-08  | Rename `internal/master/role/accses` → `access`                          | 0.5 d  | Backend  | One renaming commit; all imports updated |

**Exit criteria:** CI required on `main`, lint clean baseline, FHIR mapper
unit-test coverage ≥ 60 %, no debug `Printf` in `pkg/`.

---

### M2 — Auditability (Weeks 4–5)

> Goal: every outbound SATUSEHAT call leaves a row we can defend in a
> Kemenkes audit.

| ID    | Task                                                                     | Effort | Owner    | Acceptance |
|-------|--------------------------------------------------------------------------|--------|----------|------------|
| T-10  | Design table `satusehat_submission`                                      | 0.5 d  | Backend  | Columns: `id, resource, method, endpoint, request_id, request_body_hash, response_status, response_body_hash, satusehat_resource_id, error_code, created_at, latency_ms, org_id, user_id` |
| T-11  | Migration + GORM model + CommandRepository / QueryRepository             | 1 d    | Backend  | Goose migration runs; CQRS pair compiles |
| T-12  | Wrap `internal/interfaces/satusehat.Client.DoRequest` with an audit interceptor (decorator) | 1 d    | Backend  | Every outbound call writes a row; failures still write |
| T-13  | Add request-ID middleware (X-Request-ID propagation, ULID generation if absent) | 0.5 d  | Backend  | All logs + audit rows carry the same `request_id` |
| T-14  | Retention policy doc — keep submissions for ≥ 25 years per Permenkes 24/2022 | 0.5 d  | Backend / Compliance | `docs/RETENTION.md` exists and references the regulation |
| T-15  | Health probe `/health/satusehat` — last successful call within 5 min     | 0.5 d  | Backend  | Returns `200` when at least one submission within the window |

**Exit criteria:** Audit rows for every outbound call across all 19
resources, joinable on `request_id`, retained per policy.

---

### M3 — gRPC Catch-up (Weeks 6–8)

> Goal: bring the gRPC transport from *empty* to *parity with REST for the
> four hottest resources*.

| ID    | Task                                                                     | Effort | Owner    | Acceptance |
|-------|--------------------------------------------------------------------------|--------|----------|------------|
| T-20  | Author `proto/v1/satusehat/{encounter,episodeofcare,condition,observation}.proto` | 2 d   | Backend  | `make proto` generates clean Go code |
| T-21  | Implement gRPC handlers for the four resources reusing the existing services | 2 d   | Backend  | gRPC + REST share the **same** service layer |
| T-22  | Register handlers in `cmd/api/main.go` gRPC server                       | 0.5 d  | Backend  | `grpcurl … list` shows the four services |
| T-23  | Add buf lint + breaking-change check to CI                               | 0.5 d  | Backend  | PRs that break protobuf surface fail CI |
| T-24  | Postman → grpcurl examples in `docs/api/`                                | 0.5 d  | Backend  | One example call per service in markdown |

**Exit criteria:** Four resources callable via gRPC, behind the same
auth, with parity behaviour to REST.

---

### M4 — Resilience (Weeks 9–10)

> Goal: survive a flaky SATUSEHAT without dropping submissions.

| ID    | Task                                                                     | Effort | Owner    | Acceptance |
|-------|--------------------------------------------------------------------------|--------|----------|------------|
| T-30  | Retry middleware on `Client.DoRequest`: exponential backoff (250 ms · 2^n, cap 8 s), max 5 attempts, retry on 5xx + network errors only | 1 d    | Backend  | Unit tests with fake transport |
| T-31  | Circuit breaker (`sony/gobreaker`) per upstream (FHIR, KFA, KYC, Consent, DICOM) | 1 d    | Backend  | Opens after 5 consecutive failures, half-opens after 30 s |
| T-32  | Idempotency key support — caller supplies `Idempotency-Key`, service dedupes within 24 h via Redis | 1 d    | Backend  | Replays return the original response, not a duplicate submission |
| T-33  | Rate-limit middleware on inbound (Redis token bucket, key by `user_id` then `ip`) | 1 d    | Backend  | `cfg.Security.RateLimit.RequestsPerMinute` enforced |
| T-34  | Refresh-token revocation on logout — persist to `revoked_tokens` table; check on every refresh (closes TODO at `internal/auth/service.go:233`) | 1 d    | Backend  | Logout test passes; refresh returns 401 after logout |
| T-35  | Persist KYC verification result locally (closes TODO in `kyc/service.go`) | 0.5 d  | Backend  | New table `kyc_verification`; result row on every success |
| T-36  | Standardise cache invalidation in `role/master`, `role/pages`, `role/permission`, `role/access` — invalidate on every command op | 1 d    | Backend  | TODOs removed; integration test covers stale-cache scenario |

**Exit criteria:** Service tolerates 30 s of SATUSEHAT 5xx without losing a
submission; logout truly invalidates; cache is never stale after a write.

---

### M5 — Production Rollout (Weeks 11–12)

> Goal: from "running on a dev box" to "running in the hospital DC, one
> instance, watched."

| ID    | Task                                                                     | Effort | Owner    | Acceptance |
|-------|--------------------------------------------------------------------------|--------|----------|------------|
| T-40  | Production `docker-compose.prod.yml` review with ops; document the external network requirement | 0.5 d  | Backend / Ops | `docs/deployment.md` updated; one-shot setup script |
| T-41  | Grafana dashboard JSON (request volume, p50/p95/p99 latency, error rate by code, DB pool, cache hit ratio) | 1 d    | Backend  | Dashboard imports cleanly; screenshot in `docs/` |
| T-42  | Prometheus alert rules (SATUSEHAT 5xx > 5/min, p95 > 5 s for 5 min, DB pool exhaustion) | 0.5 d  | Backend / Ops | Rules under `deploy/prom/alerts.yaml` |
| T-43  | Backup strategy for `satusehat_submission` and master DB (daily pg_dump, 30-day retention onsite, 1-year offsite encrypted) | 0.5 d  | Ops      | `docs/BACKUP.md` exists |
| T-44  | DR runbook — what to do when SATUSEHAT is down, when Redis is down, when the DB primary fails | 1 d    | Backend / Ops | `docs/RUNBOOK.md` exists |
| T-45  | Soft launch — 1 hospital, 1 day of shadow traffic                        | 1 d    | All      | No P0 incidents; audit table populated |
| T-46  | GA — promote to all sites that depend on SATUSEHAT submission            | 0.5 d  | All      | Stakeholder sign-off |

**Exit criteria:** Service running in production with full observability,
alerts, backups, and a runbook.

---

## 3. Backlog (post-M5, not yet scheduled)

- **B-01** Multi-tenant support: one binary, multiple `org_id`s per
  request based on JWT claim (depends on T-04 pattern across all 19).
- **B-02** Event-driven mode: re-enable the commented-out Kafka producer
  (`cmd/api/main.go:137–141`) so that submissions are emitted as events
  for downstream consumers (data warehouse, BI).
- **B-03** Background worker for ImagingStudy DICOM uploads (re-enable
  `cmd/api/main.go:286–300` behind a worker process flag).
- **B-04** OpenTelemetry tracing (replace ad-hoc request-id with
  `traceparent`).
- **B-05** Bulk submission endpoints (`POST /v1/Encounter/bundle`) for
  back-fill scenarios.
- **B-06** Mobile-friendly DTOs (slimmer JSON than the full FHIR shape).
- **B-07** Finish or remove the `tools/generate.go` code generator.
- **B-08** Postman → automated contract tests in CI.

---

## 4. Risk Register

| ID   | Risk                                                              | Likelihood | Impact   | Mitigation                                                   |
|------|-------------------------------------------------------------------|------------|----------|---------------------------------------------------------------|
| RK-1 | Refactor breaks an in-production FHIR resource                    | Med        | High     | T-03 unit tests must land before T-04 refactor cascade        |
| RK-2 | SATUSEHAT contract changes mid-quarter                            | Low        | High     | Contract tests behind a build tag, run nightly against staging |
| RK-3 | Hospital DC ops cannot operate Prometheus stack                   | Med        | Med      | Provide a one-shot Compose + Grafana with default dashboards |
| RK-4 | Single instance is a SPOF                                         | High       | Med      | M5+ plan to run two instances behind nginx; cache is shared so it's safe |
| RK-5 | Audit table grows unbounded                                       | High       | Low      | Partition by month + offsite archive (T-14)                  |
| RK-6 | Refresh-token revocation table grows unbounded                    | Med        | Low      | TTL cleanup job; keep only until exp + 7 d                    |
| RK-7 | gRPC contract evolves and breaks mobile clients                   | Med        | Med      | T-23 buf breaking-change check                                |

---

## 5. Dependencies & Coordination

| Dependency                          | Needed by | Status |
|-------------------------------------|-----------|--------|
| SATUSEHAT staging credentials       | M1+       | ✓ in `.env` |
| Keycloak realm config               | M1+       | depends on hospital IAM team |
| External Docker network `service-general_default` | M5 | ✓ created by `service-general` compose |
| Prometheus + Grafana in hospital DC | M5        | ⚠ to confirm with ops |
| Offsite backup target                | M5        | ⚠ to confirm with ops |

---

## 6. Definition of Done (per task)

A task is **Done** only when:

1. Code merged to `main` via PR.
2. CI green (lint + test + audit + security-check + Docker build).
3. Tests cover the happy path and at least one error path.
4. `docs/DEVLOG.md` has a new entry referencing the task ID.
5. If user-visible: `docs/PRD.md` updated; if architectural: `docs/analysis.qmd` updated.
6. If touching ops: `docs/deployment.md` / `RUNBOOK.md` updated.
7. No new TODO without a backlog ID.

---

## 7. Cadence

- **Daily**: 15-min stand-up over chat; update task status in the issue tracker.
- **Weekly**: 30-min review — what shipped, what's blocked, what's next.
- **Per milestone**: written retrospective appended to `DEVLOG.md` with
  scope `decision`.
