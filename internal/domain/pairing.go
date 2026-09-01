package domain

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/google/uuid"
)

// PairingCodeTTL is how long a generated pairing code remains redeemable.
const PairingCodeTTL = 10 * time.Minute

// codeSpace is the exclusive upper bound for a 6-digit code (000000-999999).
var codeSpace = big.NewInt(1000000)

// PairingCode is a short-lived, single-use code a Parent app generates and
// a Child app redeems to create a FamilyLink.
type PairingCode struct {
	ID        uuid.UUID
	Code      string
	ParentID  uuid.UUID
	ExpiresAt time.Time
	UsedAt    *time.Time
	CreatedAt time.Time
}

// IsExpired reports whether the code's TTL has elapsed as of now.
func (p PairingCode) IsExpired(now time.Time) bool {
	return now.After(p.ExpiresAt)
}

// IsUsed reports whether the code has already been redeemed.
func (p PairingCode) IsUsed() bool {
	return p.UsedAt != nil
}

// GenerateCode returns a cryptographically random 6-digit numeric code,
// zero-padded so leading zeros are preserved (e.g. "004217").
func GenerateCode() (string, error) {
	n, err := rand.Int(rand.Reader, codeSpace)
	if err != nil {
		return "", fmt.Errorf("generate pairing code: %w", err)
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

// ValidateCodeFormat returns ErrInvalidCodeFormat unless code is exactly six
// ASCII digits.
func ValidateCodeFormat(code string) error {
	if len(code) != 6 {
		return ErrInvalidCodeFormat
	}
	for _, r := range code {
		if r < '0' || r > '9' {
			return ErrInvalidCodeFormat
		}
	}
	return nil
}
