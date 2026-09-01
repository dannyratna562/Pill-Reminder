# User Story: Delete Pill Schedule Endpoint

## Story
As the Child app, I want to delete a pill schedule, so that it stops
reminding the parent once a medication is no longer needed.

## Context
- Depends on [[story-02-data-model-migration]] and the same auth pattern as
  [[story-03-create-endpoint]].

## Acceptance Criteria
- [ ] `DELETE /pill-schedules/{id}` removes (or soft-deletes / sets
      `active = false` — pick one and document it) the schedule.
- [ ] Validates ownership: only the linked child can delete.
- [ ] Returns 404 if not found, 403 if not authorized.
- [ ] Returns success confirmation on delete.
- [ ] Unit tests cover: happy path, not found, unauthorized, double-delete.

## Out of Scope
- Bulk delete
- Triggering resync push (covered in [[story-07-push-notify]])

## Notes for Implementation
- Prefer soft delete (`active = false`) over hard delete — keeps history
  and makes the push-notify story simpler (it can just push "changed"
  rather than needing to distinguish delete from deactivate). Flag if hard
  delete is preferred instead, and note the tradeoff in the PR description.
