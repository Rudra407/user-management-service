package services_test

import (
	"context"
	"errors"
	"testing"

	"github.com/user/user-management-service/config"
	"github.com/user/user-management-service/internal/models"
	"github.com/user/user-management-service/internal/services"
	"github.com/user/user-management-service/utils"
)

// MockUserRepo is a mock implementation of the UserRepository interface
type MockUserRepo struct {
	users         map[uint]*models.User
	emailToUserID map[string]uint
	nextID        uint
}

func NewMockUserRepo() *MockUserRepo {
	return &MockUserRepo{
		users:         make(map[uint]*models.User),
		emailToUserID: make(map[string]uint),
		nextID:        1,
	}
}

func (m *MockUserRepo) Create(ctx context.Context, user *models.User) error {
	// Check if email already exists
	if _, exists := m.emailToUserID[user.Email]; exists {
		return errors.New("email already exists")
	}

	user.ID = m.nextID
	m.nextID++
	m.users[user.ID] = user
	m.emailToUserID[user.Email] = user.ID
	return nil
}

func (m *MockUserRepo) FindByID(ctx context.Context, id uint) (*models.User, error) {
	user, exists := m.users[id]
	if !exists {
		return nil, errors.New("user not found")
	}
	return user, nil
}

func (m *MockUserRepo) FindByEmail(ctx context.Context, email string) (*models.User, error) {
	id, exists := m.emailToUserID[email]
	if !exists {
		return nil, errors.New("user not found")
	}
	return m.users[id], nil
}

func (m *MockUserRepo) Update(ctx context.Context, user *models.User) error {
	if _, exists := m.users[user.ID]; !exists {
		return errors.New("user not found")
	}

	// If email changed, update the email map
	oldEmail := m.users[user.ID].Email
	if oldEmail != user.Email {
		delete(m.emailToUserID, oldEmail)
		m.emailToUserID[user.Email] = user.ID
	}

	m.users[user.ID] = user
	return nil
}

func (m *MockUserRepo) Delete(ctx context.Context, id uint) error {
	user, exists := m.users[id]
	if !exists {
		return errors.New("user not found")
	}

	delete(m.emailToUserID, user.Email)
	delete(m.users, id)
	return nil
}

func (m *MockUserRepo) List(ctx context.Context, offset, limit int) ([]models.User, int64, error) {
	// Convert map to slice
	allUsers := make([]models.User, 0, len(m.users))
	for _, user := range m.users {
		allUsers = append(allUsers, *user)
	}

	// Apply offset and limit
	total := int64(len(allUsers))
	start := offset
	if start >= len(allUsers) {
		return []models.User{}, total, nil
	}

	end := offset + limit
	if end > len(allUsers) {
		end = len(allUsers)
	}

	return allUsers[start:end], total, nil
}

func TestUserService_Register(t *testing.T) {
	// Setup
	mockRepo := NewMockUserRepo()
	cfg := &config.Config{}
	cfg.JWT.Secret = "test-secret"
	cfg.JWT.Expiry = 24
	logger := utils.NewLogger("info")

	userService := services.NewUserService(mockRepo, cfg, logger)
	ctx := context.Background()

	// Test register
	user, err := userService.RegisterUser(ctx, "Test User", "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user.Name != "Test User" {
		t.Errorf("Expected name 'Test User', got %s", user.Name)
	}

	// Test duplicate email
	_, err = userService.RegisterUser(ctx, "Another User", "test@example.com", "password456")
	if err == nil {
		t.Error("Expected error for duplicate email, got nil")
	}
}

func TestUserService_Login(t *testing.T) {
	// Setup
	mockRepo := NewMockUserRepo()
	cfg := &config.Config{}
	cfg.JWT.Secret = "test-secret"
	cfg.JWT.Expiry = 24
	logger := utils.NewLogger("info")

	userService := services.NewUserService(mockRepo, cfg, logger)
	ctx := context.Background()

	// Register a user for testing login
	user := &models.User{
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}

	// Hash the password
	if err := user.BeforeSave(); err != nil {
		t.Fatalf("Failed to hash password: %v", err)
	}

	// Add user directly to mock repo
	user.ID = 1
	mockRepo.users[user.ID] = user
	mockRepo.emailToUserID[user.Email] = user.ID

	// Test valid login
	token, err := userService.Login(ctx, "test@example.com", "password123")
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if token == "" {
		t.Error("Expected token, got empty string")
	}

	// Test invalid login
	_, err = userService.Login(ctx, "test@example.com", "wrongpassword")
	if err == nil {
		t.Error("Expected error for wrong password, got nil")
	}
}

func TestUserService_GetUserByID(t *testing.T) {
	// Setup
	mockRepo := NewMockUserRepo()
	cfg := &config.Config{}
	logger := utils.NewLogger("info")

	userService := services.NewUserService(mockRepo, cfg, logger)
	ctx := context.Background()

	// Add a test user
	testUser := &models.User{
		ID:       1,
		Name:     "Test User",
		Email:    "test@example.com",
		Password: "password123",
	}
	mockRepo.users[testUser.ID] = testUser

	// Test get user by ID
	user, err := userService.GetUserByID(ctx, 1)
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if user.ID != 1 || user.Name != "Test User" {
		t.Errorf("Expected user with ID 1 and name 'Test User', got ID %d and name %s", user.ID, user.Name)
	}

	// Test get non-existent user
	_, err = userService.GetUserByID(ctx, 999)
	if err == nil {
		t.Error("Expected error for non-existent user, got nil")
	}
}
