# Contributing to Smart Doc API (Backend)

Thank you for contributing. This repository is maintained as an open-source, extensible backend platform for auth, RBAC, document management, and RAG workflows.

## Table of Contents

1. Prerequisites
2. Local Setup
3. Development Workflow
4. Run, Test, and Debug
5. Engineering Standards
6. Adding New Features or Modules
7. Pull Request Checklist
8. Communication and Collaboration

## 1. Prerequisites

- Go 1.25+
- PostgreSQL 15+ (pgvector recommended)
- MinIO-compatible object storage
- Gemini API key for RAG features

## 2. Local Setup

1. Fork this repository and clone your fork.
2. Create `./.env` in project root.

```env
# Database
DATABASE_DSN=host=localhost user=postgres password=postgres dbname=postgres port=5432 sslmode=disable
REQUIRE_PGVECTOR=false

# Auth
JWT_SECRET=replace-with-a-strong-secret

# Object storage (MinIO)
MINIO_ENDPOINT=http://127.0.0.1:9000
MINIO_ACCESS_KEY=minioadmin
MINIO_SECRET_KEY=minioadmin
MINIO_BUCKET=docfiles
MINIO_USE_SSL=false

# RAG rate limiting
RAG_QUERY_RATE_LIMIT=30
RAG_QUERY_WINDOW_SECONDS=60

# Gemini RAG provider
GEMINI_API_KEY=replace-with-gemini-api-key
GEMINI_EMBED_MODEL=gemini-embedding-001
GEMINI_EMBED_DIM=768
GEMINI_CHAT_MODEL=gemini-2.5-flash
GEMINI_CHAT_FALLBACK_MODELS=gemini-2.5-flash-lite,gemini-2.0-flash,gemini-1.5-flash
```

3. Start PostgreSQL and MinIO.
4. Run the API.

```bash
go run main.go
```

5. Open Swagger docs:

- [http://localhost:8080/swagger/index.html](http://localhost:8080/swagger/index.html)

### Seeded Default Admin

On first boot, the seed process creates:

- Email: `admin@example.com`
- Password: `admin@1234`

Change this in non-local environments.

## 3. Development Workflow

### Branch Strategy

- Base branch: `main`
- One branch per feature/fix
- Keep branches short-lived and rebased with `main`

Branch naming:

- `feat/<scope>-<short-description>`
- `fix/<scope>-<short-description>`
- `docs/<scope>-<short-description>`
- `refactor/<scope>-<short-description>`
- `chore/<scope>-<short-description>`

Examples:

- `feat/rag-metadata-filtering`
- `fix/jwt-permission-hydration`
- `refactor/document-service-error-handling`

### Commit Convention (Conventional Commits)

Use:

```text
<type>(<scope>): <summary>
```

Allowed types: `feat`, `fix`, `docs`, `refactor`, `test`, `chore`, `build`, `ci`, `perf`.

Examples:

- `feat(rag): add category-aware similarity weighting`
- `fix(auth): reject expired bearer tokens`
- `docs(api): document permission headers`

## 4. Run, Test, and Debug

### Core Commands

```bash
go run main.go
go test ./...
go vet ./...
```

### Formatting

```bash
gofmt -w main.go internal/**/*.go pkg/**/*.go
```

If your shell does not expand `**`, run `gofmt` on explicit file paths.

### Current Test Status

- This repository currently has no first-party `_test.go` files in `internal/` or `pkg/`.
- Contributors should add tests with new logic changes and include manual verification steps until baseline test coverage is established.

### Local Debugging

- Use `go run main.go` and inspect logs for:
  - DB connection/migration
  - MinIO initialization
  - RAG provider readiness
- Validate critical endpoints:
  - `POST /api/auth/login`
  - `POST /api/rag/query`
  - `GET /api/rag/source/:id`
  - CRUD endpoints under `/api/users`, `/api/roles`, `/api/permissions`, `/api/documents`
- Use Swagger UI for request/response validation.

## 5. Engineering Standards

### Code Style

- Follow idiomatic Go and keep package boundaries clear.
- Keep handlers thin; put business rules in services.
- Repositories should only contain persistence concerns.
- Avoid global mutable state unless explicitly synchronized.
- Return typed/service-layer errors and map them in handlers.

### Naming and Structure

- Package names: lowercase, short, domain-oriented
- Exported identifiers: `PascalCase`
- Unexported identifiers: `camelCase`
- Keep files focused by responsibility (handler/service/repository separation)

### Backward Compatibility

- Do not break existing endpoint contracts without documenting migration.
- Keep response shapes stable for FE consumers.
- For any breaking API change, add versioning strategy note in PR.

## 6. Adding New Features or Modules

Use this layering order for all new backend capabilities:

1. **Model**: add/extend entities in `internal/models/*`.
2. **Repository**: data access contract + implementation in `internal/repositories/*`.
3. **Service**: business logic in `internal/services/*`.
4. **Handler**: request/response mapping in `internal/handlers/*`.
5. **Route**: endpoint registration and middleware wiring in `internal/routes/*`.
6. **Docs**: update README + Swagger annotations when contracts change.

### Middleware and Permissions

- Route-level permission checks must use existing middleware primitives:
  - `AuthMiddleware`
  - `PermissionsMiddleware`
  - `RequirePermission`
  - `RequireHeaderPermission`
- New protected routes must define explicit permission names and corresponding seed updates.

### RAG Extensions

When changing RAG behavior:

- Keep chunking, embedding, and generation components interface-driven.
- Preserve graceful degradation when Gemini is unavailable.
- Include performance notes for large-document indexing changes.

## 7. Pull Request Checklist

Before requesting review, ensure all items are complete:

- [ ] Branch name follows strategy
- [ ] Commits follow Conventional Commits
- [ ] `go test ./...` passes
- [ ] `go vet ./...` passes
- [ ] Code is formatted with `gofmt`
- [ ] API contract changes documented
- [ ] Swagger/docs updated where applicable
- [ ] Backward compatibility impact assessed and documented
- [ ] Manual verification steps included for changed endpoints
- [ ] No secrets or environment-specific credentials committed

## 8. Communication and Collaboration

### Issues

- Use issue templates for bug reports and feature requests.
- One issue per problem/feature.
- Include clear reproduction or scope boundaries.

### Design Discussions

Open an issue/discussion before implementation if your change includes:

- New cross-cutting dependency
- New architectural pattern
- Breaking API behavior
- Changes in permission model semantics

### Review Expectations

- Keep review comments factual and actionable.
- Resolve review threads explicitly.
- Prefer incremental commits over force-push rewrites unless requested by maintainer.

---

By contributing, you help keep this API stable, secure, and extensible for integrators.
