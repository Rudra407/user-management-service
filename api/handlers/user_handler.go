package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/user/user-management-service/api/middleware"
	"github.com/user/user-management-service/internal/services"
	"github.com/user/user-management-service/utils"
)

// UserHandler handles HTTP requests for user operations
type UserHandler struct {
	UserService *services.UserService
	Logger      *utils.Logger
}

// NewUserHandler creates a new user handler
func NewUserHandler(userService *services.UserService, logger *utils.Logger) *UserHandler {
	return &UserHandler{
		UserService: userService,
		Logger:      logger,
	}
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Name           string `json:"name" validate:"required"`
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required,min=6"`
	OrganizationID uint   `json:"organization_id" validate:"required"`
	Role           string `json:"role"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email          string `json:"email" validate:"required,email"`
	Password       string `json:"password" validate:"required"`
	OrganizationID uint   `json:"organization_id"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email" validate:"omitempty,email"`
	Password string `json:"password" validate:"omitempty,min=6"`
}

// UpdateUserRoleRequest represents a request to update a user's role
type UpdateUserRoleRequest struct {
	Role string `json:"role" validate:"required"`
}

// Register handles user registration
func (h *UserHandler) Register(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	var req RegisterRequest
	if err := c.Bind(&req); err != nil {
		log.WithError(err).Warn("Invalid request payload")
		return utils.ValidationErrorResponse(c, "Invalid request payload", []string{err.Error()})
	}

	user, err := h.UserService.RegisterUser(ctx, req.Name, req.Email, req.Password, req.OrganizationID, req.Role)
	if err != nil {
		log.WithError(err).Error("Failed to register user")
		return utils.ErrorResponse(c, http.StatusBadRequest, "Failed to register user", []string{err.Error()})
	}

	return utils.SuccessResponse(c, user, "User registered successfully")
}

// Login handles user login
func (h *UserHandler) Login(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	var req LoginRequest
	if err := c.Bind(&req); err != nil {
		log.WithError(err).Warn("Invalid request payload")
		return utils.ValidationErrorResponse(c, "Invalid request payload", []string{err.Error()})
	}

	var token string
	var err error

	// If organization ID is provided, use organization-specific login
	if req.OrganizationID > 0 {
		token, _, err = h.UserService.LoginWithOrganization(ctx, req.Email, req.Password, req.OrganizationID)
	} else {
		// Otherwise use regular login (backward compatibility)
		token, err = h.UserService.Login(ctx, req.Email, req.Password)
	}

	if err != nil {
		log.WithError(err).Warn("Login failed")
		return utils.UnauthorizedErrorResponse(c, "Invalid credentials")
	}

	return utils.SuccessResponse(c, map[string]string{"token": token}, "Login successful")
}

// GetProfile handles get user profile
func (h *UserHandler) GetProfile(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	// Get user ID from context (set by JWT middleware)
	userID, err := middleware.GetUserID(c)
	if err != nil {
		log.WithError(err).Warn("Failed to get user ID from context")
		return utils.UnauthorizedErrorResponse(c, "Unauthorized")
	}

	user, err := h.UserService.GetUserByID(ctx, userID)
	if err != nil {
		log.WithError(err).Error("Failed to get user profile")
		return utils.NotFoundErrorResponse(c, "User not found")
	}

	return utils.SuccessResponse(c, user, "User profile retrieved successfully")
}

// GetUserByID handles get user by ID
func (h *UserHandler) GetUserByID(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	// Parse user ID from path parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		log.WithError(err).Warn("Invalid user ID")
		return utils.ValidationErrorResponse(c, "Invalid user ID", []string{err.Error()})
	}

	// Get organization ID from context (set by JWT middleware)
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		log.WithError(err).Warn("Failed to get organization ID from context")
		return utils.UnauthorizedErrorResponse(c, "Organization context required")
	}

	// Get requesting user's role
	role, err := middleware.GetUserRole(c)
	if err != nil {
		log.WithError(err).Warn("Failed to get user role from context")
		return utils.UnauthorizedErrorResponse(c, "Role context required")
	}

	user, err := h.UserService.GetUserByID(ctx, uint(id))
	if err != nil {
		log.WithError(err).Error("Failed to get user")
		return utils.NotFoundErrorResponse(c, "User not found")
	}

	// If not admin and user is not in same organization, deny access
	if role != "admin" && user.OrganizationID != orgID {
		log.WithField("user_id", id).Warn("Unauthorized attempt to access user from different organization")
		return utils.ForbiddenErrorResponse(c, "Access denied to user from different organization")
	}

	return utils.SuccessResponse(c, user, "User retrieved successfully")
}

// UpdateUser handles update user
func (h *UserHandler) UpdateUser(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	// Get user ID from context (set by JWT middleware)
	userID, err := middleware.GetUserID(c)
	if err != nil {
		log.WithError(err).Warn("Failed to get user ID from context")
		return utils.UnauthorizedErrorResponse(c, "Unauthorized")
	}

	var req UpdateUserRequest
	if err := c.Bind(&req); err != nil {
		log.WithError(err).Warn("Invalid request payload")
		return utils.ValidationErrorResponse(c, "Invalid request payload", []string{err.Error()})
	}

	user, err := h.UserService.UpdateUser(ctx, userID, req.Name, req.Email, req.Password)
	if err != nil {
		log.WithError(err).Error("Failed to update user")
		return utils.ErrorResponse(c, http.StatusBadRequest, "Failed to update user", []string{err.Error()})
	}

	return utils.SuccessResponse(c, user, "User updated successfully")
}

