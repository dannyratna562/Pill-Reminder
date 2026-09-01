package domain

import (
	"time"

	"github.com/google/uuid"
)

// FamilyLink records that a child app instance is linked to a parent app
// instance; pill schedules are scoped to ParentID and only the linked
// ChildID may write to them.
type FamilyLink struct {
	ID        uuid.UUID
	ChildID   uuid.UUID
	ParentID  uuid.UUID
	CreatedAt time.Time
}
