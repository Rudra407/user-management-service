package handlers

import (
	"net/http"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/user/user-management-service/internal/services"
	"github.com/user/user-management-service/utils"
)

// OrganizationHandler handles HTTP requests for organization operations
type OrganizationHandler struct {
	OrganizationService *services.OrganizationService
	Logger              *utils.Logger
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(organizationService *services.OrganizationService, logger *utils.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		OrganizationService: organizationService,
		Logger:              logger,
	}
}

// CreateOrganizationRequest represents an organization creation request
type CreateOrganizationRequest struct {
	Name          string `json:"name" validate:"required"`
	DisplayName   string `json:"display_name" validate:"required"`
	Description   string `json:"description"`
	Website       string `json:"website"`
	AdminName     string `json:"admin_name" validate:"required"`
	AdminEmail    string `json:"admin_email" validate:"required,email"`
	AdminPassword string `json:"admin_password" validate:"required,min=6"`
}

// UpdateOrganizationRequest represents an organization update request
type UpdateOrganizationRequest struct {
	DisplayName string `json:"display_name"`
	Description string `json:"description"`
	Website     string `json:"website"`
}

// CreateOrganization handles organization creation
func (h *OrganizationHandler) CreateOrganization(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	var req CreateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		log.WithError(err).Warn("Invalid request payload")
		return utils.ValidationErrorResponse(c, "Invalid request payload", []string{err.Error()})
	}

	org, admin, err := h.OrganizationService.CreateOrganization(
		ctx,
		req.Name,
		req.DisplayName,
		req.Description,
		req.Website,
		req.AdminName,
		req.AdminEmail,
		req.AdminPassword,
	)
	if err != nil {
		log.WithError(err).Error("Failed to create organization")
		return utils.ErrorResponse(c, http.StatusBadRequest, "Failed to create organization", []string{err.Error()})
	}

	// Hide admin password in response
	admin.Password = ""

	response := map[string]interface{}{
		"organization": org,
		"admin":        admin,
	}

	return utils.SuccessResponse(c, response, "Organization created successfully")
}

// GetOrganization handles get organization by ID
func (h *OrganizationHandler) GetOrganization(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	// Parse organization ID from path parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		log.WithError(err).Warn("Invalid organization ID")
		return utils.ValidationErrorResponse(c, "Invalid organization ID", []string{err.Error()})
	}

	org, err := h.OrganizationService.GetOrganizationByID(ctx, uint(id))
	if err != nil {
		log.WithError(err).Error("Failed to get organization")
		return utils.NotFoundErrorResponse(c, "Organization not found")
	}

	return utils.SuccessResponse(c, org, "Organization retrieved successfully")
}

// UpdateOrganization handles update organization
func (h *OrganizationHandler) UpdateOrganization(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	// Parse organization ID from path parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		log.WithError(err).Warn("Invalid organization ID")
		return utils.ValidationErrorResponse(c, "Invalid organization ID", []string{err.Error()})
	}

	var req UpdateOrganizationRequest
	if err := c.Bind(&req); err != nil {
		log.WithError(err).Warn("Invalid request payload")
		return utils.ValidationErrorResponse(c, "Invalid request payload", []string{err.Error()})
	}

	org, err := h.OrganizationService.UpdateOrganization(
		ctx,
		uint(id),
		req.DisplayName,
		req.Description,
		req.Website,
	)
	if err != nil {
		log.WithError(err).Error("Failed to update organization")
		return utils.ErrorResponse(c, http.StatusBadRequest, "Failed to update organization", []string{err.Error()})
	}

	return utils.SuccessResponse(c, org, "Organization updated successfully")
}

// DeleteOrganization handles delete organization
func (h *OrganizationHandler) DeleteOrganization(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	// Parse organization ID from path parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		log.WithError(err).Warn("Invalid organization ID")
		return utils.ValidationErrorResponse(c, "Invalid organization ID", []string{err.Error()})
	}

	if err := h.OrganizationService.DeleteOrganization(ctx, uint(id)); err != nil {
		log.WithError(err).Error("Failed to delete organization")
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to delete organization", []string{err.Error()})
	}

	return utils.SuccessResponse(c, nil, "Organization deleted successfully")
}

// ListOrganizations handles list organizations
func (h *OrganizationHandler) ListOrganizations(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	orgs, total, err := h.OrganizationService.ListOrganizations(ctx, page, perPage)
	if err != nil {
		log.WithError(err).Error("Failed to list organizations")
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list organizations", []string{err.Error()})
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
		Message:    "Organizations retrieved successfully",
		Data:       orgs,
		PageInfo:   pageInfo,
		TotalCount: total,
	}

	return c.JSON(http.StatusOK, response)
}

// ListOrganizationUsers handles list users for an organization
func (h *OrganizationHandler) ListOrganizationUsers(c echo.Context) error {
	ctx := utils.NewRequestContext()
	log := h.Logger.WithContext(ctx)

	// Parse organization ID from path parameter
	idParam := c.Param("id")
	id, err := strconv.ParseUint(idParam, 10, 32)
	if err != nil {
		log.WithError(err).Warn("Invalid organization ID")
		return utils.ValidationErrorResponse(c, "Invalid organization ID", []string{err.Error()})
	}

	// Parse pagination parameters
	page, _ := strconv.Atoi(c.QueryParam("page"))
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))

	users, total, err := h.OrganizationService.ListOrganizationUsers(ctx, uint(id), page, perPage)
	if err != nil {
		log.WithError(err).Error("Failed to list organization users")
		return utils.ErrorResponse(c, http.StatusInternalServerError, "Failed to list organization users", []string{err.Error()})
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
		Message:    "Organization users retrieved successfully",
		Data:       users,
		PageInfo:   pageInfo,
		TotalCount: total,
	}

	return c.JSON(http.StatusOK, response)
}

// RegisterRoutes registers the organization routes
func (h *OrganizationHandler) RegisterRoutes(e *echo.Echo, jwtMiddleware, adminMiddleware echo.MiddlewareFunc) {
	// Public routes for organization creation
	e.POST("/api/organizations", h.CreateOrganization)

	// Protected routes for organizations
	orgGroup := e.Group("/api/organizations")
	orgGroup.Use(jwtMiddleware)

	// Routes for all authenticated users
	orgGroup.GET("", h.ListOrganizations)
	orgGroup.GET("/:id", h.GetOrganization)

	// Admin-only routes
	adminGroup := orgGroup.Group("")
	adminGroup.Use(adminMiddleware)
	adminGroup.PUT("/:id", h.UpdateOrganization)
	adminGroup.DELETE("/:id", h.DeleteOrganization)

	// Get users for an organization (admin only)
	orgGroup.GET("/:id/users", h.ListOrganizationUsers, adminMiddleware)
}
