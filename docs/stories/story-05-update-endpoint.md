# User Story: Update Pill Schedule Endpoint

## Story
As the Child app, I want to edit an existing pill schedule (name or times),
so that changes are reflected without creating a duplicate entry.

## Context
- Depends on [[story-02-data-model-migration]] and [[story-03-create-endpoint]]
  (same validation and auth rules apply).

## Acceptance Criteria
- [ ] `PUT /pill-schedules/{id}` accepts updated `pill_name` and/or `times`.
- [ ] Validates ownership: only the linked child (via parent_id) can update.
- [ ] Returns the updated schedule on success.
- [ ] Returns 404 if the schedule doesn't exist, 403 if not authorized.
- [ ] Returns the same validation errors as create for bad input.
- [ ] Unit tests cover: happy path, partial update (name only, times only),
      not found, unauthorized.

## Out of Scope
- Bulk update of multiple schedules at once
- Triggering resync push (covered in [[story-07-push-notify]])

## Notes for Implementation
- Reuse the validation logic from [[story-03-create-endpoint]] rather than
  duplicating it — factor it into a shared function if it isn't already.
