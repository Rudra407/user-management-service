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
	UserRepo         repositories.UserRepository
	OrganizationRepo repositories.OrganizationRepository
	Config           *config.Config
	Logger           *utils.Logger
}

// NewUserService creates a new user service
func NewUserService(userRepo repositories.UserRepository, orgRepo repositories.OrganizationRepository, config *config.Config, logger *utils.Logger) *UserService {
	return &UserService{
		UserRepo:         userRepo,
		OrganizationRepo: orgRepo,
		Config:           config,
		Logger:           logger,
	}
}

// RegisterUser registers a new user within an organization
func (s *UserService) RegisterUser(ctx context.Context, name, email, password string, organizationID uint, role string) (*models.User, error) {
	log := s.Logger.WithContext(ctx)

	// Validate input
	if err := s.validateRegistration(name, email, password); err != nil {
		log.WithError(err).Warn("Registration validation failed")
		return nil, err
	}

	// Verify the organization exists
	org, err := s.OrganizationRepo.FindByID(ctx, organizationID)
	if err != nil {
		log.WithField("org_id", organizationID).Warn("Organization not found during registration")
		return nil, errors.New("organization not found")
	}

	// Check if email already exists in the organization
	existingUser, err := s.UserRepo.FindByEmailAndOrganization(ctx, email, organizationID)
	if err == nil && existingUser != nil {
		log.WithFields(map[string]interface{}{
			"email":  email,
			"org_id": organizationID,
		}).Warn("Email already registered in this organization")
		return nil, errors.New("email already registered in this organization")
	}

	// Use default role if not provided
	if role == "" {
		role = "user"
	}

	// Create user
	user := &models.User{
		Name:           name,
		Email:          email,
		Password:       password,
		OrganizationID: org.ID,
		Role:           role,
	}

	if err := s.UserRepo.Create(ctx, user); err != nil {
		log.WithError(err).Error("Failed to create user")
		return nil, err
	}

	log.WithFields(map[string]interface{}{
		"user_id": user.ID,
		"org_id":  org.ID,
	}).Info("User registered successfully")
	return user, nil
}

// LoginWithOrganization authenticates a user for a specific organization and returns a JWT token
func (s *UserService) LoginWithOrganization(ctx context.Context, email, password string, organizationID uint) (string, *models.User, error) {
	log := s.Logger.WithContext(ctx)

	user, err := s.UserRepo.FindByEmailAndOrganization(ctx, email, organizationID)
	if err != nil {
		log.WithFields(map[string]interface{}{
			"email":  email,
			"org_id": organizationID,
		}).Warn("User not found during login")
		return "", nil, errors.New("invalid email or password")
	}

	if err := user.ValidatePassword(password); err != nil {
		log.WithFields(map[string]interface{}{
			"user_id": user.ID,
			"org_id":  organizationID,
		}).Warn("Invalid password during login")
		return "", nil, errors.New("invalid email or password")
	}

	// Generate JWT token with user ID and organization ID
	token, err := utils.GenerateTokenWithOrganization(user.ID, user.OrganizationID, user.Role, s.Config.JWT.Secret, s.Config.JWT.Expiry)
	if err != nil {
		log.WithError(err).Error("Failed to generate JWT token")
		return "", nil, errors.New("authentication failed")
	}

	log.WithFields(map[string]interface{}{
		"user_id": user.ID,
		"org_id":  user.OrganizationID,
	}).Info("User logged in successfully")
	return token, user, nil
}

// Login authenticates a user and returns a JWT token
// This method is maintained for backward compatibility
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

	// Generate JWT token with both user ID and organization ID
	token, err := utils.GenerateTokenWithOrganization(user.ID, user.OrganizationID, user.Role, s.Config.JWT.Secret, s.Config.JWT.Expiry)
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
		// Check if new email already exists in the same organization
		existingUser, err := s.UserRepo.FindByEmailAndOrganization(ctx, email, user.OrganizationID)
		if err == nil && existingUser != nil && existingUser.ID != id {
			log.WithFields(map[string]interface{}{
				"email":  email,
				"org_id": user.OrganizationID,
			}).Warn("Email already in use in this organization")
			return nil, errors.New("email already in use in this organization")
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

// UpdateUserRole updates a user's role
func (s *UserService) UpdateUserRole(ctx context.Context, id uint, role string) (*models.User, error) {
	log := s.Logger.WithContext(ctx)

	user, err := s.UserRepo.FindByID(ctx, id)
	if err != nil {
		log.WithError(err).WithField("user_id", id).Warn("Failed to find user for role update")
		return nil, err
	}

	// Validate role
	if role != "admin" && role != "user" {
		log.WithField("role", role).Warn("Invalid role")
		return nil, errors.New("invalid role, must be 'admin' or 'user'")
	}

	user.Role = role

	if err := s.UserRepo.Update(ctx, user); err != nil {
		log.WithError(err).WithField("user_id", id).Error("Failed to update user role")
		return nil, err
	}

	log.WithFields(map[string]interface{}{
		"user_id": id,
		"role":    role,
	}).Info("User role updated successfully")
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

// ListOrganizationUsers lists users for a specific organization with pagination
func (s *UserService) ListOrganizationUsers(ctx context.Context, organizationID uint, page, perPage int) ([]models.User, int64, error) {
	log := s.Logger.WithContext(ctx)

	// Verify the organization exists
	_, err := s.OrganizationRepo.FindByID(ctx, organizationID)
	if err != nil {
		log.WithField("org_id", organizationID).Warn("Organization not found when listing users")
		return nil, 0, errors.New("organization not found")
	}

	if page < 1 {
		page = 1
	}

	if perPage < 1 {
		perPage = 10
	}

	offset := (page - 1) * perPage

	users, total, err := s.UserRepo.ListByOrganization(ctx, organizationID, offset, perPage)
	if err != nil {
		log.WithError(err).Error("Failed to list organization users")
		return nil, 0, err
	}

	log.WithFields(map[string]interface{}{
		"org_id": organizationID,
		"total":  total,
	}).Debug("Organization users listed successfully")
	return users, total, nil
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
