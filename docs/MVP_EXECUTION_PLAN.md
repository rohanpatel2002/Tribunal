# TRIBUNAL MVP Execution Plan (2-3 Weeks)

## Objective

Deliver a working MVP that can:
1. Receive GitHub pull request webhook payloads
2. Run deterministic AI-authorship heuristics on changed files
3. Produce a machine-readable risk report
4. Persist analysis metadata in PostgreSQL
5. Return actionable reviewer guidance in < 30 seconds for small PRs

---

## MVP Contract

### Inputs
- GitHub `pull_request` webhook payload (`opened`, `synchronize`, `reopened`)
- File-level change metadata:
  - `path`
  - `patch` (diff text)
  - `status` (`added`, `modified`, `removed`)

### Outputs
- JSON response for each analyzed file:
  - `aiScore` (0.0 to 1.0)
  - `isAIGenerated` (bool)
  - `confidence` (0.0 to 1.0)
  - `signals` (`style`, `pattern`, `risk`)
  - `riskLevel` (`LOW`, `MEDIUM`, `HIGH`, `CRITICAL`)
  - `summary` (short reviewer note)
- PR summary:
  - total files analyzed
  - AI-flagged files count
  - risk distribution
  - recommendation (`APPROVE`, `REVIEW_REQUIRED`, `BLOCK`)

### Success Criteria
- Handles webhook payload validation and malformed input safely
- Deterministic scoring (same input => same output)
- p95 API latency < 2s for 20-file payload without external LLM calls
- Unit tests for scoring and classification logic
- Build/test commands run green in local environment

---

## Scope (In vs Out)

### In Scope (MVP)
- Go webhook service (`/health`, `/webhook/github`, `/analyze`)
- Heuristic scoring engine (3-signal baseline)
- PostgreSQL schema + repository interface stubs
- Local run support and test harness

### Out of Scope (Post-MVP)
- GitHub check-run writebacks
- Full Claude-based semantic analyzer
- Multi-platform support (GitLab/Gitea)
- Production authz, tenancy, and enterprise RBAC

---

## Key Edge Cases

1. Empty PR / empty file patches
2. Very large patch body (truncate + safe scoring)
3. Binary or non-text files
4. Unsupported file extensions
5. Missing required webhook fields
6. Duplicate webhook delivery IDs (idempotency)

---

## Delivery Plan

## Week 1 — Backend Core
- Scaffold Go module with handler, detector, and models
- Implement deterministic 3-signal scoring
- Add unit tests for detector + risk classifier
- Add minimal schema definitions in `schema/postgres.sql`

## Week 2 — Integration Quality
- Add webhook payload contract validation
- Add repository layer + simple persistence
- Add structured logging and error responses
- Add smoke test script / sample payload fixtures

## Week 3 — Productization
- Add GitHub check-run adapter skeleton
- Add local docker-compose for postgres + service
- Tighten docs (`QUICK_START`, API examples)
- Final hardening, bug fixes, and demo readiness

---

## Definition of Done (for MVP)

- `go test ./...` passes
- Service starts locally and responds on `/health`
- Sample payload to `/analyze` returns deterministic report JSON
- README includes exact quickstart for MVP service
- No P0/P1 known defects open for demo path
