package domain_test

import (
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/pillreminder/backend/internal/domain"
)

func TestParseTimeOfDay(t *testing.T) {
	validTests := []struct {
		name string
		in   string
		hour int
		min  int
	}{
		{"midnight", "00:00", 0, 0},
		{"morning", "08:00", 8, 0},
		{"end of day", "23:59", 23, 59},
	}
	for _, tt := range validTests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := domain.ParseTimeOfDay(tt.in)
			if err != nil {
				t.Fatalf("ParseTimeOfDay(%q) error = %v", tt.in, err)
			}
			if got.Hour() != tt.hour || got.Minute() != tt.min {
				t.Errorf("ParseTimeOfDay(%q) = %02d:%02d, want %02d:%02d", tt.in, got.Hour(), got.Minute(), tt.hour, tt.min)
			}
		})
	}

	invalidTests := []struct {
		name string
		in   string
	}{
		{"single digit hour", "8:00"},
		{"hour 24", "24:00"},
		{"hour 25", "25:00"},
		{"minute 60", "08:60"},
		{"with seconds", "08:00:00"},
		{"no colon", "0800"},
		{"empty", ""},
		{"leading whitespace", " 08:00"},
		{"trailing whitespace", "08:00 "},
		{"non-ascii digits", "０８:００"}, // fullwidth digits
	}
	for _, tt := range invalidTests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := domain.ParseTimeOfDay(tt.in)
			if !errors.Is(err, domain.ErrInvalidTimeFormat) {
				t.Errorf("ParseTimeOfDay(%q) error = %v, want ErrInvalidTimeFormat", tt.in, err)
			}
		})
	}
}

func TestTimeOfDay_StringRoundTrip(t *testing.T) {
	values := []string{"00:00", "08:00", "23:59", "12:30", "01:05"}
	for _, s := range values {
		t.Run(s, func(t *testing.T) {
			parsed, err := domain.ParseTimeOfDay(s)
			if err != nil {
				t.Fatalf("ParseTimeOfDay(%q) error = %v", s, err)
			}
			if got := parsed.String(); got != s {
				t.Errorf("String() = %q, want %q", got, s)
			}
		})
	}
}

func TestNormalizeTimes(t *testing.T) {
	mustParse := func(s string) domain.TimeOfDay {
		tt, err := domain.ParseTimeOfDay(s)
		if err != nil {
			t.Fatalf("ParseTimeOfDay(%q) error = %v", s, err)
		}
		return tt
	}

	t.Run("sorts unsorted input", func(t *testing.T) {
		in := []domain.TimeOfDay{mustParse("20:30"), mustParse("08:00"), mustParse("12:00")}
		got := domain.NormalizeTimes(in)
		want := []string{"08:00", "12:00", "20:30"}
		if len(got) != len(want) {
			t.Fatalf("NormalizeTimes() len = %d, want %d", len(got), len(want))
		}
		for i, w := range want {
			if got[i].String() != w {
				t.Errorf("NormalizeTimes()[%d] = %s, want %s", i, got[i].String(), w)
			}
		}
	})

	t.Run("collapses duplicates", func(t *testing.T) {
		in := []domain.TimeOfDay{mustParse("08:00"), mustParse("08:00"), mustParse("20:30")}
		got := domain.NormalizeTimes(in)
		want := []string{"08:00", "20:30"}
		if len(got) != len(want) {
			t.Fatalf("NormalizeTimes() len = %d, want %d", len(got), len(want))
		}
		for i, w := range want {
			if got[i].String() != w {
				t.Errorf("NormalizeTimes()[%d] = %s, want %s", i, got[i].String(), w)
			}
		}
	})

	t.Run("empty stays empty", func(t *testing.T) {
		got := domain.NormalizeTimes(nil)
		if len(got) != 0 {
			t.Errorf("NormalizeTimes(nil) = %v, want empty", got)
		}
	})

	t.Run("does not mutate input", func(t *testing.T) {
		in := []domain.TimeOfDay{mustParse("20:30"), mustParse("08:00"), mustParse("08:00")}
		inCopy := make([]domain.TimeOfDay, len(in))
		copy(inCopy, in)

		_ = domain.NormalizeTimes(in)

		for i := range in {
			if in[i] != inCopy[i] {
				t.Errorf("NormalizeTimes() mutated input at index %d: got %v, want %v", i, in[i], inCopy[i])
			}
		}
	})
}

func TestValidateTimes(t *testing.T) {
	mustParse := func(s string) domain.TimeOfDay {
		tt, err := domain.ParseTimeOfDay(s)
		if err != nil {
			t.Fatalf("ParseTimeOfDay(%q) error = %v", s, err)
		}
		return tt
	}

	t.Run("empty", func(t *testing.T) {
		err := domain.ValidateTimes(nil)
		if !errors.Is(err, domain.ErrEmptyTimes) {
			t.Errorf("ValidateTimes(nil) error = %v, want ErrEmptyTimes", err)
		}
	})

	t.Run("exactly max", func(t *testing.T) {
		times := make([]domain.TimeOfDay, 0, domain.MaxTimesPerSchedule)
		for i := 0; i < domain.MaxTimesPerSchedule; i++ {
			times = append(times, mustParse(hhmm(i)))
		}
		if err := domain.ValidateTimes(times); err != nil {
			t.Errorf("ValidateTimes(24 times) error = %v, want nil", err)
		}
	})

	t.Run("over max", func(t *testing.T) {
		times := make([]domain.TimeOfDay, 0, domain.MaxTimesPerSchedule+1)
		for i := 0; i < domain.MaxTimesPerSchedule+1; i++ {
			times = append(times, mustParse(hhmm(i)))
		}
		err := domain.ValidateTimes(times)
		if !errors.Is(err, domain.ErrTooManyTimes) {
			t.Errorf("ValidateTimes(25 times) error = %v, want ErrTooManyTimes", err)
		}
	})
}

