package repositories

import (
	"context"
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/user/user-management-service/internal/models"
	"github.com/user/user-management-service/utils"
)

// OrganizationRepository defines the interface for organization repository
type OrganizationRepository interface {
	Create(ctx context.Context, org *models.Organization) error
	FindByID(ctx context.Context, id uint) (*models.Organization, error)
	FindByName(ctx context.Context, name string) (*models.Organization, error)
	Update(ctx context.Context, org *models.Organization) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int) ([]models.Organization, int64, error)
}

// OrganizationRepositoryImpl handles database interactions for organizations
type OrganizationRepositoryImpl struct {
	DB     *gorm.DB
	Logger *utils.Logger
}

// NewOrganizationRepository creates a new organization repository
func NewOrganizationRepository(db *gorm.DB, logger *utils.Logger) *OrganizationRepositoryImpl {
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

// FindByName finds an organization by name
func (r *OrganizationRepositoryImpl) FindByName(ctx context.Context, name string) (*models.Organization, error) {
	log := r.Logger.WithContext(ctx)

	var org models.Organization
	if err := r.DB.Where("name = ?", name).First(&org).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.WithField("name", name).Warn("Organization not found by name")
			return nil, errors.New("organization not found")
		}
		log.WithError(err).Error("Failed to find organization by name")
		return nil, err
	}

	log.WithField("name", name).Debug("Organization found by name")
	return &org, nil
}

// Update updates an organization
func (r *OrganizationRepositoryImpl) Update(ctx context.Context, org *models.Organization) error {
	log := r.Logger.WithContext(ctx)

	if err := r.DB.Save(org).Error; err != nil {
		log.WithError(err).Error("Failed to update organization")
		return err
	}

	log.WithField("org_id", org.ID).Info("Organization updated successfully")
	return nil
}

// Delete soft deletes an organization
func (r *OrganizationRepositoryImpl) Delete(ctx context.Context, id uint) error {
	log := r.Logger.WithContext(ctx)

	// Also need to handle users belonging to the organization
	tx := r.DB.Begin()

	// First soft delete all users belonging to the organization
	if err := tx.Where("organization_id = ?", id).Delete(&models.User{}).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to delete organization's users")
		return err
	}

	// Then soft delete the organization
	if err := tx.Where("id = ?", id).Delete(&models.Organization{}).Error; err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to delete organization")
		return err
	}

	if err := tx.Commit().Error; err != nil {
		log.WithError(err).Error("Failed to commit transaction")
		return err
	}

	log.WithField("org_id", id).Info("Organization and its users deleted successfully")
	return nil
}

// List returns a list of organizations
func (r *OrganizationRepositoryImpl) List(ctx context.Context, offset, limit int) ([]models.Organization, int64, error) {
	log := r.Logger.WithContext(ctx)

	var orgs []models.Organization
	var count int64

	if err := r.DB.Model(&models.Organization{}).Count(&count).Error; err != nil {
		log.WithError(err).Error("Failed to count organizations")
		return nil, 0, err
	}

	if err := r.DB.Offset(offset).Limit(limit).Find(&orgs).Error; err != nil {
		log.WithError(err).Error("Failed to list organizations")
		return nil, 0, err
	}

	log.WithFields(logrus.Fields{
		"count":  count,
		"offset": offset,
		"limit":  limit,
	}).Debug("Organizations listed successfully")

	return orgs, count, nil
}