// UpdateUserRole handles updating a user's role
func (h *UserHandler) UpdateUserRole(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	// Parse user ID from path parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		log.WithError(err).Warn("Invalid user ID")
		return utils.ValidationErrorResponse(c, "Invalid user ID", []string{err.Error()})
	}

	var req UpdateUserRoleRequest
	if err := c.Bind(&req); err != nil {
		log.WithError(err).Warn("Invalid request payload")
		return utils.ValidationErrorResponse(c, "Invalid request payload", []string{err.Error()})
	}

	// Get organization ID from context (set by JWT middleware)
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		log.WithError(err).Warn("Failed to get organization ID from context")
		return utils.UnauthorizedErrorResponse(c, "Organization context required")
	}

	// Verify user belongs to the same organization
	user, err := h.UserService.GetUserByID(ctx, uint(id))
	if err != nil {
		log.WithError(err).Error("Failed to get user")
		return utils.NotFoundErrorResponse(c, "User not found")
	}

	if user.OrganizationID != orgID {
		log.WithField("user_id", id).Warn("Unauthorized attempt to update role of user from different organization")
		return utils.ForbiddenErrorResponse(c, "Access denied to user from different organization")
	}

	updatedUser, err := h.UserService.UpdateUserRole(ctx, uint(id), req.Role)
	if err != nil {
		log.WithError(err).Error("Failed to update user role")
		return utils.ErrorResponse(c, http.StatusBadRequest, "Failed to update user role", []string{err.Error()})
	}

	return utils.SuccessResponse(c, updatedUser, "User role updated successfully")
}

// DeleteUser handles delete user
func (h *UserHandler) DeleteUser(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	// Get user ID from context (set by JWT middleware)
	userID, err := middleware.GetUserID(c)
	if err != nil {
		log.WithError(err).Warn("Failed to get user ID from context")
		return utils.UnauthorizedErrorResponse(c, "Unauthorized")
	}

	if err := h.UserService.DeleteUser(ctx, userID); err != nil {
		log.WithError(err).Error("Failed to delete user")
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete user", []string{err.Error()})
	}

	return utils.SuccessResponse(c, nil, "User deleted successfully")
}

// DeleteUserByID handles deletion of a user by ID (admin only)
func (h *UserHandler) DeleteUserByID(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	// Parse user ID from path parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		log.WithError(err).Warn("Invalid user ID")
		return utils.ValidationErrorResponse(c, "Invalid user ID", []string{err.Error()})
	}

	// Get organization ID from context (set by JWT middleware)
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		log.WithError(err).Warn("Failed to get organization ID from context")
		return utils.UnauthorizedErrorResponse(c, "Organization context required")
	}

	// Verify user belongs to the same organization
	user, err := h.UserService.GetUserByID(ctx, uint(id))
	if err != nil {
		log.WithError(err).Error("Failed to get user")
		return utils.NotFoundErrorResponse(c, "User not found")
	}

	if user.OrganizationID != orgID {
		log.WithField("user_id", id).Warn("Unauthorized attempt to delete user from different organization")
		return utils.ForbiddenErrorResponse(c, "Access denied to user from different organization")
	}

	if err := h.UserService.DeleteUser(ctx, uint(id)); err != nil {
		log.WithError(err).Error("Failed to delete user")
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete user", []string{err.Error()})
	}

	return utils.SuccessResponse(c, nil, "User deleted successfully")
}

// ListUsers handles list users
func (h *UserHandler) ListUsers(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	// Get organization ID from context (set by JWT middleware)
	orgID, err := middleware.GetOrganizationID(c)
	if err != nil {
		log.WithError(err).Warn("Failed to get organization ID from context")
		return utils.UnauthorizedErrorResponse(c, "Organization context required")
	}

	// Get organization-specific users
	users, total, err := h.UserService.ListOrganizationUsers(ctx, orgID, page, perPage)
	if err != nil {
		log.WithError(err).Error("Failed to list users")
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list users", []string{err.Error()})
	}

	// Calculate total pages
	totalPages := total / int64(perPage)
	if total%int64(perPage) > 0 {
		totalPages++
	}

	pageInfo := &utils.PageInfo{
		Page:      page,
		PerPage:   perPage,
		TotalPage: totalPages,
	}

	response := utils.Response{
		Status:     "success",
		RequestID:  c.Response().Header().Get(echo.HeaderXRequestID),
		Message:    "Users retrieved successfully",
		Data:       users,
		PageInfo:   pageInfo,
		TotalCount: total,
	}

	return c.JSON(http.StatusOK, response)
}

// RegisterRoutes registers the user routes
func (h *UserHandler) RegisterRoutes(e *echo.Echo, jwtMiddleware, adminMiddleware echo.MiddlewareFunc) {
	// Public routes
	e.POST("/api/register", h.Register)
	e.POST("/api/login", h.Login)

	// Protected routes
	userGroup := e.Group("/api/users")
	userGroup.Use(jwtMiddleware)

	userGroup.GET("", h.ListUsers)
	userGroup.GET("/profile", h.GetProfile)
	userGroup.GET("/:id", h.GetUserByID)
	userGroup.PUT("", h.UpdateUser)
	userGroup.DELETE("", h.DeleteUser)

	// Admin-only routes
	adminGroup := userGroup.Group("")
	adminGroup.Use(adminMiddleware)
	adminGroup.PUT("/:id/role", h.UpdateUserRole)
	adminGroup.DELETE("/:id", h.DeleteUserByID)
}
