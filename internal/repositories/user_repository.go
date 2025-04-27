package repositories

import (
	"context"
	"errors"

	"github.com/jinzhu/gorm"
	"github.com/sirupsen/logrus"
	"github.com/user/user-management-service/internal/models"
	"github.com/user/user-management-service/utils"
)

// UserRepository defines the interface for user repository
type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindByID(ctx context.Context, id uint) (*models.User, error)
	FindByEmail(ctx context.Context, email string) (*models.User, error)
	FindByEmailAndOrganization(ctx context.Context, email string, orgID uint) (*models.User, error)
	Update(ctx context.Context, user *models.User) error
	Delete(ctx context.Context, id uint) error
	List(ctx context.Context, offset, limit int) ([]models.User, int64, error)
	ListByOrganization(ctx context.Context, orgID uint, offset, limit int) ([]models.User, int64, error)
}

// UserRepositoryImpl handles database interactions for users
type UserRepositoryImpl struct {
	DB     *gorm.DB
	Logger *utils.Logger
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB, logger *utils.Logger) *UserRepositoryImpl {
	return &UserRepositoryImpl{
		DB:     db,
		Logger: logger,
	}
}

// Create creates a new user
func (r *UserRepositoryImpl) Create(ctx context.Context, user *models.User) error {
	log := r.Logger.WithContext(ctx)

	if err := r.DB.Create(user).Error; err != nil {
		log.WithError(err).Error("Failed to create user")
		return err
	}

	log.WithField("user_id", user.ID).Info("User created successfully")
	return nil
}

// FindByID finds a user by ID
func (r *UserRepositoryImpl) FindByID(ctx context.Context, id uint) (*models.User, error) {
	log := r.Logger.WithContext(ctx)

	var user models.User
	if err := r.DB.Preload("Organization").First(&user, id).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.WithField("user_id", id).Warn("User not found")
			return nil, errors.New("user not found")
		}
		log.WithError(err).Error("Failed to find user by ID")
		return nil, err
	}

	log.WithField("user_id", id).Debug("User found by ID")
	return &user, nil
}

// FindByEmail finds a user by email
func (r *UserRepositoryImpl) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	log := r.Logger.WithContext(ctx)

	var user models.User
	if err := r.DB.Preload("Organization").Where("email = ?", email).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.WithField("email", email).Warn("User not found by email")
			return nil, errors.New("user not found")
		}
		log.WithError(err).Error("Failed to find user by email")
		return nil, err
	}

	log.WithField("email", email).Debug("User found by email")
	return &user, nil
}

// FindByEmailAndOrganization finds a user by email within a specific organization
func (r *UserRepositoryImpl) FindByEmailAndOrganization(ctx context.Context, email string, orgID uint) (*models.User, error) {
	log := r.Logger.WithContext(ctx)

	var user models.User
	if err := r.DB.Preload("Organization").Where("email = ? AND organization_id = ?", email, orgID).First(&user).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			log.WithFields(logrus.Fields{
				"email":  email,
				"org_id": orgID,
			}).Warn("User not found by email in organization")
			return nil, errors.New("user not found")
		}
		log.WithError(err).Error("Failed to find user by email and organization")
		return nil, err
	}

	log.WithFields(logrus.Fields{
		"email":  email,
		"org_id": orgID,
	}).Debug("User found by email and organization")
	return &user, nil
}

// Update updates a user
func (r *UserRepositoryImpl) Update(ctx context.Context, user *models.User) error {
	log := r.Logger.WithContext(ctx)

	if err := r.DB.Save(user).Error; err != nil {
		log.WithError(err).Error("Failed to update user")
		return err
	}

	log.WithField("user_id", user.ID).Info("User updated successfully")
	return nil
}

// Delete soft deletes a user
func (r *UserRepositoryImpl) Delete(ctx context.Context, id uint) error {
	log := r.Logger.WithContext(ctx)

	if err := r.DB.Where("id = ?", id).Delete(&models.User{}).Error; err != nil {
		log.WithError(err).Error("Failed to delete user")
		return err
	}

	log.WithField("user_id", id).Info("User deleted successfully")
	return nil
}

// List returns a list of users
func (r *UserRepositoryImpl) List(ctx context.Context, offset, limit int) ([]models.User, int64, error) {
	log := r.Logger.WithContext(ctx)

	var users []models.User
	var count int64

	if err := r.DB.Model(&models.User{}).Count(&count).Error; err != nil {
		log.WithError(err).Error("Failed to count users")
		return nil, 0, err
	}

	if err := r.DB.Preload("Organization").Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		log.WithError(err).Error("Failed to list users")
		return nil, 0, err
	}

	log.WithFields(logrus.Fields{
		"count":  count,
		"offset": offset,
		"limit":  limit,
	}).Debug("Users listed successfully")

	return users, count, nil
}

// ListByOrganization returns a list of users for a specific organization
func (r *UserRepositoryImpl) ListByOrganization(ctx context.Context, orgID uint, offset, limit int) ([]models.User, int64, error) {
	log := r.Logger.WithContext(ctx)

	var users []models.User
	var count int64

	if err := r.DB.Model(&models.User{}).Where("organization_id = ?", orgID).Count(&count).Error; err != nil {
		log.WithError(err).Error("Failed to count users in organization")
		return nil, 0, err
	}

	if err := r.DB.Preload("Organization").Where("organization_id = ?", orgID).Offset(offset).Limit(limit).Find(&users).Error; err != nil {
		log.WithError(err).Error("Failed to list users in organization")
		return nil, 0, err
	}

	log.WithFields(logrus.Fields{
		"org_id": orgID,
		"count":  count,
		"offset": offset,
		"limit":  limit,
	}).Debug("Users in organization listed successfully")

	return users, count, nil
}
