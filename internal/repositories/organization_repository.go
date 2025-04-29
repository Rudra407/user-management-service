package repositories

import (
	"context"
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/user/user-management-service/internal/models"
	"github.com/user/user-management-service/utils"
)

// OrganizationRepository defines the interface for organization repository
type OrganizationRepository interface {
	Create(ctx context.Context, org *models.Organization) error
	FindByID(ctx context.Context, id uint) (*models.Organization, error)
	AddUserToOrg(ctx context.Context, userOrg *models.UserOrganization) error
	IsUserInAnyOrg(ctx context.Context, userID uint) (bool, error)
}

// OrganizationRepositoryImpl handles database interactions for organizations
type OrganizationRepositoryImpl struct {
	DB     *gorm.DB
	Logger *utils.Logger
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(db *gorm.DB, logger *utils.Logger) OrganizationRepository {
	return &OrganizationRepositoryImpl{
		DB:     db,
		Logger: logger,
	}
}

// Create creates a new organization
func (r *OrganizationRepositoryImpl) Create(ctx context.Context, org *models.Organization) error {
	log := r.Logger.WithContext(ctx)

	if err := r.DB.Create(org).Error; err != nil {
		log.WithError(err).Error("Failed to create organization")
		return err
	}

	log.WithField("org_id", org.ID).Info("Organization created successfully")
	return nil
}

// FindByID finds an organization by ID
func (r *OrganizationRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.Organization, error) {
	log := r.Logger.WithContext(ctx)

	var org models.Organization
	if err := r.DB.First(&org, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.WithField("org_id", id).Warn("Organization not found")
			return nil, errors.New("organization not found")
		}
		log.WithError(err).Error("Failed to find organization by ID")
		return nil, err
	}

	log.WithField("org_id", id).Debug("Organization found by ID")
	return &org, nil
}

// AddUserToOrg adds a user to an organization
func (r *OrganizationRepositoryImpl) AddUserToOrg(ctx context.Context, userOrg *models.UserOrganization) error {
	log := r.Logger.WithContext(ctx)

	// Check if this user-org relationship already exists
	var count int64
	if err := r.DB.Model(&models.UserOrganization{}).
		Where("user_id = ? AND organization_id = ? AND active = true", userOrg.UserID, userOrg.OrganizationID).
		Count(&count).Error; err != nil {
		log.WithError(err).Error("Failed to check if user is in organization")
		return err
	}

	if count > 0 {
		log.WithFields(map[string]interface{}{
			"user_id": userOrg.UserID,
			"org_id":  userOrg.OrganizationID,
		}).Warn("User is already a member of this organization")
		return errors.New("user is already a member of this organization")
	}

	// Create the user-organization relationship
	if err := r.DB.Create(userOrg).Error; err != nil {
		log.WithError(err).Error("Failed to add user to organization")
		return err
	}

	log.WithFields(map[string]interface{}{
		"user_id": userOrg.UserID,
		"org_id":  userOrg.OrganizationID,
		"role":    userOrg.Role,
	}).Info("User added to organization successfully")
	return nil
}

// IsUserInAnyOrg checks if a user is in any organization
func (r *OrganizationRepositoryImpl) IsUserInAnyOrg(ctx context.Context, userID uint) (bool, error) {
	log := r.Logger.WithContext(ctx)

	var count int64
	if err := r.DB.Model(&models.UserOrganization{}).
		Where("user_id = ? AND active = true", userID).
		Count(&count).Error; err != nil {
		log.WithError(err).Error("Failed to check if user is in any organization")
		return false, err
	}

	return count > 0, nil
}
