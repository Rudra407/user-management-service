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

// UserService handles business logic for users
type UserService struct {
	UserRepo repositories.UserRepository
	Config   *config.Config
	Logger   *utils.Logger
	OrgRepo  repositories.OrganizationRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo repositories.UserRepository, config *config.Config, logger *utils.Logger, orgRepo repositories.OrganizationRepository) *UserService {
	return &UserService{
		UserRepo: userRepo,
		Config:   config,
		Logger:   logger,
		OrgRepo:  orgRepo,
	}
}

// RegisterUser registers a new user and optionally assigns them to an organization
func (s *UserService) RegisterUser(ctx context.Context, name, email, password string, organizationID *uint) (*models.User, error) {
	log := s.Logger.WithContext(ctx)

	// Validate input
	if err := s.validateRegistration(name, email, password); err != nil {
		log.WithError(err).Warn("Registration validation failed")
		return nil, err
	}

	// Check if email already exists
	existingUser, err := s.UserRepo.FindByEmail(ctx, email)
	if err == nil && existingUser != nil {
		log.WithField("email", email).Warn("Email already registered")
		return nil, errors.New("email already registered")
	}

	// An organization ID is now required
	if organizationID == nil {
		return nil, errors.New("organization ID is required")
	}

	// Validate the organization exists
	org, err := s.OrgRepo.FindByID(ctx, *organizationID)
	if err != nil {
		log.WithError(err).WithField("org_id", *organizationID).Warn("Organization not found")
		return nil, errors.New("organization not found")
	}

	if !org.Active {
		log.WithField("org_id", *organizationID).Warn("Organization is inactive")
		return nil, errors.New("organization is inactive")
	}

	// Create user with direct organization link
	user := &models.User{
		Name:           name,
		Email:          email,
		Password:       password,
		OrganizationID: organizationID,
	}

	if err := s.UserRepo.Create(ctx, user); err != nil {
		log.WithError(err).Error("Failed to create user")
		return nil, err
	}

	// Also maintain the user_organization relationship for roles and additional data
	userOrg := &models.UserOrganization{
		UserID:         user.ID,
		OrganizationID: *organizationID,
		Role:           models.RoleMember,
		Active:         true,
	}

	if err := s.OrgRepo.AddUserToOrg(ctx, userOrg); err != nil {
		log.WithError(err).Error("Failed to add user to organization")
		// If adding to organization fails, we don't rollback user creation but return the error
		return user, errors.New("user created but failed to add to organization: " + err.Error())
	}

	log.WithFields(map[string]interface{}{
		"user_id": user.ID,
		"org_id":  *organizationID,
	}).Info("User added to organization")

	log.WithField("user_id", user.ID).Info("User registered successfully")
	return user, nil
}

// Login authenticates a user and returns a JWT token
func (s *UserService) Login(ctx context.Context, email, password string) (string, error) {
	log := s.Logger.WithContext(ctx)

	user, err := s.UserRepo.FindByEmail(ctx, email)
	if err != nil {
		log.WithField("email", email).Warn("User not found during login")
		return "", errors.New("invalid email or password")
	}

	if err := user.ValidatePassword(password); err != nil {
		log.WithField("user_id", user.ID).Warn("Invalid password during login")
		return "", errors.New("invalid email or password")
	}

	// Generate JWT token
	token, err := utils.GenerateToken(user.ID, s.Config.JWT.Secret, s.Config.JWT.Expiry)
	if err != nil {
		log.WithError(err).Error("Failed to generate JWT token")
		return "", errors.New("authentication failed")
	}

	log.WithField("user_id", user.ID).Info("User logged in successfully")
	return token, nil
}

// GetUserByID gets a user by ID
func (s *UserService) GetUserByID(ctx context.Context, id uint) (*models.User, error) {
	log := s.Logger.WithContext(ctx)

	user, err := s.UserRepo.FindByID(ctx, id)
	if err != nil {
		log.WithError(err).WithField("user_id", id).Warn("Failed to get user by ID")
		return nil, err
	}

	log.WithField("user_id", id).Debug("User retrieved successfully")
	return user, nil
}

