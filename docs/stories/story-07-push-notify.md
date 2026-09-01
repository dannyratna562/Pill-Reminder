# User Story: Push Notify on Schedule Change

## Story
As the Parent app, I want to be notified whenever my pill schedule changes,
so that I can resync and update my local reminder notifications promptly.

## Context
- Depends on [[story-03-create-endpoint]], [[story-05-update-endpoint]],
  and [[story-06-delete-endpoint]] all existing, since this hooks into each.
- This is the backend half of the sync story from the architecture
  diagram — the Parent app's local notification scheduling is a separate,
  client-side story.

## Acceptance Criteria
- [ ] After a successful create, update, or delete, the backend sends a
      push notification (FCM for Android, APNs for iOS) to the linked
      parent device.
- [ ] Push payload is a lightweight "schedule changed, please resync"
      signal — not the full schedule data (Parent app calls
      [[story-04-list-endpoint]] to get current state).
- [ ] Push failures (e.g. device token invalid/expired) are logged and do
      not fail the underlying create/update/delete request.
- [ ] Unit tests cover: push triggered on each of create/update/delete,
      push failure doesn't block the API response.

## Out of Scope
- Registering/storing device push tokens (assumes a token registration
  mechanism exists or is a separate prerequisite story if it doesn't —
  flag this if missing)
- Retry/backoff logic for failed pushes (log and move on for MVP)
- The Parent app's local notification scheduling logic itself

## Notes for Implementation
- If device token registration doesn't exist yet in this repo, stop and
  flag it — this story can't be completed without a `device_token` stored
  somewhere against the parent record.
