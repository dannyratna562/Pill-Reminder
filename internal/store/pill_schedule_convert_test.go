package store

import (
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/pillreminder/backend/internal/domain"
	"github.com/pillreminder/backend/internal/store/sqlcgen"
)

func TestTimeOfDay_PgTimeRoundTrip(t *testing.T) {
	values := []string{"00:00", "08:00", "12:30", "20:30", "23:59"}
	var times []domain.TimeOfDay
	for _, s := range values {
		tt, err := domain.ParseTimeOfDay(s)
		if err != nil {
			t.Fatalf("ParseTimeOfDay(%q) error = %v", s, err)
		}
		times = append(times, tt)
	}

	pgTimes := toPgTimes(times)
	if len(pgTimes) != len(times) {
		t.Fatalf("toPgTimes() len = %d, want %d", len(pgTimes), len(times))
	}

	for i, pt := range pgTimes {
		got, err := toDomainTimeOfDay(pt)
		if err != nil {
			t.Fatalf("toDomainTimeOfDay(%v) error = %v", pt, err)
		}
		if got != times[i] {
			t.Errorf("round-trip[%d] = %v, want %v", i, got, times[i])
		}
		if got.String() != values[i] {
			t.Errorf("round-trip[%d].String() = %q, want %q", i, got.String(), values[i])
		}
	}
}

func TestToDomainTimeOfDay_Errors(t *testing.T) {
	tests := []struct {
		name string
		in   pgtype.Time
	}{
		{"invalid/null", pgtype.Time{Valid: false}},
		{"exactly 24h", pgtype.Time{Microseconds: 24 * 60 * 60 * 1_000_000, Valid: true}},
		{"not minute-aligned", pgtype.Time{Microseconds: 8*60*60*1_000_000 + 30*1_000_000, Valid: true}}, // 08:00:30
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := toDomainTimeOfDay(tt.in)
			if !errors.Is(err, domain.ErrInvalidTimeFormat) {
				t.Errorf("toDomainTimeOfDay(%v) error = %v, want ErrInvalidTimeFormat", tt.in, err)
			}
		})
	}
}

func TestToDomainPillSchedule(t *testing.T) {
	id := uuid.New()
	parentID := uuid.New()
	createdAt := time.Date(2026, 1, 1, 8, 0, 0, 0, time.UTC)
	updatedAt := time.Date(2026, 1, 2, 9, 0, 0, 0, time.UTC)

	row := sqlcgen.PillSchedule{
		ID:       id,
		ParentID: parentID,
		PillName: "Aspirin",
		Times: []pgtype.Time{
			{Microseconds: 8 * 60 * 60 * 1_000_000, Valid: true},       // 08:00
			{Microseconds: (20*60 + 30) * 60 * 1_000_000, Valid: true}, // 20:30
		},
		Active:    true,
		CreatedAt: createdAt,
		UpdatedAt: updatedAt,
	}

	got, err := toDomainPillSchedule(row)
	if err != nil {
		t.Fatalf("toDomainPillSchedule() error = %v", err)
	}

	if got.ID != id {
		t.Errorf("ID = %v, want %v", got.ID, id)
	}
	if got.ParentID != parentID {
		t.Errorf("ParentID = %v, want %v", got.ParentID, parentID)
	}
	if got.PillName != "Aspirin" {
		t.Errorf("PillName = %q, want %q", got.PillName, "Aspirin")
	}
	if !got.Active {
		t.Errorf("Active = false, want true")
	}
	if !got.CreatedAt.Equal(createdAt) {
		t.Errorf("CreatedAt = %v, want %v", got.CreatedAt, createdAt)
	}
	if !got.UpdatedAt.Equal(updatedAt) {
		t.Errorf("UpdatedAt = %v, want %v", got.UpdatedAt, updatedAt)
	}
	if len(got.Times) != 2 || got.Times[0].String() != "08:00" || got.Times[1].String() != "20:30" {
		t.Errorf("Times = %v, want [08:00 20:30]", got.Times)
	}
}

func TestToDomainPillSchedule_InvalidTimePropagates(t *testing.T) {
	row := sqlcgen.PillSchedule{
		ID:       uuid.New(),
		ParentID: uuid.New(),
		PillName: "Aspirin",
		Times:    []pgtype.Time{{Valid: false}},
		Active:   true,
	}

	_, err := toDomainPillSchedule(row)
	if !errors.Is(err, domain.ErrInvalidTimeFormat) {
		t.Errorf("toDomainPillSchedule() error = %v, want ErrInvalidTimeFormat", err)
	}
}
