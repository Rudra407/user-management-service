package services

import (
	"context"
	"errors"

	"github.com/user/user-management-service/config"
	"github.com/user/user-management-service/internal/models"
	"github.com/user/user-management-service/internal/repositories"
	"github.com/user/user-management-service/utils"
)

// OrganizationService defines the interface for organization service
type OrganizationService interface {
	// CreateOrganization creates a new organization
	CreateOrganization(ctx context.Context, name, description string) (*models.Organization, error)
}

// OrganizationServiceImpl implements OrganizationService
type OrganizationServiceImpl struct {
	OrgRepo repositories.OrganizationRepository
	Config  *config.Config
	Logger  *utils.Logger
}

// NewOrganizationService creates a new organization service
func NewOrganizationService(
	orgRepo repositories.OrganizationRepository,
	config *config.Config,
	logger *utils.Logger,
) OrganizationService {
	return &OrganizationServiceImpl{
		OrgRepo: orgRepo,
		Config:  config,
		Logger:  logger,
	}
}

// CreateOrganization creates a new organization
func (s *OrganizationServiceImpl) CreateOrganization(
	ctx context.Context,
	name, description string,
) (*models.Organization, error) {
	log := s.Logger.WithContext(ctx)

	if name == "" {
		return nil, errors.New("organization name is required")
	}

	org := &models.Organization{
		Name:        name,
		Description: description,
		Active:      true,
	}

	if err := s.OrgRepo.Create(ctx, org); err != nil {
		log.WithError(err).Error("Failed to create organization")
		return nil, err
	}

	log.WithField("org_id", org.ID).Info("Organization created successfully")
	return org, nil
}
