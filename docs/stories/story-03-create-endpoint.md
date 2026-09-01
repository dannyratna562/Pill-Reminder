# User Story: Create Pill Schedule Endpoint

## Story
As the Child app, I want to call an API to create a pill schedule, so that a
new pill reminder is stored for the linked parent.

## Context
- Depends on [[story-02-data-model-migration]] for the table/struct, and
  [[story-01-family-pairing]] for a valid `parent_id` to write against.

## Acceptance Criteria
- [ ] `POST /pill-schedules` accepts `{ parent_id, pill_name, times[] }`.
- [ ] Validates: `parent_id` exists and is linked to the requesting child,
      `pill_name` non-empty, `times` non-empty and correctly formatted.
- [ ] Returns the created schedule (with generated `id`) on success.
- [ ] Returns clear validation errors (400) for bad input.
- [ ] Returns 403/404 if `parent_id` isn't linked to the requesting child.
- [ ] Unit tests cover: happy path, missing fields, invalid time format,
      unauthorized parent_id.

## Out of Scope
- List, update, delete (separate stories)
- Triggering the push notification to the Parent app (covered in
  [[story-07-push-notify]])

## Notes for Implementation
- This is the first write endpoint — establish the request/response
  envelope and error format here, since later endpoints (update, delete)
  should follow the same pattern.
