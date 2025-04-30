package models

import (
	"time"

	"github.com/jinzhu/gorm"
)

// UserRole defines the role of a user in an organization
type UserRole string

const (
	// RoleAdmin is an admin in an organization
	RoleAdmin UserRole = "ADMIN"
	// RoleMember is a regular member in an organization
	RoleMember UserRole = "MEMBER"
)

// UserOrganization represents the relationship between a user and an organization
type UserOrganization struct {
	ID             uint         `gorm:"primary_key" json:"id"`
	UserID         uint         `gorm:"not null;unique_index:idx_user_active" json:"user_id"` // Unique index on user_id ensures user can only be in one organization
	User           User         `gorm:"foreignkey:UserID" json:"-"`
	OrganizationID uint         `gorm:"not null" json:"organization_id"`
	Organization   Organization `gorm:"foreignkey:OrganizationID" json:"-"`
	Role           UserRole     `gorm:"size:20;not null;default:'MEMBER'" json:"role"`
	Active         bool         `gorm:"default:true;unique_index:idx_user_active" json:"active"` // Part of the unique index with user_id
	CreatedAt      time.Time    `json:"created_at"`
	UpdatedAt      time.Time    `json:"updated_at"`
	DeletedAt      *time.Time   `sql:"index" json:"-"`
}

// TableName specifies the table name
func (UserOrganization) TableName() string {
	return "user_organizations"
}

// SetupUserOrganizationTable sets up the user_organizations table
func SetupUserOrganizationTable(db *gorm.DB) error {
	if err := db.AutoMigrate(&UserOrganization{}).Error; err != nil {
		return err
	}

	// Add a composite unique index on user_id and organization_id to prevent duplicate users in the same org
	if err := db.Model(&UserOrganization{}).AddUniqueIndex("idx_user_org_unique", "user_id", "organization_id").Error; err != nil {
		return err
	}

	return nil
}
