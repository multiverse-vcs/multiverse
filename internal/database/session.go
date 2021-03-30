package database

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

// Session contains session details.
type Session struct {
	// ID is the unique session ID.
	ID string `gorm:"primaryKey"`
	// UserID is the owner's ID.
	UserID uint `gorm:"index"`
	// User is the session user.
	User User
	// CreatedAt is the time the session was created.
	CreatedAt time.Time
	// UpdatedAt is the time the session was updated.
	UpdatedAt time.Time
}

// BeforeCreate validates fields before creating.
func (s *Session) BeforeCreate(tx *gorm.DB) error {
	s.ID = uuid.New().String()
	return nil
}

func (s *Session) Create(db *gorm.DB) error {
	return db.Create(s).Error
}

func (s *Session) Find(db *gorm.DB, id string) error {
	return db.Preload("User").First(s, "id = ?", id).Error
}
