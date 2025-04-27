package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// Organization represents an organization in the system
type Organization struct {
	ID          uint       `gorm:"primary_key" json:"id"`
	Name        string     `gorm:"size:100;not null" json:"name"`
	DisplayName string     `gorm:"size:100;not null" json:"display_name"`
	Description string     `gorm:"size:500" json:"description"`
	Website     string     `gorm:"size:200" json:"website"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
	DeletedAt   *time.Time `sql:"index" json:"-"`
}

// TableName specifies the table name
func (Organization) TableName() string {
	return "organizations"
}

// SetupOrganizationTable sets up the organization table
func SetupOrganizationTable(db *gorm.DB) error {
	return db.AutoMigrate(&Organization{}).Error
}
