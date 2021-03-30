package database

import (
	"errors"
	"regexp"

	"gorm.io/gorm"
)

var repoNamePattern = regexp.MustCompile(`^[a-zA-Z0-9]+(?:[-_]?[a-zA-Z0-9]+)+$`)

// Repo contains repository info.
type Repo struct {
	// UserID is the owner's ID.
	UserID uint `gorm:"index:name_user_id,unique"`
	// Name is the repository name.
	Name string `gorm:"index:name_user_id,unique"`
	// Description is the repository description.
	Description string
	// User is the owner of the repository.
	User User
	// CID is the content identifier of the repository.
	CID string

	gorm.Model
}

// BeforeSave validates fields before saving.
func (r *Repo) BeforeSave(tx *gorm.DB) error {
	if len(r.Name) < 2 || len(r.Name) > 32 {
		return errors.New("name must be between 2 and 32 characters")
	}

	if !repoNamePattern.MatchString(r.Name) {
		return errors.New("name can only contain alphanumeric characters separated by _ or -")
	}

	if len(r.Description) > 80 {
		return errors.New("description must be less than 80 characters")
	}

	return nil
}

func (r *Repo) Create(db *gorm.DB) error {
	return db.Create(r).Error
}

func (r *Repo) UpdateCID(db *gorm.DB) error {
	return db.Model(r).Update("CID", r.CID).Error
}

func (r *Repo) Find(db *gorm.DB, id interface{}) error {
	return db.First(r, id).Error
}

func (r *Repo) FindByNameAndUserID(db *gorm.DB, name string, userID uint) error {
	return db.First(r, "name = ? AND user_id = ?", name, userID).Error
}
