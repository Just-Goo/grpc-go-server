package database

import (
	"time"

	"github.com/google/uuid"
)

type DummyOrm struct {
	UserID    uuid.UUID `gorm:"primaryKey"`
	Username  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// gorm uses plural form for table names by default (for this case - dummies)
func (DummyOrm) TableName() string { // we can override it by implementing this method
	return "dummy"
}
