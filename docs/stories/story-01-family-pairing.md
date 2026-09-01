# User Story: Family Pairing (Child ↔ Parent Link)

## Story
As a family setting up the app, I want the Child app and Parent app to be
linked to each other, so that pill schedules created in the Child app are
scoped to the correct parent device.

## Context
- Prerequisite for all other backend stories — schedules must be scoped to
  a `parent_id`, and only the linked child should be able to write to it.
- MVP: one child, one parent per link is enough. Don't build multi-child or
  multi-parent support yet.

## Acceptance Criteria
- [ ] Backend generates a short-lived pairing code on request (e.g. from the
      Parent app, on first launch).
- [ ] Child app can submit the pairing code to link itself to that parent.
- [ ] A `family_link` record is created: `{ id, child_id, parent_id, created_at }`.
- [ ] Pairing code expires after a reasonable window (e.g. 10 minutes) and is
      single-use.
- [ ] Attempting to use an expired or already-used code returns a clear error.
- [ ] Unit tests cover: code generation, successful pairing, expired code,
      reused code.

## Out of Scope
- Un-pairing / re-pairing flow (assume happy path for MVP)
- Multiple children per parent or multiple parents per child
- Any UI — this story is backend only (API + data model)

## Notes for Implementation
- Keep the pairing code mechanism simple (e.g. 6-digit numeric code) — this
  is a low-stakes internal family flow, not a security-critical auth system.
- Store `child_id` / `parent_id` as whatever identity concept the backend
  already uses (device ID, user ID) — flag if no identity system exists yet,
  since that would be a blocking dependency for this story.
