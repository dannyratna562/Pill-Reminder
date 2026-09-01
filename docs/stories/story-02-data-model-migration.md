# User Story: Pill Schedule Data Model & Migration

## Story
As the backend, I need a `pill_schedule` table and corresponding Go struct,
so that later stories can create, read, update, and delete pill schedules.

## Context
- Depends on [[story-01-family-pairing]] for `parent_id` to be meaningful,
  but can be built in parallel if `parent_id` is treated as an opaque
  foreign key for now.
- This story is schema + struct only — no API endpoints yet.

## Acceptance Criteria
- [ ] Migration creates `pill_schedule` table:
      `id (uuid, pk)`, `parent_id (fk)`, `pill_name (text)`,
      `times (array of time-of-day, repeats daily)`, `active (bool)`,
      `created_at`, `updated_at`.
- [ ] Go struct + basic model methods (e.g. validation of `times` format)
      exist in the domain layer.
- [ ] Migration is reversible (down migration included).
- [ ] Migration runs cleanly against a fresh database.
- [ ] Unit tests cover model validation (e.g. rejecting empty `times`,
      invalid time format).

## Out of Scope
- Any HTTP endpoint (create/list/update/delete come in later stories)
- Timezone handling logic beyond storing time-of-day values correctly
  (full DST/timezone edge cases are covered in the scheduling story, but
  make sure the column type won't make that harder later)

## Notes for Implementation
- Follow the migration tool/convention already used in the repo — check
  CLAUDE.md or existing migration files before introducing a new one.
- `times` as an array of daily times (not full timestamps) — a schedule is
  "8am and 8pm every day," not tied to a specific date.
