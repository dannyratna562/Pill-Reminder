-- pill_schedules stores recurring daily pill reminders.
--
-- Why `times` is TIME(0)[] and not TIMESTAMPTZ or TIMETZ:
--   A schedule entry like "8am and 8pm every day" is a recurring wall-clock
--   time-of-day, not an absolute instant in time, so TIMESTAMPTZ (which
--   anchors to a specific date) is the wrong shape. TIMETZ is also wrong:
--   it stores a *fixed* UTC offset captured at write time, which silently
--   breaks across DST transitions and when the family changes timezone —
--   the Postgres docs themselves discourage TIMETZ for this reason. Instead
--   we store a bare wall-clock time and let the Parent app do the
--   local-time conversion on-device using its own tz database at
--   notification-scheduling time (not at write time). This also means a
--   future per-parent `timezone TEXT` column can be added later without a
--   data migration, since `times` never needs to change value when a
--   parent's timezone does. Second precision (TIME(0)) is used because
--   sub-second precision is meaningless for a reminder time.
--
-- Why `parent_id` has no foreign key:
--   There is no parent/user table in this schema yet. Per
--   internal/domain/identity.go's device-supplied identity model,
--   pairing_codes.parent_id and family_links.parent_id are both plain
--   unconstrained UUID NOT NULL columns with no backing table to reference.
--   A FK to family_links(parent_id) isn't legal today either — it would
--   require a UNIQUE(parent_id) constraint there, which is out of scope for
--   this story and would incorrectly forbid a parent from having more than
--   one linked child (multi-parent/multi-child households).
--
-- Why there's no updated_at trigger:
--   This is intentional, not an oversight. Stories 05 (update) and 06
--   (delete/deactivate) will set `updated_at = now()` explicitly in their
--   own sqlc queries, so the write path stays visible in one place instead
--   of behind a trigger.
-- pill_schedules_times_are_valid backs the pill_schedules_times_minute_aligned
-- CHECK constraint below. Postgres CHECK constraints cannot contain
-- subqueries directly (even ones built from unnest()), so the per-element
-- validation has to live in a function instead of an inline `EXISTS (SELECT
-- ... FROM unnest(times) ...)` expression. It exists purely to keep the DB
-- in sync with the same two rules internal/store/pill_schedule_convert.go's
-- toDomainTimeOfDay enforces on read: every element must be minute-aligned
-- (no seconds) and strictly less than 24:00:00. Without this, a row written
-- via direct SQL (bypassing domain validation) could pass the DB but fail
-- every subsequent read, including story-04's list endpoint.
CREATE FUNCTION pill_schedules_times_are_valid(times TIME(0)[])
RETURNS boolean
LANGUAGE sql
IMMUTABLE
PARALLEL SAFE
AS $$
    SELECT NOT EXISTS (
        SELECT 1 FROM unnest(times) AS t
        WHERE t >= '24:00:00'::time OR date_trunc('minute', t) <> t
    );
$$;

CREATE TABLE pill_schedules (
    id         UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    parent_id  UUID NOT NULL,
    pill_name  TEXT NOT NULL,
    times      TIME(0)[] NOT NULL,
    active     BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT now(),

    CONSTRAINT pill_schedules_pill_name_not_blank  CHECK (btrim(pill_name) <> ''),
    CONSTRAINT pill_schedules_pill_name_len        CHECK (char_length(pill_name) <= 100),
    CONSTRAINT pill_schedules_times_cardinality    CHECK (cardinality(times) BETWEEN 1 AND 24),
    CONSTRAINT pill_schedules_times_no_nulls       CHECK (array_position(times, NULL) IS NULL),
    CONSTRAINT pill_schedules_times_minute_aligned CHECK (pill_schedules_times_are_valid(times))
);

CREATE INDEX idx_pill_schedules_parent_active ON pill_schedules (parent_id) WHERE active;
