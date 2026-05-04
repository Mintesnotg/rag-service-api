# Backend API Platform Blueprint

This document defines how `go-api` should evolve into an extensible backend platform for external developers and integrators.

## 1. Modular Architecture Plan

## 1.1 Current Strength

The current layered architecture (`handlers -> services -> repositories`) is a strong base and should be formalized into explicit module contracts.

## 1.2 Target Module Boundaries

```text
internal/
  platform/
    contracts/
    errors/
    events/
  modules/
    auth/
    permissions/
    documents/
    rag/
      providers/
      indexing/
      retrieval/
  transport/
    http/
      handlers/
      routes/
      middleware/
```

Rules:

- `platform/contracts`: stable interfaces used by modules.
- `modules/*`: business capabilities owned by domain modules.
- `transport/http/*`: adapters only; no core business rules.

## 2. Plugin / Extension Strategy

## 2.1 Provider Plugins (RAG)

Introduce provider plugins for:

- Embedding providers
- LLM generation providers
- Optional re-rankers

Provider contract (conceptual):

```go
type EmbedProvider interface {
    Name() string
    Embed(ctx context.Context, text string) ([]float64, error)
}

type ChatProvider interface {
    Name() string
    GenerateAnswer(ctx context.Context, prompt string, contexts []string) (string, error)
}
```

## 2.2 Registry and Selection

- Register providers through a runtime registry.
- Select provider by config (`RAG_EMBED_PROVIDER`, `RAG_CHAT_PROVIDER`).
- Fallback chain remains supported by ordered provider list.

## 2.3 Safe Extension Rules

- Extensions must not mutate shared global state unsafely.
- Extension initialization failures should degrade gracefully where possible.
- All extension errors should map to structured API errors.

## 3. API Design Improvements for Extensibility

## 3.1 Versioned API Contracts

- Keep current routes under `/api/*` as `v1`.
- Introduce `/api/v2/*` for future breaking changes.
- Provide deprecation headers and migration guides.

## 3.2 Standard Response Envelope

Adopt consistent structure for success and errors:

```json
{
  "data": {},
  "meta": {"request_id": "..."},
  "error": null
}
```

Error envelope:

```json
{
  "data": null,
  "meta": {"request_id": "..."},
  "error": {
    "code": "permission_denied",
    "message": "...",
    "details": {}
  }
}
```

## 3.3 External Integration Capabilities

- Cursor-based pagination for list endpoints.
- Idempotency keys for mutating endpoints.
- Webhook events for document and indexing lifecycle.
- Structured audit logs with request correlation IDs.

## 4. External Developer Integration Guide (Target)

External developers should be able to:

1. Authenticate with JWT and documented claims.
2. Query RAG endpoints with optional metadata filters.
3. Subscribe to document/index lifecycle webhooks.
4. Integrate with provider-specific behavior through configuration only.

## 5. Operational Guardrails

- Add OpenAPI drift checks in CI.
- Add compatibility tests for v1 endpoint response shapes.
- Add extension-provider conformance tests.
- Publish explicit support policy for API versions.

## 6. Rollout Plan

1. Extract provider registry and migrate Gemini implementation into provider module.
2. Add structured error envelope and request ID propagation.
3. Add webhook dispatcher for document/index events.
4. Add API versioning policy and CI contract checks.

This approach keeps existing consumers working while turning the backend into a scalable platform for contributors and integrators.
