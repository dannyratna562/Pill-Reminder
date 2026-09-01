package store

import (
	"fmt"

	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pillreminder/backend/internal/domain"
	"github.com/pillreminder/backend/internal/store/sqlcgen"
)

const (
	microsecondsPerMinute = int64(60 * 1_000_000)
	microsecondsPerDay    = int64(24 * 60 * 60 * 1_000_000)
)

// toDomainPillSchedule maps a sqlcgen.PillSchedule row into the domain
// representation, validating and converting each stored TIME value.
func toDomainPillSchedule(row sqlcgen.PillSchedule) (domain.PillSchedule, error) {
	times := make([]domain.TimeOfDay, 0, len(row.Times))
	for i, pt := range row.Times {
		t, err := toDomainTimeOfDay(pt)
		if err != nil {
			return domain.PillSchedule{}, fmt.Errorf("times[%d]: %w", i, err)
		}
		times = append(times, t)
	}

	return domain.PillSchedule{
		ID:        row.ID,
		ParentID:  row.ParentID,
		PillName:  row.PillName,
		Times:     times,
		Active:    row.Active,
		CreatedAt: row.CreatedAt,
		UpdatedAt: row.UpdatedAt,
	}, nil
}

// toDomainTimeOfDay converts a Postgres TIME value into a domain.TimeOfDay.
// It fails loudly (rather than silently truncating) if the value is NULL,
// outside the valid [0, 24h) range (Postgres TIME technically permits the
// literal '24:00:00'), or not aligned to a whole minute — the latter two
// are only reachable via direct SQL writes that bypass domain validation.
func toDomainTimeOfDay(t pgtype.Time) (domain.TimeOfDay, error) {
	if !t.Valid {
		return domain.TimeOfDay{}, fmt.Errorf("%w: null time value", domain.ErrInvalidTimeFormat)
	}
	if t.Microseconds < 0 || t.Microseconds >= microsecondsPerDay {
		return domain.TimeOfDay{}, fmt.Errorf("%w: time value out of range", domain.ErrInvalidTimeFormat)
	}
	if t.Microseconds%microsecondsPerMinute != 0 {
		return domain.TimeOfDay{}, fmt.Errorf("%w: time value not minute-aligned", domain.ErrInvalidTimeFormat)
	}

	totalMinutes := int(t.Microseconds / microsecondsPerMinute)
	return domain.NewTimeOfDay(totalMinutes/60, totalMinutes%60)
}

// toPgTimes converts domain TimeOfDay values into pgtype.Time values
// suitable for writing to the pill_schedules.times column.
func toPgTimes(times []domain.TimeOfDay) []pgtype.Time {
	out := make([]pgtype.Time, len(times))
	for i, t := range times {
		out[i] = pgtype.Time{
			Microseconds: int64(t.MinutesSinceMidnight()) * 60 * 1_000_000,
			Valid:        true,
		}
	}
	return out
}
