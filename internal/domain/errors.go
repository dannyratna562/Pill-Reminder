package domain

import "errors"

var (
	ErrPairingCodeNotFound = errors.New("pairing code not found")
	ErrPairingCodeExpired  = errors.New("pairing code expired")
	ErrPairingCodeUsed     = errors.New("pairing code already used")
	ErrInvalidCodeFormat   = errors.New("invalid pairing code format")
)