// hhmm returns a distinct valid "HH:MM" string for hour h in [0,23]; values
// of h beyond 23 clamp to 23:MM using the minute component to stay unique
// and in-range, which is enough for the 24/25-element test cases above.
func hhmm(h int) string {
	if h <= 23 {
		return fmt.Sprintf("%02d:00", h)
	}
	return fmt.Sprintf("23:%02d", h-23)
}

func TestValidatePillName(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		err := domain.ValidatePillName("")
		if !errors.Is(err, domain.ErrEmptyPillName) {
			t.Errorf("ValidatePillName(\"\") error = %v, want ErrEmptyPillName", err)
		}
	})

	t.Run("whitespace only", func(t *testing.T) {
		err := domain.ValidatePillName("   ")
		if !errors.Is(err, domain.ErrEmptyPillName) {
			t.Errorf("ValidatePillName(\"   \") error = %v, want ErrEmptyPillName", err)
		}
	})

	t.Run("exactly max length", func(t *testing.T) {
		name := strings.Repeat("a", domain.MaxPillNameLen)
		if err := domain.ValidatePillName(name); err != nil {
			t.Errorf("ValidatePillName(100 chars) error = %v, want nil", err)
		}
	})

	t.Run("over max length", func(t *testing.T) {
		name := strings.Repeat("a", domain.MaxPillNameLen+1)
		err := domain.ValidatePillName(name)
		if !errors.Is(err, domain.ErrPillNameTooLong) {
			t.Errorf("ValidatePillName(101 chars) error = %v, want ErrPillNameTooLong", err)
		}
	})

	t.Run("100 multi-byte runes is within limit even though it is 200 bytes", func(t *testing.T) {
		// "П" is a 2-byte UTF-8 Cyrillic character, so 100 repetitions is
		// 100 runes but 200 bytes: a byte-counting check would wrongly
		// reject this even though it's well within the 100-character limit
		// (and within Postgres's char_length(...) <= 100 CHECK constraint).
		name := strings.Repeat("П", domain.MaxPillNameLen)
		if got := len(name); got != domain.MaxPillNameLen*2 {
			t.Fatalf("test setup: len(name) = %d bytes, want %d", got, domain.MaxPillNameLen*2)
		}
		if err := domain.ValidatePillName(name); err != nil {
			t.Errorf("ValidatePillName(100 Cyrillic runes) error = %v, want nil", err)
		}
	})

	t.Run("101 multi-byte runes exceeds limit", func(t *testing.T) {
		name := strings.Repeat("П", domain.MaxPillNameLen+1)
		err := domain.ValidatePillName(name)
		if !errors.Is(err, domain.ErrPillNameTooLong) {
			t.Errorf("ValidatePillName(101 Cyrillic runes) error = %v, want ErrPillNameTooLong", err)
		}
	})
}

func TestPillSchedule_Validate(t *testing.T) {
	validTime, err := domain.ParseTimeOfDay("08:00")
	if err != nil {
		t.Fatalf("ParseTimeOfDay() error = %v", err)
	}

	t.Run("happy path", func(t *testing.T) {
		s := domain.PillSchedule{
			PillName: "Aspirin",
			Times:    []domain.TimeOfDay{validTime},
		}
		if err := s.Validate(); err != nil {
			t.Errorf("Validate() error = %v, want nil", err)
		}
	})

	t.Run("empty pill name", func(t *testing.T) {
		s := domain.PillSchedule{
			PillName: "",
			Times:    []domain.TimeOfDay{validTime},
		}
		if err := s.Validate(); !errors.Is(err, domain.ErrEmptyPillName) {
			t.Errorf("Validate() error = %v, want ErrEmptyPillName", err)
		}
	})

	t.Run("pill name too long", func(t *testing.T) {
		s := domain.PillSchedule{
			PillName: strings.Repeat("a", domain.MaxPillNameLen+1),
			Times:    []domain.TimeOfDay{validTime},
		}
		if err := s.Validate(); !errors.Is(err, domain.ErrPillNameTooLong) {
			t.Errorf("Validate() error = %v, want ErrPillNameTooLong", err)
		}
	})

	t.Run("empty times", func(t *testing.T) {
		s := domain.PillSchedule{
			PillName: "Aspirin",
			Times:    nil,
		}
		if err := s.Validate(); !errors.Is(err, domain.ErrEmptyTimes) {
			t.Errorf("Validate() error = %v, want ErrEmptyTimes", err)
		}
	})

	t.Run("too many times", func(t *testing.T) {
		times := make([]domain.TimeOfDay, domain.MaxTimesPerSchedule+1)
		for i := range times {
			times[i] = validTime
		}
		s := domain.PillSchedule{
			PillName: "Aspirin",
			Times:    times,
		}
		if err := s.Validate(); !errors.Is(err, domain.ErrTooManyTimes) {
			t.Errorf("Validate() error = %v, want ErrTooManyTimes", err)
		}
	})
}
