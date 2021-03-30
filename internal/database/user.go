package database

import (
	"errors"
	"regexp"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	userUsernamePattern = regexp.MustCompile(`^[a-zA-Z0-9]+(?:[-_]?[a-zA-Z0-9]+)+$`)
	userEmailPattern    = regexp.MustCompile(`^[^@]+@[^\.]+\..+$`)
)

// User contains account details.
type User struct {
	// Username is an alias for the user.
	Username string `gorm:"uniqueIndex"`
	// Email is the account email address.
	Email string `gorm:"uniqueIndex"`
	// Password is the plain text password.
	Password string `gorm:"-"`
	// PasswordHash contains the hased password.
	PasswordHash []byte

	gorm.Model
}

// BeforeSave validates fields before saving.
func (u *User) BeforeSave(tx *gorm.DB) error {
	if len(u.Username) < 2 || len(u.Username) > 32 {
		return errors.New("username must be between 2 and 32 characters")
	}

	if !userUsernamePattern.MatchString(u.Username) {
		return errors.New("username can only contain alphanumeric characters separated by _ or -")
	}

	if !userEmailPattern.MatchString(u.Email) {
		return errors.New("email address format is invalid")
	}

	return nil
}

// BeforeCreate validates fields before creating.
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if len(u.Password) < 8 || len(u.Password) > 64 {
		return errors.New("password must be between 8 and 64 characters")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.PasswordHash = hash
	return nil
}

func (u *User) Create(db *gorm.DB) error {
	return db.Create(u).Error
}

func (u *User) Find(db *gorm.DB, id interface{}) error {
	return db.First(u, id).Error
}

func (u *User) FindByEmail(db *gorm.DB, email string) error {
	return db.First(u, "email = ?", email).Error
}

func (u *User) FindByUsername(db *gorm.DB, username string) error {
	return db.First(u, "username = ?", username).Error
}

func (u *User) FindByEmailOrUsername(db *gorm.DB, email, username string) error {
	return db.First(u, "email = ? OR username = ?", email, username).Error
}
