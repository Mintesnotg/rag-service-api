# Backend API Issue Backlog

This backlog is curated for maintainers to create high-quality GitHub issues quickly.

## Recommended Label Set

- `enhancement`: user-facing or integrator-facing feature work
- `good first issue`: narrow scope, low risk
- `help wanted`: community contributions encouraged
- `area:rag`, `area:auth`, `area:docs`, `area:infra`, `area:api`: domain labels
- `priority:p1`, `priority:p2`, `priority:p3`: priority labels

---

## API-001: Provider Registry for Embedding and LLM Backends

**Labels:** `enhancement`, `area:rag`, `area:api`, `priority:p1`, `help wanted`

### Description

Replace hard-wired Gemini initialization with a provider registry that allows selecting embedding and chat providers by configuration (for example Gemini now, OpenAI/other providers later) without changing core RAG orchestration.

### Acceptance Criteria

- [ ] Introduce provider interfaces and a provider factory/registry.
- [ ] Existing Gemini flow works unchanged through the registry.
- [ ] Provider selection is controlled by env/config.
- [ ] Clear error reporting when provider config is missing or invalid.
- [ ] Contributor docs explain how to add a new provider.

### Technical Hints

- Keep `Embedder` and `LLM` interfaces as stable contracts.
- Move provider-specific env parsing into provider packages.
- Register provider constructors in one package-level registry.

---

## API-002: Asynchronous Indexing Queue with Retry and Dead-letter Handling

**Labels:** `enhancement`, `area:infra`, `area:rag`, `priority:p1`

### Description

Move document indexing from request-coupled background behavior to an explicit queue-based pipeline with retry policy and dead-letter state to improve reliability and observability.

### Acceptance Criteria

- [ ] Add job model/state for indexing tasks.
- [ ] Queue worker supports retry with bounded attempts.
- [ ] Failed jobs move to dead-letter state with error reason.
- [ ] Admin can query indexing status per document.
- [ ] Indexing side effects are idempotent.

### Technical Hints

- Start with database-backed queue for simplicity.
- Use status values (`pending`, `running`, `succeeded`, `failed`, `dead_letter`).
- Add structured logs containing document ID and job ID.

---

## API-003: Metadata-aware Similarity Search Filters

**Labels:** `enhancement`, `area:rag`, `area:api`, `priority:p2`, `help wanted`

### Description

Extend `/api/rag/query` to support structured filters (category, document owner, created range, visibility tags) so clients can scope retrieval without post-filtering.

### Acceptance Criteria

- [ ] Query endpoint accepts optional filter object.
- [ ] Repository similarity query applies filters server-side.
- [ ] API validation rejects unsupported filter keys.
- [ ] Existing clients remain compatible when filters are omitted.
- [ ] Swagger docs updated with filter schema and examples.

### Technical Hints

- Introduce a filter DTO in handler/service boundaries.
- Keep filter parsing separate from vector similarity logic.
- Add database indexes for frequently filtered columns.

---

## API-004: Event Webhooks for Document and Index Lifecycle

**Labels:** `enhancement`, `area:api`, `area:infra`, `priority:p2`

### Description

Publish signed webhook events for document lifecycle and indexing status so external systems can integrate without polling.

### Acceptance Criteria

- [ ] Emit events for document create/update/delete.
- [ ] Emit events for indexing started/succeeded/failed.
- [ ] Support HMAC signature validation for subscribers.
- [ ] Add retry and backoff for delivery failures.
- [ ] Document event payload schema and signature verification.

### Technical Hints

- Create an event dispatcher abstraction decoupled from handlers.
- Persist delivery attempts for operational audit.
- Include versioned `event_type` and `event_version` fields.

---

## API-005: OpenAPI Contract Guardrail in CI

**Labels:** `enhancement`, `area:docs`, `area:dx`, `good first issue`, `priority:p3`

### Description

Add CI validation that regenerates Swagger/OpenAPI artifacts and fails if generated files are out of date, preventing undocumented API drift.

### Acceptance Criteria

- [ ] Add script to regenerate docs deterministically.
- [ ] Add CI step that detects swagger drift.
- [ ] Document local command for contributors.
- [ ] Include contribution guidance for API contract changes.

### Technical Hints

- Use `swag init -g main.go -o docs` as baseline generator command.
- Compare generated files to repository state in CI.
- Keep docs generation in a single script entrypoint.
