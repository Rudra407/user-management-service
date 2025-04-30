package handlers

import (
	"net/http"
	"strconv"

	"github.com/golang-jwt/jwt/v5"
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
	Name     string `json:"name" validate:"required"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// LoginRequest represents a user login request
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

// UpdateUserRequest represents a user update request
type UpdateUserRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email" validate:"omitempty,email"`
	Password string `json:"password" validate:"omitempty,min=6"`
}

// Register handles user registration
func (h *UserHandler) Register(c echo.Context) error {
	ctx := c.Request().Context()
	log := h.Logger.WithContext(ctx)

	// Parse request body
	var req struct {
		Name           string `json:"name"`
		Email          string `json:"email"`
		Password       string `json:"password"`
		OrganizationID uint   `json:"organization_id"` // Required field now
	}

	if err := c.Bind(&req); err != nil {
		log.WithError(err).Warn("Invalid request payload")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Validate request
	if req.Name == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Name is required"})
	}

	if req.Email == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Email is required"})
	}

	if req.Password == "" {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password is required"})
	}

	if len(req.Password) < 6 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Password must be at least 6 characters"})
	}

	if req.OrganizationID == 0 {
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Organization ID is required"})
	}

	// Register user
	orgID := req.OrganizationID
	user, err := h.UserService.RegisterUser(ctx, req.Name, req.Email, req.Password, &orgID)
	if err != nil {
		log.WithError(err).Error("Failed to register user")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": err.Error()})
	}

	log.WithField("user_id", user.ID).Info("User registered successfully")
	return c.JSON(http.StatusCreated, user)
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

	token, err := h.UserService.Login(ctx, req.Email, req.Password)
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

	user, err := h.UserService.GetUserByID(ctx, uint(id))
	if err != nil {
		log.WithError(err).Error("Failed to get user")
		return utils.NotFoundErrorResponse(c, "User not found")
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

// ListUsers handles list users
func (h *UserHandler) ListUsers(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	if page < 1 {
		page = 1
	}

	if perPage < 1 {
		perPage = 10
	}

	users, total, err := h.UserService.ListUsers(ctx, page, perPage)
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

// GetUserOrganization retrieves the organization for the authenticated user
func (h *UserHandler) GetUserOrganization(c echo.Context) error {
	// Get user from JWT token
	user := c.Get("user").(*jwt.Token)
	claims := user.Claims.(*utils.JWTClaims)
	userID := claims.UserID
	ctx := c.Request().Context()

	// Find the user's organization
	organization, err := h.UserService.GetUserOrganization(ctx, userID)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to retrieve user organization"})
	}

	if organization == nil {
		return c.JSON(http.StatusNotFound, map[string]string{"error": "User is not part of any organization"})
	}

	return c.JSON(http.StatusOK, organization)
}

// RegisterRoutes registers user routes
func (h *UserHandler) RegisterRoutes(e *echo.Echo, jwtMiddleware echo.MiddlewareFunc) {
	// Public routes - no authentication needed
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
	userGroup.GET("/organization", h.GetUserOrganization)
}
