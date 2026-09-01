# User Story: List Pill Schedules Endpoint

## Story
As the Child app or Parent app, I want to fetch all pill schedules for a
parent, so that I can display the current reminders or resync the local
notification schedule.

## Context
- Depends on [[story-02-data-model-migration]] and [[story-01-family-pairing]].
- Used by both apps: Child app to show what's configured, Parent app to
  resync after a push notification.

## Acceptance Criteria
- [ ] `GET /pill-schedules?parent_id=` returns all active schedules for
      that parent.
- [ ] Returns 403/404 if the requester isn't linked to that `parent_id`.
- [ ] Empty list (not an error) when no schedules exist yet.
- [ ] Response includes enough fields for the Parent app to schedule local
      notifications without a second call (`id`, `pill_name`, `times`).
- [ ] Unit tests cover: happy path, empty result, unauthorized access.

## Out of Scope
- Pagination (family-scale data volume doesn't need it for MVP)
- Filtering by active/inactive (return active only for now)

## Notes for Implementation
- Keep the response shape consistent with the create endpoint's schedule
  object from [[story-03-create-endpoint]] — the Parent app will map both
  responses to the same local model.
