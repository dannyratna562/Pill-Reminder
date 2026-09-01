package domain

import "errors"

var (
	ErrPairingCodeNotFound = errors.New("pairing code not found")
	ErrPairingCodeExpired  = errors.New("pairing code expired")
	ErrPairingCodeUsed     = errors.New("pairing code already used")
	ErrInvalidCodeFormat   = errors.New("invalid pairing code format")

	ErrEmptyPillName     = errors.New("pill name must not be empty")
	ErrPillNameTooLong   = errors.New("pill name exceeds maximum length")
	ErrEmptyTimes        = errors.New("times must not be empty")
	ErrTooManyTimes      = errors.New("too many times for schedule")
	ErrInvalidTimeFormat = errors.New("invalid time-of-day format")
)
