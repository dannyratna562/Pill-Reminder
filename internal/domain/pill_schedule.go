package domain

import (
	"fmt"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"github.com/google/uuid"
)

// MaxPillNameLen is the maximum number of characters allowed in a
// PillSchedule's PillName.
const MaxPillNameLen = 100

// MaxTimesPerSchedule is the maximum number of times-of-day a single
// PillSchedule may have.
const MaxTimesPerSchedule = 24

// TimeOfDay is a minute-granularity wall-clock time (e.g. "08:00"), not
// tied to any date or timezone. It can only be constructed through
// NewTimeOfDay or ParseTimeOfDay so that an in-memory TimeOfDay is always
// valid.
type TimeOfDay struct {
	hour   uint8
	minute uint8
}

// NewTimeOfDay validates hour (0-23) and minute (0-59) and returns the
// corresponding TimeOfDay.
func NewTimeOfDay(hour, minute int) (TimeOfDay, error) {
	if hour < 0 || hour > 23 || minute < 0 || minute > 59 {
		return TimeOfDay{}, ErrInvalidTimeFormat
	}
	return TimeOfDay{hour: uint8(hour), minute: uint8(minute)}, nil
}

// ParseTimeOfDay parses s as a strict zero-padded 24-hour "HH:MM" string
// (e.g. "08:00", "23:59"). Unlike time.Parse, this rejects any deviation
// from that exact shape: no seconds, no single-digit hours, no whitespace,
// no non-ASCII digits.
func ParseTimeOfDay(s string) (TimeOfDay, error) {
	if len(s) != 5 {
		return TimeOfDay{}, ErrInvalidTimeFormat
	}
	if s[2] != ':' {
		return TimeOfDay{}, ErrInvalidTimeFormat
	}
	hourTens, ok := digitValue(s[0])
	if !ok {
		return TimeOfDay{}, ErrInvalidTimeFormat
	}
	hourOnes, ok := digitValue(s[1])
	if !ok {
		return TimeOfDay{}, ErrInvalidTimeFormat
	}
	minuteTens, ok := digitValue(s[3])
	if !ok {
		return TimeOfDay{}, ErrInvalidTimeFormat
	}
	minuteOnes, ok := digitValue(s[4])
	if !ok {
		return TimeOfDay{}, ErrInvalidTimeFormat
	}

	hour := hourTens*10 + hourOnes
	minute := minuteTens*10 + minuteOnes

	if hour > 23 || minute > 59 {
		return TimeOfDay{}, ErrInvalidTimeFormat
	}

	return TimeOfDay{hour: uint8(hour), minute: uint8(minute)}, nil
}

// digitValue reports the numeric value of an ASCII '0'-'9' byte, or
// ok=false for anything else (including non-ASCII digits, which Go's byte
// indexing would not decode as single digits anyway).
func digitValue(b byte) (int, bool) {
	if b < '0' || b > '9' {
		return 0, false
	}
	return int(b - '0'), true
}

// ParseTimes parses each element of raw as a TimeOfDay, returning the
// first parse error encountered, wrapped with the offending index.
func ParseTimes(raw []string) ([]TimeOfDay, error) {
	times := make([]TimeOfDay, 0, len(raw))
	for i, s := range raw {
		t, err := ParseTimeOfDay(s)
		if err != nil {
			return nil, fmt.Errorf("times[%d]: %w", i, err)
		}
		times = append(times, t)
	}
	return times, nil
}

// Hour returns the hour component (0-23).
func (t TimeOfDay) Hour() int {
	return int(t.hour)
}

// Minute returns the minute component (0-59).
func (t TimeOfDay) Minute() int {
	return int(t.minute)
}

// MinutesSinceMidnight returns the number of minutes elapsed since 00:00.
func (t TimeOfDay) MinutesSinceMidnight() int {
	return int(t.hour)*60 + int(t.minute)
}

// String renders t as zero-padded "HH:MM", e.g. "08:00".
func (t TimeOfDay) String() string {
	return fmt.Sprintf("%02d:%02d", t.hour, t.minute)
}

// NormalizeTimes returns a new slice containing the distinct values of
// times sorted in ascending order. The input slice is never mutated.
func NormalizeTimes(times []TimeOfDay) []TimeOfDay {
	if len(times) == 0 {
		return []TimeOfDay{}
	}

	sorted := make([]TimeOfDay, len(times))
	copy(sorted, times)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].MinutesSinceMidnight() < sorted[j].MinutesSinceMidnight()
	})

	out := make([]TimeOfDay, 0, len(sorted))
	for i, t := range sorted {
		if i == 0 || t != out[len(out)-1] {
			out = append(out, t)
		}
	}
	return out
}

// PillSchedule represents a recurring daily reminder to take a pill at one
// or more times of day.
type PillSchedule struct {
	ID        uuid.UUID
	ParentID  uuid.UUID
	PillName  string
	Times     []TimeOfDay
	Active    bool
	CreatedAt time.Time
	UpdatedAt time.Time
}

// ValidatePillName reports whether name is a valid PillSchedule.PillName:
// non-empty after trimming whitespace, and no longer than MaxPillNameLen
// characters (counted on the untrimmed input).
func ValidatePillName(name string) error {
	if strings.TrimSpace(name) == "" {
		return ErrEmptyPillName
	}
	if utf8.RuneCountInString(name) > MaxPillNameLen {
		return ErrPillNameTooLong
	}
	return nil
}

// ValidateTimes reports whether times is a valid PillSchedule.Times: at
// least one entry, and no more than MaxTimesPerSchedule entries.
func ValidateTimes(times []TimeOfDay) error {
	if len(times) == 0 {
		return ErrEmptyTimes
	}
	if len(times) > MaxTimesPerSchedule {
		return ErrTooManyTimes
	}
	return nil
}

// Validate reports whether s has a valid PillName and Times.
func (s PillSchedule) Validate() error {
	if err := ValidatePillName(s.PillName); err != nil {
		return err
	}
	if err := ValidateTimes(s.Times); err != nil {
		return err
	}
	return nil
}
