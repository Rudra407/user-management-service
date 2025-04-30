package handlers

import (
	"net/http"

	"github.com/labstack/echo/v4"
	"github.com/user/user-management-service/internal/services"
	"github.com/user/user-management-service/utils"
)

// OrganizationHandler handles HTTP requests for organizations
type OrganizationHandler struct {
	OrgService services.OrganizationService
	Logger     *utils.Logger
}

// NewOrganizationHandler creates a new organization handler
func NewOrganizationHandler(orgService services.OrganizationService, logger *utils.Logger) *OrganizationHandler {
	return &OrganizationHandler{
		OrgService: orgService,
		Logger:     logger,
	}
}

// RegisterRoutes registers organization routes
func (h *OrganizationHandler) RegisterRoutes(e *echo.Echo) {
	// Public routes for organization creation - no JWT middleware needed
	e.POST("/api/organizations", h.CreateOrganization)
}

// CreateOrganization creates a new organization
func (h *OrganizationHandler) CreateOrganization(c echo.Context) error {
	ctx := c.Request().Context()
	log := h.Logger.WithContext(ctx)

	// Parse request body
	var req struct {
		Name        string `json:"name"`
		Description string `json:"description"`
	}

	if err := c.Bind(&req); err != nil {
		log.WithError(err).Warn("Invalid request payload")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Invalid request payload"})
	}

	// Validate request
	if req.Name == "" {
		log.Warn("Name is required")
		return c.JSON(http.StatusBadRequest, map[string]string{"error": "Organization name is required"})
	}

	// Create organization
	org, err := h.OrgService.CreateOrganization(ctx, req.Name, req.Description)
	if err != nil {
		log.WithError(err).Error("Failed to create organization")
		return c.JSON(http.StatusInternalServerError, map[string]string{"error": "Failed to create organization"})
	}

	log.WithField("org_id", org.ID).Info("Organization created successfully")
	return c.JSON(http.StatusCreated, org)
}
