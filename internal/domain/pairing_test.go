package domain

import (
	"testing"
	"time"
)

func TestGenerateCode_Format(t *testing.T) {
	for i := 0; i < 200; i++ {
		code, err := GenerateCode()
		if err != nil {
			t.Fatalf("GenerateCode() error = %v", err)
		}
		if err := ValidateCodeFormat(code); err != nil {
			t.Fatalf("GenerateCode() = %q, fails ValidateCodeFormat: %v", code, err)
		}
	}
}

func TestPairingCode_IsExpired(t *testing.T) {
	expiresAt := time.Date(2026, 1, 1, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		now  time.Time
		want bool
	}{
		{"before expiry", expiresAt.Add(-time.Minute), false},
		{"at expiry", expiresAt, false},
		{"after expiry", expiresAt.Add(time.Minute), true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := PairingCode{ExpiresAt: expiresAt}
			if got := p.IsExpired(tt.now); got != tt.want {
				t.Errorf("IsExpired(%v) = %v, want %v", tt.now, got, tt.want)
			}
		})
	}
}

func TestPairingCode_IsUsed(t *testing.T) {
	if (PairingCode{}).IsUsed() {
		t.Error("zero-value PairingCode.IsUsed() = true, want false")
	}
	used := time.Now()
	if !(PairingCode{UsedAt: &used}).IsUsed() {
		t.Error("PairingCode with UsedAt set: IsUsed() = false, want true")
	}
}

func TestValidateCodeFormat(t *testing.T) {
	tests := []struct {
		name    string
		code    string
		wantErr bool
	}{
		{"valid", "042817", false},
		{"too short", "4281", true},
		{"too long", "04281700", true},
		{"non-numeric", "abcdef", true},
		{"empty", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCodeFormat(tt.code)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCodeFormat(%q) error = %v, wantErr %v", tt.code, err, tt.wantErr)
			}
		})
	}
}
