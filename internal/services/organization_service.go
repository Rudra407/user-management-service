package services

import (
	"context"
	"errors"
	"strings"

	"github.com/user/user-management-service/config"
	"github.com/user/user-management-service/internal/models"
	"github.com/user/user-management-service/internal/repositories"
	"github.com/user/user-management-service/utils"
)

// OrganizationService handles business logic for organizations
type OrganizationService struct {
	OrganizationRepo repositories.OrganizationRepository
	UserRepo         repositories.UserRepository
	Config           *config.Config
	Logger           *utils.Logger
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(
	orgRepo repositories.OrganizationRepository,
	userRepo repositories.UserRepository,
	config *config.Config,
	logger *utils.Logger,
) *OrganizationService {
	return &OrganizationService{
		OrganizationRepo: orgRepo,
		UserRepo:         userRepo,
		Config:           config,
		Logger:           logger,
	}
}

// CreateOrganization creates a new organization with an admin user
func (s *OrganizationService) CreateOrganization(
	ctx context.Context,
	name, displayName, description, website string,
	adminName, adminEmail, adminPassword string,
) (*models.Organization, *models.User, error) {
	log := s.Logger.WithContext(ctx)

	// Validate organization input
	if err := s.validateOrganization(name, displayName); err != nil {
		log.WithError(err).Warn("Organization validation failed")
		return nil, nil, err
	}

	// Validate admin user input
	if err := s.validateAdminUser(adminName, adminEmail, adminPassword); err != nil {
		log.WithError(err).Warn("Admin user validation failed")
		return nil, nil, err
	}

	// Check if organization name already exists
	existingOrg, err := s.OrganizationRepo.FindByName(ctx, name)
	if err == nil && existingOrg != nil {
		log.WithField("name", name).Warn("Organization name already exists")
		return nil, nil, errors.New("organization name already exists")
	}

	// Create organization
	org := &models.Organization{
		Name:        name,
		DisplayName: displayName,
		Description: description,
		Website:     website,
	}

	// Start a transaction
	tx := s.OrganizationRepo.(*repositories.OrganizationRepositoryImpl).DB.Begin()
	txCtx := context.WithValue(ctx, "tx", tx)

	if err := s.OrganizationRepo.Create(txCtx, org); err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to create organization")
		return nil, nil, err
	}

	// Create admin user
	admin := &models.User{
		Name:           adminName,
		Email:          adminEmail,
		Password:       adminPassword,
		OrganizationID: org.ID,
		Role:           "admin",
	}

	if err := s.UserRepo.Create(txCtx, admin); err != nil {
		tx.Rollback()
		log.WithError(err).Error("Failed to create admin user")
		return nil, nil, err
	}

	if err := tx.Commit().Error; err != nil {
		log.WithError(err).Error("Failed to commit transaction")
		return nil, nil, err
	}

	log.WithField("org_id", org.ID).Info("Organization and admin user created successfully")
	return org, admin, nil
}

// GetOrganizationByID gets an organization by ID
func (s *OrganizationService) GetOrganizationByID(ctx context.Context, id uint) (*models.Organization, error) {
	log := s.Logger.WithContext(ctx)

	org, err := s.OrganizationRepo.FindByID(ctx, id)
	if err != nil {
		log.WithError(err).WithField("org_id", id).Warn("Failed to get organization by ID")
		return nil, err
	}

	log.WithField("org_id", id).Debug("Organization retrieved successfully")
	return org, nil
}

// UpdateOrganization updates an organization
func (s *OrganizationService) UpdateOrganization(
	ctx context.Context,
	id uint,
	displayName, description, website string,
) (*models.Organization, error) {
	log := s.Logger.WithContext(ctx)

	org, err := s.OrganizationRepo.FindByID(ctx, id)
	if err != nil {
		log.WithError(err).WithField("org_id", id).Warn("Failed to find organization for update")
		return nil, err
	}

	// Update fields if provided
	if displayName != "" {
		org.DisplayName = displayName
	}

	if description != "" {
		org.Description = description
	}

	if website != "" {
		org.Website = website
	}

	if err := s.OrganizationRepo.Update(ctx, org); err != nil {
		log.WithError(err).WithField("org_id", id).Error("Failed to update organization")
		return nil, err
	}

	log.WithField("org_id", id).Info("Organization updated successfully")
	return org, nil
}

// DeleteOrganization deletes an organization and its users
func (s *OrganizationService) DeleteOrganization(ctx context.Context, id uint) error {
	log := s.Logger.WithContext(ctx)

	// Check if organization exists
	_, err := s.OrganizationRepo.FindByID(ctx, id)
	if err != nil {
		log.WithError(err).WithField("org_id", id).Warn("Failed to find organization for deletion")
		return err
	}

	if err := s.OrganizationRepo.Delete(ctx, id); err != nil {
		log.WithError(err).WithField("org_id", id).Error("Failed to delete organization")
		return err
	}

	log.WithField("org_id", id).Info("Organization deleted successfully")
	return nil
}

// ListOrganizations lists organizations with pagination
func (s *OrganizationService) ListOrganizations(ctx context.Context, page, perPage int) ([]models.Organization, int64, error) {
	log := s.Logger.WithContext(ctx)

	if page < 1 {
		page = 1
	}

	if perPage < 1 {
		perPage = 10
	}

	offset := (page - 1) * perPage

	orgs, total, err := s.OrganizationRepo.List(ctx, offset, perPage)
	if err != nil {
		log.WithError(err).Error("Failed to list organizations")
		return nil, 0, err
	}

	log.WithField("total", total).Debug("Organizations listed successfully")
	return orgs, total, nil
}

// ListOrganizationUsers lists users belonging to a specific organization
func (s *OrganizationService) ListOrganizationUsers(ctx context.Context, orgID uint, page, perPage int) ([]models.User, int64, error) {
	log := s.Logger.WithContext(ctx)

	// Check if organization exists
	_, err := s.OrganizationRepo.FindByID(ctx, orgID)
	if err != nil {
		log.WithError(err).WithField("org_id", orgID).Warn("Organization not found")
		return nil, 0, err
	}

	if page < 1 {
		page = 1
	}

	if perPage < 1 {
		perPage = 10
	}

	offset := (page - 1) * perPage

	users, total, err := s.UserRepo.ListByOrganization(ctx, orgID, offset, perPage)
	if err != nil {
		log.WithError(err).Error("Failed to list organization users")
		return nil, 0, err
	}

	log.WithFields(map[string]interface{}{
		"org_id": orgID,
		"total":  total,
	}).Debug("Organization users listed successfully")
	return users, total, nil
}

// validateOrganization validates organization input
func (s *OrganizationService) validateOrganization(name, displayName string) error {
	if name == "" {
		return errors.New("name is required")
	}

	if displayName == "" {
		return errors.New("display name is required")
	}

	// Check for spaces or special characters in name (used for URL/tenant ID)
	if strings.ContainsAny(name, " !@#$%^&*()+={}[]|\\:;\"'<>,.?/") {
		return errors.New("name must not contain spaces or special characters")
	}

	return nil
}

// validateAdminUser validates admin user input
func (s *OrganizationService) validateAdminUser(name, email, password string) error {
	if name == "" {
		return errors.New("admin name is required")
	}

	if email == "" {
		return errors.New("admin email is required")
	}

	if !strings.Contains(email, "@") {
		return errors.New("invalid admin email format")
	}

	if password == "" {
		return errors.New("admin password is required")
	}

	if len(password) < 6 {
		return errors.New("admin password must be at least 6 characters")
	}

	return nil
}
