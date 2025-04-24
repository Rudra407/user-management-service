package middleware

import (
	"net/http"
	"strings"

	"github.com/labstack/echo/v4"
	"github.com/user/user-management-service/config"
	"github.com/user/user-management-service/utils"
)

// JWTMiddleware creates a middleware that validates JWT tokens
func JWTMiddleware(config *config.Config, logger *utils.Logger) echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			// Get the Authorization header
			authHeader := c.Request().Header.Get("Authorization")
			if authHeader == "" {
				logger.WithField("path", c.Request().URL.Path).Warn("Missing Authorization header")
				return utils.UnauthorizedErrorResponse(c, "Missing Authorization header")
			}

			// Extract the token
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				logger.WithField("path", c.Request().URL.Path).Warn("Invalid Authorization format")
				return utils.UnauthorizedErrorResponse(c, "Invalid Authorization format")
			}

			tokenString := tokenParts[1]

			// Validate the token
			claims, err := utils.ValidateToken(tokenString, config.JWT.Secret)
			if err != nil {
				logger.WithField("error", err.Error()).Warn("Invalid JWT token")
				return utils.UnauthorizedErrorResponse(c, "Invalid token")
			}

			// Set the user ID in context
			c.Set("user_id", claims.UserID)
			return next(c)
		}
	}
}

// GetUserID gets the user ID from context
func GetUserID(c echo.Context) (uint, error) {
	userID, ok := c.Get("user_id").(uint)
	if !ok {
		return 0, echo.NewHTTPError(http.StatusUnauthorized, "Unauthorized")
	}
	return userID, nil
}
