package models

import (
	"errors"
	"time"

	"github.com/jinzhu/gorm"
	"golang.org/x/crypto/bcrypt"
)

// User represents a user in the system
type User struct {
	ID             uint         `gorm:"primary_key" json:"id"`
	Name           string       `gorm:"size:100;not null" json:"name"`
	Email          string       `gorm:"size:100;not null;unique" json:"email"`
	Password       string       `gorm:"size:100;not null" json:"-"`
	OrganizationID *uint        `gorm:"index" json:"organization_id"`
	Organization   Organization `gorm:"foreignkey:OrganizationID" json:"-"`
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	DeletedAt      *time.Time   `sql:"index" json:"-"`
}

// BeforeSave hashes the password before saving
func (u *User) BeforeSave() error {
	if len(u.Password) == 0 {
		return errors.New("password cannot be empty")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(u.Password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	u.Password = string(hashedPassword)
	return nil
}

// ValidatePassword validates the user's password
func (u *User) ValidatePassword(password string) error {
	return bcrypt.CompareHashAndPassword([]byte(u.Password), []byte(password))
}

// TableName specifies the table name
func (User) TableName() string {
	return "users"
}

// SetupUserTable sets up the user table
func SetupUserTable(db *gorm.DB) error {
	return db.AutoMigrate(&User{}).Error
}
