# CLAUDE.md — Pill Reminder Backend (Go)

> Read automatically by the coding agent before it touches code.
> Update this file every time an agent's output drifts from what you'd
> actually write — that's the whole point of it.

## Context

- Backend for the pill reminder app: Child app sets pill schedules, this
  service stores them and notifies the Parent app to resync local
  notifications. See the architecture: Child app → Backend API (Go) →
  Postgres, and Backend → Push service → Parent app → local alarm.
- MVP scope only: family pairing, pill schedule CRUD, push-on-change.
  No confirmation flow, no missed-dose escalation, no doctor comms.
- Implementation order (respect dependencies — do not build out of order):
  1. story-01-family-pairing
  2. story-02-data-model-migration
  3. story-03-create-endpoint
  4. story-04-list-endpoint
  5. story-05-update-endpoint
  6. story-06-delete-endpoint
  7. story-07-push-notify

## Build & Test Commands

```
# Get dependencies
go mod tidy

# Build
go build ./...

# Run all tests
go test ./... -v

# Run a single package's tests
go test ./internal/<package>/... -v

# Vet / static checks
go vet ./...
[add golangci-lint run if configured]

# Run migrations (golang-migrate)
migrate -path ./migrations -database $DATABASE_URL up

# Run the service locally
go run ./cmd/api
```

## Architecture Conventions

- Package layout:
  - `/cmd/api` — main entrypoint
  - `/internal/domain` — core types, validation, business rules (no HTTP
    or DB imports here)
  - `/internal/api` — HTTP handlers, routing, request/response mapping
  - `/internal/store` — Postgres access (sqlc-generated queries)
  - `/internal/push` — FCM/APNs integration (story-07)
  - `/migrations` — SQL migration files
- HTTP framework: net/http + chi router (stdlib-adjacent, minimal
  dependencies, easy for an agent to reason about)
- Data access: sqlc-generated queries only, no raw SQL in handlers, no ORM
- Config: env vars via `/internal/config`, no hardcoded values, no secrets
  in code
- Logging: structured logging via `log/slog`, no fmt.Println in production
  code paths
- Dependency injection: constructor injection, no DI framework

## API Conventions (apply to every story from story-03 onward)

- Request/response envelope: [fill in once story-03 establishes it, e.g.
  `{ "data": {...} }` on success, `{ "error": { "code", "message" } }` on
  failure]
- Status codes: 400 for validation errors, 403 for authorization failures
  (not linked to the parent_id), 404 for not-found, 200/201 for success.
- Every write endpoint validates the requester is linked to the
  `parent_id` in question via the `family_link` table from story-01 —
  this check must not be duplicated ad hoc per handler; factor it into a
  shared middleware or helper.

## Data & Migrations

- Migration tool: golang-migrate
- Every migration needs a down migration.
- `times` on `pill_schedule` stores time-of-day values (repeats daily),
  not full timestamps — see story-02 for the exact reasoning.
- Soft delete (`active = false`) is the default deletion pattern unless a
  story says otherwise — see story-06.

## Push Notifications (story-07)

- Push payload is a lightweight "schedule changed" signal, not the full
  schedule — the Parent app always resyncs via the list endpoint after a
  push, per story-04 and story-07.
- Push failures must be logged, never fail the underlying API request.
- Device token storage: [fill in — where/how tokens are registered; if
  this doesn't exist yet, story-07 should stop and flag it as a missing
  prerequisite rather than guessing]

## Testing Rules

- New logic needs unit tests; handler logic needs at least one happy-path
  and one error-path test.
- Auth/ownership checks (family_link validation) must have explicit test
  coverage on every write endpoint — this is the area most likely to be
  under-tested by default.
- Do not skip or weaken an existing test to make CI pass — fix the cause
  or flag it for a human.

## PR Rules

- The agent's job ends at "PR opened." Never merge, never push directly
  to main, never request merge permissions.
- PR description must reference the story file it implements and list
  anything from the story's "Out of Scope" section that was deliberately
  left out.

## Things NOT to Touch

- [Any shared infra, CI config, or other services in this repo that are
  off-limits for these stories]

## Definition of Done

- [ ] Acceptance criteria from the story file are all met or explicitly
      flagged as blocked
- [ ] `go vet` and lint clean
- [ ] `go test ./...` passing
- [ ] Auth/ownership checks tested (for write endpoints)
- [ ] PR opened (not merged) with description per PR Rules above

## Team Preferences

- [e.g. "Prefer explicit error returns over panics"]
- [e.g. "No new third-party packages without a comment explaining why"]
