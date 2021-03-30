package database

import (
	"gorm.io/gorm"
)

// Open returns a new database instance.
func Open(dialector gorm.Dialector) (*gorm.DB, error) {
	db, err := gorm.Open(dialector, &gorm.Config{})
	if err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&User{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Session{}); err != nil {
		return nil, err
	}

	if err := db.AutoMigrate(&Repo{}); err != nil {
		return nil, err
	}

	return db, nil
}