// UpdateUser updates a user
func (s *UserService) UpdateUser(ctx context.Context, id uint, name, email, password string) (*models.User, error) {
	log := s.Logger.WithContext(ctx)

	user, err := s.UserRepo.FindByID(ctx, id)
	if err != nil {
		log.WithError(err).WithField("user_id", id).Warn("Failed to find user for update")
		return nil, err
	}

	// Update fields if provided
	if name != "" {
		user.Name = name
	}

	if email != "" && email != user.Email {
		// Check if new email already exists
		existingUser, err := s.UserRepo.FindByEmail(ctx, email)
		if err == nil && existingUser != nil && existingUser.ID != id {
			log.WithField("email", email).Warn("Email already in use")
			return nil, errors.New("email already in use")
		}

		user.Email = email
	}

	if password != "" {
		user.Password = password
	}

	if err := s.UserRepo.Update(ctx, user); err != nil {
		log.WithError(err).WithField("user_id", id).Error("Failed to update user")
		return nil, err
	}

	log.WithField("user_id", id).Info("User updated successfully")
	return user, nil
}

// DeleteUser deletes a user
func (s *UserService) DeleteUser(ctx context.Context, id uint) error {
	log := s.Logger.WithContext(ctx)

	// Check if user exists
	_, err := s.UserRepo.FindByID(ctx, id)
	if err != nil {
		log.WithError(err).WithField("user_id", id).Warn("Failed to find user for deletion")
		return err
	}

	if err := s.UserRepo.Delete(ctx, id); err != nil {
		log.WithError(err).WithField("user_id", id).Error("Failed to delete user")
		return err
	}

	log.WithField("user_id", id).Info("User deleted successfully")
	return nil
}

// ListUsers lists users with pagination
func (s *UserService) ListUsers(ctx context.Context, page, perPage int) ([]models.User, int64, error) {
	log := s.Logger.WithContext(ctx)

	if page < 1 {
		page = 1
	}

	if perPage < 1 {
		perPage = 10
	}

	offset := (page - 1) * perPage

	users, total, err := s.UserRepo.List(ctx, offset, perPage)
	if err != nil {
		log.WithError(err).Error("Failed to list users")
		return nil, 0, err
	}

	log.WithField("total", total).Debug("Users listed successfully")
	return users, total, nil
}

// GetUserOrganization gets the organization for a user
func (s *UserService) GetUserOrganization(ctx context.Context, userID uint) (*models.Organization, error) {
	log := s.Logger.WithContext(ctx)

	// Find user to get organization ID
	user, err := s.UserRepo.FindByID(ctx, userID)
	if err != nil {
		log.WithError(err).WithField("user_id", userID).Warn("User not found")
		return nil, errors.New("user not found")
	}

	// Check if user has an organization set
	if user.OrganizationID == nil {
		log.WithField("user_id", userID).Info("User is not part of any organization")
		return nil, nil
	}

	// Get the organization details
	org, err := s.OrgRepo.FindByID(ctx, *user.OrganizationID)
	if err != nil {
		log.WithError(err).WithField("org_id", *user.OrganizationID).Error("Failed to find organization")
		return nil, err
	}

	log.WithFields(map[string]interface{}{
		"user_id": userID,
		"org_id":  org.ID,
	}).Info("User organization retrieved successfully")

	return org, nil
}

// validateRegistration validates registration input
func (s *UserService) validateRegistration(name, email, password string) error {
	if name == "" {
		return errors.New("name is required")
	}

	if email == "" {
		return errors.New("email is required")
	}

	if !strings.Contains(email, "@") {
		return errors.New("invalid email format")
	}

	if password == "" {
		return errors.New("password is required")
	}

	if len(password) < 6 {
		return errors.New("password must be at least 6 characters")
	}

	return nil
}
