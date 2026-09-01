package domain

import (
	"fmt"

	"github.com/google/uuid"
)

// ParseID parses raw as a UUID, naming field in the error so handlers can
// surface which request field was invalid.
func ParseID(raw, field string) (uuid.UUID, error) {
	id, err := uuid.Parse(raw)
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("%s: %w", field, err)
	}
	return id, nil
}
